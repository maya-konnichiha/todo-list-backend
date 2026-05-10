package task

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
)

// CreateHandler は POST /tasks を処理するハンドラ。
type CreateHandler struct {
	uc *uctask.CreateTaskUsecase
}

// NewCreateHandler は CreateHandler を生成する。
func NewCreateHandler(uc *uctask.CreateTaskUsecase) *CreateHandler {
	return &CreateHandler{uc: uc}
}

// Handle は POST /tasks を処理する。
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromHeader(r)
	if err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "X-User-Id ヘッダーは正の整数で指定してください")
		return
	}

	var req CreateTaskRequest
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

	created, err := h.uc.Execute(r.Context(), uctask.CreateInput{
		UserID:          userID,
		CategoryID:      req.CategoryID,
		TaskTitle:       req.TaskTitle,
		TaskDescription: req.TaskDescription,
		TaskDueAt:       req.TaskDueAt,
	})
	if err != nil {
		if errors.Is(err, domaintask.ErrCategoryNotAccessible) {
			errhandler.BadRequest(w, "INVALID_REQUEST", "指定された categoryId は存在しないか、本人のカテゴリではありません")
			return
		}
		errhandler.Internal(w, err)
		return
	}

	resp := ToCreateResponse(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode create task response", slog.Any("error", err))
	}
}
