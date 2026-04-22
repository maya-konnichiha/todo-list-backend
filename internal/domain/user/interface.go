package user

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

var ErrUserEmailAlreadyExists = errors.New("user_email already exists")

type CreateParams struct {
	UserName  string
	UserEmail string
}

type UserRepository interface {
	Create(ctx context.Context, params CreateParams) (*User, error)
	FindByID(ctx context.Context, userID int64) (*User, error)
}
