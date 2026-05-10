package task

import (
	"time"
	"unicode/utf8"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

const (
	taskTitleMaxLen       = 255
	taskDescriptionMaxLen = 10000
)

// CreateTaskRequest は POST /tasks のリクエスト DTO。
// userId は X-User-Id ヘッダーから取り出すためボディには含めない。
type CreateTaskRequest struct {
	CategoryID      *int64     `json:"categoryId"`
	TaskTitle       string     `json:"taskTitle"`
	TaskDescription *string    `json:"taskDescription"`
	TaskDueAt       *time.Time `json:"taskDueAt"`
}

// validateCreateRequest はリクエスト形式の検証を行う。
func validateCreateRequest(req CreateTaskRequest) string {
	if msg := validateTaskTitle(req.TaskTitle); msg != "" {
		return msg
	}
	if msg := validateTaskDescription(req.TaskDescription); msg != "" {
		return msg
	}
	if req.CategoryID != nil && *req.CategoryID <= 0 {
		return "categoryId は正の整数で指定してください"
	}
	return ""
}

// UpdateTaskRequest は PUT /tasks/{taskId} のリクエスト DTO。
// 全体更新なので全フィールドを送る前提だが、nullable な DB カラムに対応して
// CategoryID / TaskDescription / TaskDueAt は明示的に null を送れるようポインタ型にする。
type UpdateTaskRequest struct {
	CategoryID      *int64     `json:"categoryId"`
	TaskTitle       string     `json:"taskTitle"`
	TaskDescription *string    `json:"taskDescription"`
	TaskStatus      string     `json:"taskStatus"`
	TaskDueAt       *time.Time `json:"taskDueAt"`
}

// validateUpdateRequest はリクエスト形式の検証を行う。
func validateUpdateRequest(req UpdateTaskRequest) string {
	if msg := validateTaskTitle(req.TaskTitle); msg != "" {
		return msg
	}
	if msg := validateTaskDescription(req.TaskDescription); msg != "" {
		return msg
	}
	if req.TaskStatus == "" {
		return "taskStatus は必須です"
	}
	if _, ok := domaintask.ParseStatus(req.TaskStatus); !ok {
		return "taskStatus は todo / in_progress / done のいずれかで指定してください"
	}
	if req.CategoryID != nil && *req.CategoryID <= 0 {
		return "categoryId は正の整数で指定してください"
	}
	return ""
}

// PatchTaskRequest は PATCH /tasks/{taskId} のリクエスト DTO。
// 3 状態(未指定 / null / 値) を表現するため Optional[T] を使う。
// 必須カラム(taskTitle / taskStatus) は null を許可しないため、
// Optional.Value=nil が来た場合は handler 側で 400 を返す。
type PatchTaskRequest struct {
	CategoryID      Optional[int64]     `json:"categoryId"`
	TaskTitle       Optional[string]    `json:"taskTitle"`
	TaskDescription Optional[string]    `json:"taskDescription"`
	TaskStatus      Optional[string]    `json:"taskStatus"`
	TaskDueAt       Optional[time.Time] `json:"taskDueAt"`
}

// validatePatchRequest はリクエスト形式の検証を行う。
// 必須カラムが指定された場合 (Present=true) は null を弾く。
// nullable カラムは null も値もどちらも許可する。
func validatePatchRequest(req PatchTaskRequest) string {
	if !req.CategoryID.Present &&
		!req.TaskTitle.Present &&
		!req.TaskDescription.Present &&
		!req.TaskStatus.Present &&
		!req.TaskDueAt.Present {
		return "更新するフィールドを 1 つ以上指定してください"
	}

	if req.TaskTitle.Present {
		if req.TaskTitle.Value == nil {
			return "taskTitle は null にできません"
		}
		if msg := validateTaskTitle(*req.TaskTitle.Value); msg != "" {
			return msg
		}
	}
	if req.TaskStatus.Present {
		if req.TaskStatus.Value == nil {
			return "taskStatus は null にできません"
		}
		if _, ok := domaintask.ParseStatus(*req.TaskStatus.Value); !ok {
			return "taskStatus は todo / in_progress / done のいずれかで指定してください"
		}
	}
	if req.TaskDescription.Present && req.TaskDescription.Value != nil {
		if msg := validateTaskDescription(req.TaskDescription.Value); msg != "" {
			return msg
		}
	}
	if req.CategoryID.Present && req.CategoryID.Value != nil && *req.CategoryID.Value <= 0 {
		return "categoryId は正の整数で指定してください"
	}
	return ""
}

func validateTaskTitle(title string) string {
	if title == "" {
		return "taskTitle は必須です"
	}
	if utf8.RuneCountInString(title) > taskTitleMaxLen {
		return "taskTitle は 255 文字以内にしてください"
	}
	return ""
}

func validateTaskDescription(desc *string) string {
	if desc == nil {
		return ""
	}
	if utf8.RuneCountInString(*desc) > taskDescriptionMaxLen {
		return "taskDescription は 10000 文字以内にしてください"
	}
	return ""
}
