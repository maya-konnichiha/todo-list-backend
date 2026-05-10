package task

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
)

// GetHandler は GET /tasks/{taskId} を処理するハンドラ。
type GetHandler struct {
	uc *uctask.GetTaskUsecase
}

// NewGetHandler は GetHandler を生成する。
func NewGetHandler(uc *uctask.GetTaskUsecase) *GetHandler {
	return &GetHandler{uc: uc}
}

// Handle は GET /tasks/{taskId} を処理する。
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromHeader(r)
	if err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "X-User-Id ヘッダーは正の整数で指定してください")
		return
	}

	taskID, err := strconv.ParseInt(r.PathValue("taskId"), 10, 64)
	if err != nil || taskID <= 0 {
		errhandler.BadRequest(w, "INVALID_REQUEST", "taskId は正の整数で指定してください")
		return
	}

	found, err := h.uc.Execute(r.Context(), uctask.GetInput{UserID: userID, TaskID: taskID})
	if err != nil {
		if errors.Is(err, domaintask.ErrNotFound) {
			errhandler.NotFound(w, "TASK_NOT_FOUND", "タスクが見つかりません")
			return
		}
		errhandler.Internal(w, err)
		return
	}

	resp := ToGetResponse(found)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode get task response", slog.Any("error", err))
	}
}
