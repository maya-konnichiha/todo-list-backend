package category

import "context"

// CreateParams はカテゴリ作成時のパラメータ。
// usecase 層が handler から受け取った入力をこの形式に詰め替えて repository に渡す。
type CreateParams struct {
	UserID       int64
	CategoryName string
}

// CategoryRepository はカテゴリ永続化層の振る舞いを宣言する。
// 実装は internal/infrastructure/postgres/repository にあり、
// usecase はこの interface 経由で触る。
// Repository のインターフェースを定義することで Usecase が Repository に依存するのを防ぐ。
type CategoryRepository interface {
	Create(ctx context.Context, params CreateParams) (*Category, error)
	ListByUserID(ctx context.Context, userID int64) ([]*Category, error)
	SoftDelete(ctx context.Context, userID int64, categoryID int64) error

	// ExistsForUser は (categoryID, userID) のアクティブなカテゴリが存在するかを返す。
	// task 作成/更新時に「指定された categoryId が本人のものか」を検証するために使う。
	ExistsForUser(ctx context.Context, userID int64, categoryID int64) (bool, error)
}
