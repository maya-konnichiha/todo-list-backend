package category

import (
	"context"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
)

// CreateCategoryUsecase はカテゴリ作成のユースケース。
// 1 アクション = 1 構造体。repository は interface 型で受け取り、実装に依存しない。
type CreateCategoryUsecase struct {
	repo domaincategory.CategoryRepository
}

// NewCreateCategoryUsecase は CreateCategoryUsecase を生成する。
func NewCreateCategoryUsecase(repo domaincategory.CategoryRepository) *CreateCategoryUsecase {
	return &CreateCategoryUsecase{repo: repo}
}

// CreateInput は Execute の入力 DTO。handler のリクエスト形式から独立させる。
type CreateInput struct {
	UserID       int64
	CategoryName string
}

// Execute はカテゴリを作成して返す。
// バリデーションは handler 層で済んでいる前提。
func (u *CreateCategoryUsecase) Execute(ctx context.Context, in CreateInput) (*domaincategory.Category, error) {
	return u.repo.Create(ctx, domaincategory.CreateParams{
		UserID:       in.UserID,
		CategoryName: in.CategoryName,
	})
}
