// Package user は User エンティティのドメインモデルとリポジトリ契約を提供する。
//
// Go では「ディレクトリ名 = パッケージ名」が慣習。
// ディレクトリ名は小文字のみ（例: "user"）が推奨。今回は "user" で揃える。
package user

import (
	"errors"   // sentinel エラー（固定エラー値）を作るための標準ライブラリ
	"net/mail" // メールアドレスのパース。外部ライブラリ不要で使える
	"strings"  // 文字列の前後空白除去（TrimSpace）などに使用
	"time"     // 時刻型。DB の TIMESTAMPTZ と対応する
)

// --- ドメインバリデーションエラー ------------------------------------------
//
// 「どの入力が悪かったか」を呼び出し側が判別できるよう、固定のエラー値を
// 用意する。これを "sentinel error" と呼ぶ。
// 呼び出し側（usecase や handler）は `errors.Is(err, user.ErrUserNameEmpty)` で
// 分岐判定できる。
//
// 重要な層の境界:
//   - ドメインは「業務として不正」を errors で通知する。
//   - 「HTTP 400 にすべきか 422 にすべきか」は handler 層が決める。
//   - ドメインはステータスコードを一切知らない。
var (
	ErrUserNameEmpty    = errors.New("user_name is required")
	ErrUserNameTooLong  = errors.New("user_name must be 50 characters or less")
	ErrUserEmailEmpty   = errors.New("user_email is required")
	ErrUserEmailInvalid = errors.New("user_email is invalid")
)

// userNameMaxLen はユーザー名の最大文字数。
// DB スキーマ（VARCHAR(50)）と揃えている。
// マジックナンバーを定数化する理由:
//   - スキーマ変更時にここ1か所を直すだけで済む
//   - エラーメッセージと不整合になりにくい
const userNameMaxLen = 50

// User はユーザーを表すドメインエンティティ。
//
// 設計判断:
//  1. `UserID` は int64。
//     DB が BIGINT GENERATED ALWAYS AS IDENTITY で発行するため Go 側は int64 で受ける。
//     "IDENTITY" = DB が自動で連番を割り当てる仕組み（PostgreSQL の標準機能）。
//
//  2. JSON タグ（`json:"user_name"` 等）を**書かない**。
//     JSON の形は「HTTP の都合」で handler 層の DTO で扱う。
//     ドメインを JSON 表現に縛ると、API 変更のたびに内側まで揺れてしまう。
//
//  3. `DeletedAt` は `*time.Time`（time.Time のポインタ）。
//     time.Time のゼロ値は "0001-01-01 00:00:00 UTC" という実在する値なので、
//     「まだ削除されていない」をゼロ値で表すと「実在のゼロ日付」と区別できない。
//     Go で NULL 可能なカラムを表すときはポインタ型にするのが定番。
//     （nil = NULL = 未削除）
//
//  4. フィールド名は UpperCamelCase。
//     先頭大文字は Go の仕様上「パッケージ外から参照可能（公開）」の意味。
//     小文字始まりだとパッケージ内でしか使えない。
type User struct {
	UserID    int64      // DB 発行の ID。新規作成直後（永続化前）はゼロ値 0
	UserName  string     // 表示名（VARCHAR(50)）
	UserEmail string     // メールアドレス（VARCHAR(255), UNIQUE は DB 側で保証）
	CreatedAt time.Time  // 作成日時（DB の DEFAULT now() を想定）
	UpdatedAt time.Time  // 更新日時
	DeletedAt *time.Time // 論理削除日時。nil なら未削除
}

// IsDeleted は論理削除済みかを返す。
//
// 受信者を `(u *User)` にしている理由:
//   - 構造体全体をコピーせずに済む（効率）
//   - 他メソッドと受信者型を揃えると読みやすい
//
// ここでは値を変更しない参照系メソッドだが、プロジェクト方針として *User で統一する。
// （値レシーバと混在すると「このメソッドはコピーを作るのか？」が読み手に伝わりにくい）
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// ============================================================================
// コンストラクタは用途別に 2 種類に分ける。
//
//   1. NewUser        … これから DB に保存する「まだ ID を持たない User」を作る
//   2. Reconstruct    … DB から読み出したレコードを User に復元する
//
// なぜ分けるか:
//   - 新規作成時は UserID/CreatedAt/UpdatedAt を呼び出し側が持っていない（DB が発行）
//   - 読み出し時は全部揃っている
//   これを 1 つの関数で扱うと「UserID=0 は未保存の意味」等の暗黙ルールが増えて
//   事故の元になる。関数名で「どちらの場面か」を明示する方が安全。
// ============================================================================

