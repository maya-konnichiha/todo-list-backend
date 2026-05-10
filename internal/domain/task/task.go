package task

import "time"

// Task はタスクエンティティ。domain 層は外部ライブラリに依存させない。
// nullable な DB カラムは Go 側でもポインタ型で表現する。
type Task struct {
	TaskID          int64
	UserID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskStatus      Status
	TaskDueAt       *time.Time
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewTaskParams は Task 生成時のパラメータ。
// 主に repository 層で DB から取得した行を Task に復元する際に使用する。
type NewTaskParams struct {
	TaskID          int64
	UserID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskStatus      Status
	TaskDueAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TaskOption は Task のオプション設定を行う関数型。
type TaskOption func(*Task)

// NewTask はタスクを生成するコンストラクタ。
func NewTask(params NewTaskParams, opts ...TaskOption) *Task {
	t := &Task{
		TaskID:          params.TaskID,
		UserID:          params.UserID,
		CategoryID:      params.CategoryID,
		TaskTitle:       params.TaskTitle,
		TaskDescription: params.TaskDescription,
		TaskStatus:      params.TaskStatus,
		TaskDueAt:       params.TaskDueAt,
		CreatedAt:       params.CreatedAt,
		UpdatedAt:       params.UpdatedAt,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// WithDeletedAt は soft delete 時刻を設定するオプション。
func WithDeletedAt(deletedAt *time.Time) TaskOption {
	return func(t *Task) {
		t.DeletedAt = deletedAt
	}
}

// Detail は GET /tasks/{taskId} のレスポンス用に、categories と JOIN した結果を表す。
// CategoryName はカテゴリ未指定/削除済みの時 nil。
type Detail struct {
	*Task
	CategoryName *string
}
