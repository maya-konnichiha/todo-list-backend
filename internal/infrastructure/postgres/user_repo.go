// Package postgres は domain で宣言された *Repository インターフェースの
// PostgreSQL 実装を提供する。
//
// ここは「infrastructure 層」で、domain 層と違って外部ライブラリ（pgx）への
// 依存が許される。ただし usecase/handler は postgres パッケージを直接 import
// してはいけない。起点は cmd/server/main.go の DI 配線のみ。
//
// 依存方向のおさらい:
//
//	handler → usecase → domain/user (UserRepository interface)
//	                      ▲
//	                      │  implements（postgres が契約を満たしに行く）
//	                      │
//	             infrastructure/postgres.UserRepository
package postgres

import (
	"context"
	"errors" // errors.Is / errors.As で sentinel エラーと型アサーションを行う
	"fmt"    // fmt.Errorf("...: %w", err) でエラーをラップする
	"time"   // Scan 先として time.Time を宣言するために必要

	"github.com/jackc/pgx/v5"         // ErrNoRows（行が無いときのセンチネル）
	"github.com/jackc/pgx/v5/pgconn"  // *PgError（PostgreSQL 固有エラー）
	"github.com/jackc/pgx/v5/pgxpool" // コネクションプール

	"github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// UserRepository は user.UserRepository を PostgreSQL で実装した構造体。
//
// パッケージが `postgres` なので、外から呼ぶ時の型名は `postgres.UserRepository`。
// domain 側の `user.UserRepository`（interface）と名前が衝突しないのはこのため。
//
// フィールドの `pool` は pgxpool.Pool のポインタ。
//   - pgx の「コネクションプール」。同時に複数のゴルーチン（並行実行）から
//     安全に使える。
//   - 生のコネクションを1つだけ持つと、同時リクエストが直列になってしまう。
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository はコンストラクタ。
//
// 外から pool を受け取るのは DI（依存性注入）。
//   - テストで偽物のプールに差し替えやすくなる
//   - 本体が自分で pool を作らず「外から使うものを渡される」形にすると、
//     単体テストしやすい / ライフサイクル管理（Close 等）が一箇所に集約できる
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// コンパイル時に「*UserRepository が user.UserRepository を満たしているか」を検証する。
//
//	var _ user.UserRepository = (*UserRepository)(nil)
//
// 読み方:
//   - 変数名を `_`（ブランク識別子）にして値を捨てる
//   - 右辺の `(*UserRepository)(nil)` は「nil を *UserRepository 型として扱う」
//     という型変換。値そのものは捨てられるので nil でよい
//   - 代入が成立するには「右辺型が左辺 interface を実装している」必要がある
//   - メソッドが足りないと**コンパイルエラー**で早期検知できる
//
// これが無くても実際に使う箇所でコンパイルエラーは出るが、
// 宣言直下に書いておくと「このファイルは UserRepository の実装です」と
// 意図を明示できる。
var _ user.UserRepository = (*UserRepository)(nil)

// --- SQL 文 ---------------------------------------------------------------
//
// SQL は `const` 定数として切り出しておく。
//   - 長い文字列がメソッドに直接埋まるより読みやすい
//   - 別のメソッドから参照したい時に再利用できる
//   - diff が読みやすい
//
// プレースホルダについて:
//   - `$1` `$2` は PostgreSQL のプレースホルダ記法。MySQL の `?` と違うので注意。
//   - プレースホルダを使うと、値は pgx/Postgres 側で「データ」として扱われ、
//     SQL の文法として解釈されない。これが **SQL インジェクション対策**。
//       例: ユーザー名に "'; DROP TABLE users; --" を入れても、ただの文字列として
//       保存されるだけで、テーブルは消えない。
//   - 文字列結合で "INSERT ... VALUES ('" + name + "')" のように SQL を組み立てるのは**絶対にやらない**。

const createUserSQL = `
INSERT INTO users (user_name, user_email)
VALUES ($1, $2)
RETURNING user_id, user_name, user_email, created_at, updated_at, deleted_at
`

// RETURNING 句について:
//   - INSERT した行の値を**同じクエリで取得できる** PostgreSQL 拡張。
//   - これを使わないと
//       1) INSERT を発行
//       2) 別途 SELECT で生成された user_id を取り直す
//     の 2 往復が必要。RETURNING なら 1 往復で済む。
//   - さらに「INSERT と取得のあいだに別トランザクションが割り込む」問題も起きない
//     （アトミック）。

const findUserByIDSQL = `
SELECT user_id, user_name, user_email, created_at, updated_at, deleted_at
FROM users
WHERE user_id = $1
  AND deleted_at IS NULL
`

// ↑ 論理削除の扱い:
//   - users テーブルは行を DELETE せず `deleted_at` に削除日時を立てる運用。
//   - FindByID は「未削除の行」のみ返す（= `deleted_at IS NULL` で絞る）。
//   - こうしておくと、削除済みのユーザーが API から見えなくなる。
//   - カテゴリ/タスク側にも部分 index `WHERE deleted_at IS NULL` を張っているので
//     絞り込みも高速。

// --- Create ---------------------------------------------------------------

// Create は新しい User を作成し、DB が発行した ID / 日時を埋めた *user.User を返す。
func (r *UserRepository) Create(ctx context.Context, params user.CreateParams) (*user.User, error) {
	// QueryRow は「1 行返るクエリ」を実行する。結果は Scan で取り出す。
	// ctx を渡すと、context がキャンセルされた時にクエリも中断される。
	row := r.pool.QueryRow(ctx, createUserSQL, params.UserName, params.UserEmail)

	// Scan 先の変数を宣言。
	// RETURNING で返ってくるカラムの**順序と型**に合わせる。
	// 順序がズレると「name に user_id が入る」等の壊れ方をするので注意。
	var (
		id        int64
		name      string
		email     string
		createdAt time.Time
		updatedAt time.Time
		// deleted_at は NULL 許容カラム。
		// *time.Time（ポインタ）で受けると、NULL → nil、非 NULL → 実体への
		// ポインタ、という形で pgx が詰めてくれる。
		deletedAt *time.Time
	)

	err := row.Scan(&id, &name, &email, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		// --- UNIQUE 制約違反の検知 ---
		// pgx は PostgreSQL エラーを `*pgconn.PgError` 型で返す。
		// `errors.As` は「エラーチェーンの中に指定の型があれば取り出す」API。
		// `fmt.Errorf("...: %w", ...)` でラップされていても透過的に取れる。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// SQLSTATE 23505 = unique_violation。user_email の UNIQUE 制約違反。
			// DB 固有のエラーを、ここで domain 語彙に**翻訳**する。
			// usecase/handler は 23505 を知らずに済む。
			return nil, user.ErrUserEmailAlreadyExists
		}
		// それ以外は想定外エラー。%w で原因をラップして返す。
		//   - %w でラップすると呼び出し側は errors.Is / errors.As で原因を辿れる
		//   - %v だと文字列化されるだけでチェーンが切れる
		return nil, fmt.Errorf("postgres: create user: %w", err)
	}

	// DB から戻ってきた値で User を復元する。
	// NewUser ではなく Reconstruct を使うのは、既に ID/timestamps が確定しているため。
	return user.Reconstruct(user.ReconstructParams{
		UserID:    id,
		UserName:  name,
		UserEmail: email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	})
}

// --- FindByID -------------------------------------------------------------

// FindByID は指定 ID の User を取得する。
//   - 見つからない / 論理削除済み: user.ErrUserNotFound
//   - それ以外のエラー: ラップして返す
func (r *UserRepository) FindByID(ctx context.Context, userID int64) (*user.User, error) {
	row := r.pool.QueryRow(ctx, findUserByIDSQL, userID)

	var (
		id        int64
		name      string
		email     string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err := row.Scan(&id, &name, &email, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		// --- 行が無い場合の検知 ---
		// pgx.ErrNoRows は「QueryRow で 0 行だった」ときの sentinel。
		// 標準の database/sql の sql.ErrNoRows と対応する概念。
		// errors.Is で「チェーンのどこかにこの値が含まれるか」を判定する。
		if errors.Is(err, pgx.ErrNoRows) {
			// 論理削除済みの行も WHERE deleted_at IS NULL で弾かれてここに来る。
			// つまり「削除済み」も「そもそも存在しない」も同じ ErrUserNotFound で返る。
			// この仕様は interface のドキュメントで明示している。
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("postgres: find user by id: %w", err)
	}

	return user.Reconstruct(user.ReconstructParams{
		UserID:    id,
		UserName:  name,
		UserEmail: email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	})
}
