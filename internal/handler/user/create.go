package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"time"
	"unicode/utf8"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

type createRequest struct {
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
}

type createResponse struct {
	UserID    int64     `json:"userId"`
	UserName  string    `json:"userName"`
	UserEmail string    `json:"userEmail"`
	CreatedAt time.Time `json:"createdAt"`
}

// Create は POST /users を処理する。
// 認証不要エンドポイント(新規ユーザー登録)。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
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

	created, err := h.createUserUC.Execute(r.Context(), ucuser.CreateInput{
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

	resp := createResponse{
		UserID:    created.UserID,
		UserName:  created.UserName,
		UserEmail: created.UserEmail,
		CreatedAt: created.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode create user response", slog.Any("error", err))
	}
}

// validateCreateRequest はリクエストを検証し、違反があれば人間向けメッセージを返す。
func validateCreateRequest(req createRequest) string {
	if req.UserName == "" {
		return "userName は必須です"
	}
	if utf8.RuneCountInString(req.UserName) > 50 {
		return "userName は 50 文字以内にしてください"
	}
	if req.UserEmail == "" {
		return "userEmail は必須です"
	}
	if len(req.UserEmail) > 255 {
		return "userEmail は 255 文字以内にしてください"
	}
	if _, err := mail.ParseAddress(req.UserEmail); err != nil {
		return "userEmail の形式が不正です"
	}
	return ""
}
