package user

import (
	"context"

	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

type CreateUser struct {
	repo userdomain.UserRepository
}

func NewCreateUser(repo userdomain.UserRepository) *CreateUser {
	return &CreateUser{repo: repo}
}

type CreateUserParams struct {
	UserName  string
	UserEmail string
}

func (uc *CreateUser) Execute(ctx context.Context, params CreateUserParams) (*userdomain.User, error) {
	newUser, err := userdomain.NewUser(userdomain.NewUserParams{
		UserName:  params.UserName,
		UserEmail: params.UserEmail,
	})
	if err != nil {
		return nil, err
	}

	created, err := uc.repo.Create(ctx, userdomain.CreateParams{
		UserName:  newUser.UserName,
		UserEmail: newUser.UserEmail,
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}
