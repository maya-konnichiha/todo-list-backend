package user

import (
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// Handler はユーザー関連の HTTP ハンドラを束ねる。
// アクションごとのユースケースをフィールドとして持つ。
type Handler struct {
	createUserUC *ucuser.CreateUserUsecase
}

// New は Handler を生成する。
func New(createUserUC *ucuser.CreateUserUsecase) *Handler {
	return &Handler{createUserUC: createUserUC}
}