// NewUserParams は NewUser のパラメータ構造体。
//
// なぜ Params 構造体を挟むか:
//   - 引数名が呼び出し側に見える → NewUser(NewUserParams{UserName: "alice", ...})
//     位置引数だと NewUser("alice", "a@b.c") となり、順序ミスに気づきにくい
//   - 将来フィールドが増えても既存呼び出しが壊れない（後方互換）
type NewUserParams struct {
	UserName  string
	UserEmail string
}

// NewUser は「まだ永続化されていない」User を作るコンストラクタ。
//
//   - UserID, CreatedAt, UpdatedAt, DeletedAt はすべてゼロ値のまま。
//     → DB INSERT 時に DB 側で自動発行され、結果として Reconstruct で埋め直される。
//   - 返り値はポインタ `*User`。理由:
//       * エラー時に nil を返して「値が無い」を表現できる
//       * 呼び出し側でフィールドを変更する場面がある
//       * 構造体のコピーコストを避けられる
//   - 戻り値の `(*User, error)` は Go の慣用パターン（成功値, エラー）。
//     Go には例外が無く、errors は戻り値で伝える。
func NewUser(params NewUserParams) (*User, error) {
	name, email, err := normalizeAndValidate(params.UserName, params.UserEmail)
	if err != nil {
		return nil, err
	}
	// 構造体リテラル + `&` でポインタを生成する。
	// C/C++ のポインタと違い、関数を抜けても GC が面倒を見てくれるので安全。
	return &User{
		UserName:  name,
		UserEmail: email,
		// UserID / CreatedAt / UpdatedAt / DeletedAt は**あえて書かない**。
		// Go では書かなかったフィールドは自動的にゼロ値（int64 なら 0、
		// time.Time なら "0001-01-01..."、ポインタなら nil）になる。
	}, nil
}

// ReconstructParams は Reconstruct のパラメータ構造体。
// DB から SELECT した 1 行を詰め替えるときに使う。
type ReconstructParams struct {
	UserID    int64
	UserName  string
	UserEmail string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // DB の NULL は nil として受け取る
}

// Reconstruct は DB レコードから User を復元するコンストラクタ。
//
// 用途: infrastructure 層（pgx での SELECT 後）で使う。
// ここでもバリデーションを走らせているのは「DB に壊れたデータが入っていた場合の
// 早期検知」のため。普段は通る想定だが、防御の意味で残している。
//
// 注意: このコンストラクタを「新規作成の代わりに」呼ばないこと。
// UserID を呼び出し側が決め打ちで渡してしまい、DB の IDENTITY と衝突する。
func Reconstruct(params ReconstructParams) (*User, error) {
	name, email, err := normalizeAndValidate(params.UserName, params.UserEmail)
	if err != nil {
		return nil, err
	}
	return &User{
		UserID:    params.UserID,
		UserName:  name,
		UserEmail: email,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt,
		DeletedAt: params.DeletedAt,
	}, nil
}

// normalizeAndValidate は name と email の正規化＋検証を行う。
//
// 関数名が小文字始まり（`normalize...`）なので**パッケージ外からは呼べない**。
// Go では「export したいものだけ大文字で始める」ことで API 表面を絞る。
//
// "正規化" とは: 入力の揺れを吸収する処理。ここでは前後の空白を除去している。
// （" alice " と "alice" を同一視する等）
func normalizeAndValidate(rawName, rawEmail string) (string, string, error) {
	// TrimSpace は前後の空白・改行・タブなどを除去する。
	// これを省くと、空白だけの文字列 "   " が空チェックをすり抜けてしまう。
	name := strings.TrimSpace(rawName)
	email := strings.TrimSpace(rawEmail)

	if name == "" {
		return "", "", ErrUserNameEmpty
	}
	// 文字数は「バイト数」ではなく「文字数（runes）」で数える。
	// 日本語 1 文字は UTF-8 で 3 バイトあるので len(name) だと長すぎ判定になる。
	// rune = Unicode コードポイント 1 つ分。len([]rune(...)) で文字数が取れる。
	if len([]rune(name)) > userNameMaxLen {
		return "", "", ErrUserNameTooLong
	}
	if email == "" {
		return "", "", ErrUserEmailEmpty
	}
	// net/mail.ParseAddress は RFC 5322 ベースのパーサ。
	// 標準ライブラリで済むので外部依存を足さずに済む = domain 層の掟を守れる。
	// 厳密な検証ではないが、学習用途には十分。
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", ErrUserEmailInvalid
	}
	return name, email, nil
}
