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

// PatchHandler は PATCH /tasks/{taskId} を処理するハンドラ。
type PatchHandler struct {
	uc *uctask.PatchTaskUsecase
}

// NewPatchHandler は PatchHandler を生成する。
func NewPatchHandler(uc *uctask.PatchTaskUsecase) *PatchHandler {
	return &PatchHandler{uc: uc}
}

// Handle は PATCH /tasks/{taskId} を処理する。
// Optional[T] の Present/Value を usecase の PatchInput の Set*/Value にマッピングする。
func (h *PatchHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req PatchTaskRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "リクエストの形式が不正です")
		return
	}
	if msg := validatePatchRequest(req); msg != "" {
		errhandler.BadRequest(w, "INVALID_REQUEST", msg)
		return
	}

	in := uctask.PatchInput{
		UserID: userID,
		TaskID: taskID,
	}
	if req.TaskTitle.Present {
		in.TaskTitle = req.TaskTitle.Value
	}
	if req.TaskStatus.Present {
		s, _ := domaintask.ParseStatus(*req.TaskStatus.Value)
		in.TaskStatus = &s
	}
	if req.CategoryID.Present {
		in.SetCategoryID = true
		in.CategoryID = req.CategoryID.Value
	}
	if req.TaskDescription.Present {
		in.SetTaskDescription = true
		in.TaskDescription = req.TaskDescription.Value
	}
	if req.TaskDueAt.Present {
		in.SetTaskDueAt = true
		in.TaskDueAt = req.TaskDueAt.Value
	}

	if err := h.uc.Execute(r.Context(), in); err != nil {
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
