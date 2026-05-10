package category

import (
	"context"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
)

// ListCategoriesUsecase はカテゴリ一覧取得のユースケース。
type ListCategoriesUsecase struct {
	repo domaincategory.CategoryRepository
}

// NewListCategoriesUsecase は ListCategoriesUsecase を生成する。
func NewListCategoriesUsecase(repo domaincategory.CategoryRepository) *ListCategoriesUsecase {
	return &ListCategoriesUsecase{repo: repo}
}

// ListInput は Execute の入力 DTO。
type ListInput struct {
	UserID int64
}

// Execute は本人(userID)が持つアクティブなカテゴリの一覧を返す。
// 件数が 0 でも nil ではなく空スライスを返すのは repository の責務。
func (u *ListCategoriesUsecase) Execute(ctx context.Context, in ListInput) ([]*domaincategory.Category, error) {
	return u.repo.ListByUserID(ctx, in.UserID)
}
