package user

import (
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// Handler はユーザー関連の HTTP ハンドラを束ねる。
type Handler struct {
	uc *ucuser.Usecase
}

// New は Handler を生成する。
func New(uc *ucuser.Usecase) *Handler {
	return &Handler{uc: uc}
}
