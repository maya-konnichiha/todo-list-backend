package user

import (
	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

type CreateUserResponse struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

func newCreateUserResponse(u *userdomain.User) CreateUserResponse {
	return CreateUserResponse{
		UserID:    u.UserID,
		UserName:  u.UserName,
		UserEmail: u.UserEmail,
	}
}

type GetUserResponse struct {
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
}

func newGetUserResponse(u *userdomain.User) GetUserResponse {
	return GetUserResponse{
		UserName:  u.UserName,
		UserEmail: u.UserEmail,
	}
}
