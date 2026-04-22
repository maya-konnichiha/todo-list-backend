package user

import (
	"context"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// CreateUserUsecase はユーザー作成のユースケース。
// 1 アクション = 1 構造体。repository は interface 型で受け取り、実装に依存しない。
type CreateUserUsecase struct {
	repo domainuser.UserRepository
}

// NewCreateUserUsecase は CreateUserUsecase を生成する。
func NewCreateUserUsecase(repo domainuser.UserRepository) *CreateUserUsecase {
	return &CreateUserUsecase{repo: repo}
}

// CreateInput は Execute の入力 DTO。handler のリクエスト形式から独立させる。
type CreateInput struct {
	UserName  string
	UserEmail string
}

// Execute はユーザーを作成して返す。
// バリデーションは handler 層で済んでいる前提。
func (u *CreateUserUsecase) Execute(ctx context.Context, in CreateInput) (*domainuser.User, error) {
	return u.repo.Create(ctx, domainuser.CreateParams{
		UserName:  in.UserName,
		UserEmail: in.UserEmail,
	})
}
