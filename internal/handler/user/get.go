package user

import (
	"errors"
	"net/http"
	"strconv"

	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

type GetHandler struct {
	usecase *userusecase.GetUser
}

func NewGetHandler(uc *userusecase.GetUser) *GetHandler {
	return &GetHandler{usecase: uc}
}

func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "リクエストの形式が不正です")
		return
	}

	u, err := h.usecase.Execute(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNotFound):
			writeErrorJSON(w, http.StatusNotFound, "ユーザーが見つかりません")
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "内部エラーが発生しました")
		}
		return
	}

	writeJSON(w, http.StatusOK, newGetUserResponse(u))
}
