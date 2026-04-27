package user

import (
	"context"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// GetUserUsecase はユーザー取得のユースケース。
// 1 アクション = 1 構造体。repository は interface 型で受け取り、実装に依存しない。
type GetUserUsecase struct {
	repo domainuser.UserRepository
}

// NewGetUserUsecase は GetUserUsecase を生成する。
func NewGetUserUsecase(repo domainuser.UserRepository) *GetUserUsecase {
	return &GetUserUsecase{repo: repo}
}

// GetInput は Execute の入力 DTO。handler のリクエスト形式から独立させる。
type GetInput struct {
	UserID int64
}

// Execute は userID で特定したユーザーを返す。
// 存在しない(または soft delete 済み)場合は domainuser.ErrNotFound を返す。
func (u *GetUserUsecase) Execute(ctx context.Context, in GetInput) (*domainuser.User, error) {
	return u.repo.FindByID(ctx, in.UserID)
}
