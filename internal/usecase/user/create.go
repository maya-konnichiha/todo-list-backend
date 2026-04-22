package user

import (
	"context"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// CreateInput は Create の入力 DTO。handler のリクエスト形式から独立させる。
type CreateInput struct {
	UserName  string
	UserEmail string
}

// Create はユーザーを作成して返す。
// バリデーションは handler 層で済んでいる前提。
func (uc *Usecase) Create(ctx context.Context, in CreateInput) (*domainuser.User, error) {
	u := &domainuser.User{
		UserName:  in.UserName,
		UserEmail: in.UserEmail,
	}
	return uc.repo.Create(ctx, u)
}
