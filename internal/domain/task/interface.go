package task

import (
	"context"
	"time"
)

// CreateParams はタスク作成時のパラメータ。
type CreateParams struct {
	UserID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskDueAt       *time.Time
}

// ListFilter はタスク一覧取得時のフィルタ条件。
// 各フィールドは nil の場合「そのフィールドではフィルタしない」を意味する。
type ListFilter struct {
	UserID     int64
	TaskStatus *Status
	CategoryID *int64
}

// UpdateParams はタスク全体更新(PUT)時のパラメータ。
// 全フィールドを送りつけて上書きする。nullable な DB カラムはポインタで「null も指定可」とする。
type UpdateParams struct {
	UserID          int64
	TaskID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskStatus      Status
	TaskDueAt       *time.Time
}

// PatchParams はタスク部分更新(PATCH)時のパラメータ。
//
// 必須カラム(TaskTitle / TaskStatus): nil = 未指定(変更しない), non-nil = その値で更新。
// nullable カラム(CategoryID / TaskDescription / TaskDueAt): "未指定 / NULL に更新 / 値で更新"
// の 3 状態を、Set* フラグとポインタ値の組で表現する。
//   - Set*=false                   → 未指定(変更しない)
//   - Set*=true,  Value=nil        → NULL に更新
//   - Set*=true,  Value=&value     → 値で更新
type PatchParams struct {
	UserID int64
	TaskID int64

	TaskTitle  *string
	TaskStatus *Status

	SetCategoryID      bool
	CategoryID         *int64
	SetTaskDescription bool
	TaskDescription    *string
	SetTaskDueAt       bool
	TaskDueAt          *time.Time
}

// TaskRepository はタスク永続化層の振る舞いを宣言する。
type TaskRepository interface {
	Create(ctx context.Context, params CreateParams) (*Task, error)
	ListByFilter(ctx context.Context, filter ListFilter) ([]*Task, error)
	FindDetailByID(ctx context.Context, userID int64, taskID int64) (*Detail, error)
	Update(ctx context.Context, params UpdateParams) error
	Patch(ctx context.Context, params PatchParams) error
	SoftDelete(ctx context.Context, userID int64, taskID int64) error
}
