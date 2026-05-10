package category

import (
	"context"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
)

// DeleteCategoryUsecase はカテゴリ削除のユースケース。
// 削除は soft delete(deleted_at に時刻をセット)で行う。
type DeleteCategoryUsecase struct {
	repo domaincategory.CategoryRepository
}

// NewDeleteCategoryUsecase は DeleteCategoryUsecase を生成する。
func NewDeleteCategoryUsecase(repo domaincategory.CategoryRepository) *DeleteCategoryUsecase {
	return &DeleteCategoryUsecase{repo: repo}
}

// DeleteInput は Execute の入力 DTO。
// 他人のカテゴリを削除できないよう、本人 UserID とセットで渡す。
type DeleteInput struct {
	UserID     int64
	CategoryID int64
}

// Execute は対象カテゴリを soft delete する。
// 対象が存在しない(または既に削除済み、または他人のもの)場合は
// domaincategory.ErrNotFound を返す。
func (u *DeleteCategoryUsecase) Execute(ctx context.Context, in DeleteInput) error {
	return u.repo.SoftDelete(ctx, in.UserID, in.CategoryID)
}
