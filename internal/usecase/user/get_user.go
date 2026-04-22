package user

import (
	"context"

	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

type GetUser struct {
	repo userdomain.UserRepository
}

func NewGetUser(repo userdomain.UserRepository) *GetUser {
	return &GetUser{repo: repo}
}

func (uc *GetUser) Execute(ctx context.Context, userID int64) (*userdomain.User, error) {
	return uc.repo.FindByID(ctx, userID)
}
