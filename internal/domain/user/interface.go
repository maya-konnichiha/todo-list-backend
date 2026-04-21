package user

// このファイルは「User を永続化する操作の**契約**」を定義する場所。
// 実装（pgx で INSERT する具体的な SQL）はここには書かない。
// 実装は infrastructure/postgres 側に置き、ここで宣言したインターフェースを満たす。
//
// なぜ interface を domain 層に置くのか:
//   - usecase が「DB の都合」を知らずに済むため。usecase は UserRepository
//     というインターフェース（= 欲しい機能の一覧）だけに依存する。
//   - これにより
//       * テスト時にモック実装に差し替えられる
//       * DB を Postgres から MySQL に替えても usecase は変更不要
//   - 「使う側がインターフェースを定義する」のが Go の流儀（Accept Interfaces,
//     Return Structs）。ここでは domain/usecase が "使う側"。

import (
	"context" // タイムアウト/キャンセル/リクエスト単位の値を伝える標準の伝達手段
	"errors"
)

// ErrUserNotFound は指定された User が見つからない（または論理削除済み）ときに
// 返すエラー。
//
// handler 層ではこれを拾って HTTP 404 に変換する想定:
//
//	if errors.Is(err, user.ErrUserNotFound) { c.JSON(404, ...) }
//
// sentinel エラーを「repository 呼び出しの結果としての不在」を表すために置く。
// NewUser のバリデーションエラー（ErrUserNameEmpty 等）とは別種の概念。
var ErrUserNotFound = errors.New("user not found")

// ErrUserEmailAlreadyExists は Create 時に user_email が既に使われていた場合に返す。
//
// DB の UNIQUE 制約違反を repository 実装が検知し、これに変換して返す想定。
// handler では HTTP 409 Conflict に対応付けると自然。
//
// こうして「DB 固有のエラー（pgconn.PgError code=23505）」を infra 層で吸収し、
// ドメイン語彙のエラーに翻訳するのが層分離のポイント。
// usecase / handler は 23505 の存在を知らずに済む。
var ErrUserEmailAlreadyExists = errors.New("user_email already exists")

// CreateParams は UserRepository.Create のパラメータ構造体。
//
// なぜ User 構造体ではなく Params を受け取るのか:
//   - 新規作成時点では UserID / CreatedAt / UpdatedAt を呼び出し側が持っていない
//     （DB が IDENTITY と DEFAULT CURRENT_TIMESTAMP で発行する）
//   - 「ゼロ値の User を渡す」より「必要な値だけ入れた Params を渡す」方が意図が
//     明確になる
//   - 引数の増減に強い（後方互換）
type CreateParams struct {
	UserName  string
	UserEmail string
}

// UserRepository は User の永続化を抽象化するインターフェース。
//
// Go の interface は「このメソッド群を満たすものは何でも受け付ける」という契約。
// ここには**シグネチャだけ**書き、実装は書かない。
//
// メソッドの第1引数は常に `context.Context`。Go では慣習として必ず最初に置く。
//   - リクエストのキャンセル（クライアントが切断した等）を DB 呼び出しに伝搬できる
//   - タイムアウトを設定できる
//   - トランザクションや user_id などのリクエスト単位の値を載せられる
//     （material-hub ではトランザクションを context 経由で扱っている）
//
// 論理削除の扱い（このプロジェクトの方針）:
//   - FindByID は deleted_at IS NULL の行だけを見つける（削除済みは「存在しない」扱い）
//   - Create は生の INSERT なので deleted_at は NULL で作られる
//   - 「削除済みも含めて探す」必要が出てきたら別メソッド（例: FindByIDIncludingDeleted）
//     を足す方針。デフォルト挙動に "deleted も混ぜる" を選ぶとバグの温床になる。
type UserRepository interface {
	// Create は新しい User を作成する。
	//
	//   - UserID / CreatedAt / UpdatedAt は DB が発行し、返り値の *User に詰めて返す。
	//   - UNIQUE 制約違反（email 重複）のときは ErrUserEmailAlreadyExists を返す。
	//   - バリデーション（名前が空 等）は**呼び出し側（usecase）で済ませている前提**。
	//     repository は「壊れた入力」を受け取らない設計にする。
	Create(ctx context.Context, params CreateParams) (*User, error)

	// FindByID は指定 ID の User を取得する。
	//
	//   - 見つからない場合: ErrUserNotFound を返す（nil, err）
	//   - 論理削除済み（deleted_at IS NOT NULL）も「見つからない」扱い
	//   - DB エラー等の想定外: ラップしたエラーを返す
	//
	// 戻り値を (*User, error) にしている理由は Go の慣用パターン。
	// エラーが nil ならポインタが有効、エラーが non-nil ならポインタは nil、と読むのが暗黙のルール。
	FindByID(ctx context.Context, userID int64) (*User, error)
}
