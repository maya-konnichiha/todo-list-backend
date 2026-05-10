package task

import (
	"errors"
	"net/http"
	"strconv"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
)

// DeleteHandler は DELETE /tasks/{taskId} を処理するハンドラ。
type DeleteHandler struct {
	uc *uctask.DeleteTaskUsecase
}

// NewDeleteHandler は DeleteHandler を生成する。
func NewDeleteHandler(uc *uctask.DeleteTaskUsecase) *DeleteHandler {
	return &DeleteHandler{uc: uc}
}

// Handle は DELETE /tasks/{taskId} を処理する。
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	if err := h.uc.Execute(r.Context(), uctask.DeleteInput{UserID: userID, TaskID: taskID}); err != nil {
		if errors.Is(err, domaintask.ErrNotFound) {
			errhandler.NotFound(w, "TASK_NOT_FOUND", "タスクが見つかりません")
			return
		}
		errhandler.Internal(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
