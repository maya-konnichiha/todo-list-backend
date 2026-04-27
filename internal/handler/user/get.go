package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// GetHandler は GET /users/{userId} を処理するハンドラ。
// 1 アクション = 1 ハンドラ構造体。依存するユースケースのみをフィールドとして持つ。
type GetHandler struct {
	uc *ucuser.GetUserUsecase
}

// NewGetHandler は GetHandler を生成する。
func NewGetHandler(uc *ucuser.GetUserUsecase) *GetHandler {
	return &GetHandler{uc: uc}
}

// Handle は GET /users/{userId} を処理する。
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil || userID <= 0 {
		errhandler.BadRequest(w, "INVALID_REQUEST", "userId は正の整数で指定してください")
		return
	}

	found, err := h.uc.Execute(r.Context(), ucuser.GetInput{UserID: userID})
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			errhandler.NotFound(w, "USER_NOT_FOUND", "ユーザーが見つかりません")
			return
		}
		errhandler.Internal(w, err)
		return
	}

	resp := ToGetResponse(found)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode get user response", slog.Any("error", err))
	}
}
