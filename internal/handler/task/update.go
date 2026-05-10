package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
)

// UpdateHandler は PUT /tasks/{taskId} を処理するハンドラ。
type UpdateHandler struct {
	uc *uctask.UpdateTaskUsecase
}

// NewUpdateHandler は UpdateHandler を生成する。
func NewUpdateHandler(uc *uctask.UpdateTaskUsecase) *UpdateHandler {
	return &UpdateHandler{uc: uc}
}

// Handle は PUT /tasks/{taskId} を処理する。
func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateTaskRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "リクエストの形式が不正です")
		return
	}
	if msg := validateUpdateRequest(req); msg != "" {
		errhandler.BadRequest(w, "INVALID_REQUEST", msg)
		return
	}

	status, _ := domaintask.ParseStatus(req.TaskStatus)
	if err := h.uc.Execute(r.Context(), uctask.UpdateInput{
		UserID:          userID,
		TaskID:          taskID,
		CategoryID:      req.CategoryID,
		TaskTitle:       req.TaskTitle,
		TaskDescription: req.TaskDescription,
		TaskStatus:      status,
		TaskDueAt:       req.TaskDueAt,
	}); err != nil {
		switch {
		case errors.Is(err, domaintask.ErrNotFound):
			errhandler.NotFound(w, "TASK_NOT_FOUND", "タスクが見つかりません")
		case errors.Is(err, domaintask.ErrCategoryNotAccessible):
			errhandler.BadRequest(w, "INVALID_REQUEST", "指定された categoryId は存在しないか、本人のカテゴリではありません")
		default:
			errhandler.Internal(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
