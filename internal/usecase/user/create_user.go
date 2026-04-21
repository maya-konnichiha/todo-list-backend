// Package user は User エンティティに関するユースケース（アプリの手順）を提供する。
//
// 層の役割の違い:
//   - domain:  「User とは何か」「User として正しい状態は何か」（不変のビジネスルール）
//   - usecase: 「User を作る手順」「User を探す手順」（アプリ固有の流れ）
//   - handler: HTTP とのやり取り（JSON パース、ステータスコード変換）
//
// ディレクトリは internal/usecase/user/ だが、Go のパッケージ名はディレクトリ名と
// 揃えるのが慣習なので `package user` とする。
// → domain/user と名前が衝突するため、import 時に alias を付ける。
package user

import (
	"context"

	// domain の user パッケージを alias 付きで import する。
	// 自パッケージ名（user）と衝突するので alias 必須。
	// 慣習: Go は単語のくっつけ（小文字）を好むので `userdomain` とする。
	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// CreateUser は User を新規作成するユースケース。
//
// なぜ「関数」ではなく「構造体＋メソッド」で作るのか（DI の話）:
//   - 依存する Repository をコンストラクタで受け取り、内部で保持する
//   - 関数版にすると CreateUser(ctx, repo, params) と呼び出し側が毎回 repo を
//     渡すことになり冗長
//   - テスト時に「モックに差し替えた uc を一度作って、何度も .Execute(...) する」
//     形が自然
//
// なぜ Repository を内部で `new` しないのか（= 自分で作らず外から受け取る理由）:
//   - 自前で作ると DB 接続プールの作成責務がここに漏れてくる
//   - テストで実 DB を使う羽目になる（モックに差し替え不可）
//   - ライフサイクル（接続 close タイミング等）が散らばる
//   → 依存は**外から渡す**のが鉄則。これが「依存性注入（Dependency Injection）」。
//
// 組み立ての絵:
//
//	cmd/server/main.go で
//	  pool        := pgxpool.New(...)              // ← 接続プールを作る
//	  userRepo    := postgres.NewUserRepository(pool)
//	  createUser  := usecaseUser.NewCreateUser(userRepo)   // ← ここで DI
//	  ... createUser を handler に渡す ...
type CreateUser struct {
	// 型は **具象型**（postgres.UserRepository）ではなく **interface**（user.UserRepository）で持つ。
	//   - usecase が「Postgres の事情」を知らずに済む
	//   - テストで偽物 repo を渡せる（Go の interface は implicit = 暗黙実装なので、
	//     モック側で同じメソッドを揃えれば自動的に interface を満たす）
	repo userdomain.UserRepository
}

// NewCreateUser はコンストラクタ。
//
// 返り値を *CreateUser（ポインタ）にしている理由:
//   - 状態（repo）を持つので、呼び出すたびにコピーしたくない
//   - フィールドが増えてもコピーコストが一定
//   - nil を返せば「作れなかった」を表現できる（今回は常に成功するが慣例）
func NewCreateUser(repo userdomain.UserRepository) *CreateUser {
	return &CreateUser{repo: repo}
}

// CreateUserParams は Execute の入力構造体。
//
// handler 層のリクエスト DTO とは別物である点に注意:
//   - リクエスト DTO: JSON の形（例: {"user_name": "...", "user_email": "..."}）
//   - CreateUserParams: usecase にとって必要な素の入力
//   - handler で DTO → Params に詰め替える
//
// 分けるメリット:
//   - API の JSON 形が変わっても usecase 側は揺れない
//   - HTTP 以外（CLI、ワーカー、cron 等）から呼びたくなった時に DTO を経由しなくて済む
type CreateUserParams struct {
	UserName  string
	UserEmail string
}

// Execute はユーザー作成の手順を実行する。
//
// 手順は 3 ステップ:
//  1. domain の NewUser で入力をバリデーション（& 正規化）
//  2. Repository の Create で DB に INSERT
//  3. DB が発行した ID / 時刻入りの User を返す
//
// context.Context について:
//   - 第 1 引数に ctx を取り、repo の呼び出しへそのまま渡す（伝搬）
//   - handler から始まる「1 リクエストの寿命」が ctx に紐づいている
//     * クライアントが切断すれば ctx.Done() が閉じ、DB クエリも中断される
//     * タイムアウトが設定されていれば、その時間で切れる
//   - usecase 層で ctx を加工（WithValue 等）する必要はまず無い。
//     上から下へ**素通し**するのが基本動作。
func (uc *CreateUser) Execute(ctx context.Context, params CreateUserParams) (*userdomain.User, error) {
	// -----------------------------------------------------------------
	// 1. バリデーション（ドメインルールの適用）
	//
	// ここで NewUser を呼ぶのは「業務として正しい入力か」を確認するため。
	//
	// handler 側にも簡易バリデーション（required, format 等）を置く予定だが、
	// それとは目的が違う:
	//   - handler: 「HTTP リクエストとして妥当か」（早期 400 返却のための最適化）
	//   - domain : 「User として成立するか」（**最後の砦**。HTTP 以外経由でも守られる）
	//
	// NewUser の戻り値を捨てずに使う理由:
	//   - TrimSpace 済みの UserName / UserEmail が取れる
	//   - その「正規化後の値」を repo に渡すことで、DB に前後空白入りの
	//     文字列が保存されるのを防げる
	// -----------------------------------------------------------------
	newUser, err := userdomain.NewUser(userdomain.NewUserParams{
		UserName:  params.UserName,
		UserEmail: params.UserEmail,
	})
	if err != nil {
		// sentinel エラー（ErrUserNameEmpty 等）は**そのまま返す**。
		//
		// エラーハンドリングの方針:
		//   - 情報を足さないときはラップしない（素通し）
		//   - 情報を足すとき（原因コンテキストを追加したい時等）だけ fmt.Errorf("...: %w", err) で包む
		//   - ラップするときは必ず %w（%v はチェーンを切ってしまう）
		//
		// ここで包まずに返すと、handler 側が
		//   if errors.Is(err, userdomain.ErrUserNameEmpty) { c.JSON(400, ...) }
		// のように判別できる。%w で包んでも errors.Is は透過するので実害はないが、
		// 包む動機がない（情報を追加していない）ので素通しが正解。
		return nil, err
	}

	// -----------------------------------------------------------------
	// 2. 永続化
	//
	// repo には **NewUser が正規化した値** を渡す（params の生値ではなく）。
	// こうしないと TrimSpace の意味が無くなる。
	// -----------------------------------------------------------------
	created, err := uc.repo.Create(ctx, userdomain.CreateParams{
		UserName:  newUser.UserName,
		UserEmail: newUser.UserEmail,
	})
	if err != nil {
		// repo からのエラーも原則**素通し**する。
		//   - ErrUserEmailAlreadyExists は handler が errors.Is で拾って 409 に変換
		//   - DB 接続エラー等の想定外エラーは postgres 層で既に %w ラップ済み
		//
		// ここで "create user: %w" を付けてラップすることもできるが、
		// 追加情報が無いので素通し。原則は「情報を足すならラップ、足さないなら透過」。
		return nil, err
	}

	// -----------------------------------------------------------------
	// 3. 返却
	//
	// created には DB が発行した user_id / created_at / updated_at が詰まっている。
	// handler 側でレスポンス DTO に変換する（例: UserID を int64 → string にする等）。
	// -----------------------------------------------------------------
	return created, nil
}
