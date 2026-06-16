package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// CreateHandler は POST /users を処理するハンドラ。
// 1 アクション = 1 ハンドラ構造体。依存するユースケースのみをフィールドとして持つ。
type CreateHandler struct {
	uc *ucuser.CreateUserUsecase
}

// NewCreateHandler は CreateHandler を生成する。
func NewCreateHandler(uc *ucuser.CreateUserUsecase) *CreateHandler {
	return &CreateHandler{uc: uc}
}

// Handle は POST /users を処理する。
// 認証不要エンドポイント(新規ユーザー登録)。
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "リクエストの形式が不正です")
		return
	}
	if msg := validateCreateRequest(req); msg != "" {
		errhandler.BadRequest(w, "INVALID_REQUEST", msg)
		return
	}

	created, err := h.uc.Execute(r.Context(), ucuser.CreateInput{
		UserName:  req.UserName,
		UserEmail: req.UserEmail,
	})
	if err != nil {
		if errors.Is(err, domainuser.ErrEmailAlreadyRegistered) {
			errhandler.Conflict(w, "EMAIL_ALREADY_REGISTERED", "このメールアドレスは既に登録されています")
			return
		}
		errhandler.Internal(w, err)
		return
	}

	resp := ToCreateResponse(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode create user response", slog.Any("error", err))
	}// 難しい
}
