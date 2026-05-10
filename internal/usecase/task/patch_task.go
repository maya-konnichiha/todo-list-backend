package task

import (
	"context"
	"time"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// PatchTaskUsecase はタスク部分更新(PATCH)のユースケース。
// 指定されたフィールドだけを更新する。
type PatchTaskUsecase struct {
	taskRepo     domaintask.TaskRepository
	categoryRepo domaincategory.CategoryRepository
}

// NewPatchTaskUsecase は PatchTaskUsecase を生成する。
func NewPatchTaskUsecase(taskRepo domaintask.TaskRepository, categoryRepo domaincategory.CategoryRepository) *PatchTaskUsecase {
	return &PatchTaskUsecase{taskRepo: taskRepo, categoryRepo: categoryRepo}
}

// PatchInput は Execute の入力 DTO。
// PatchParams と同様、nullable フィールドは「未指定 / NULL / 値」の 3 状態を
// Set* フラグとポインタ値の組で表現する。
type PatchInput struct {
	UserID int64
	TaskID int64

	TaskTitle  *string
	TaskStatus *domaintask.Status

	SetCategoryID      bool
	CategoryID         *int64
	SetTaskDescription bool
	TaskDescription    *string
	SetTaskDueAt       bool
	TaskDueAt          *time.Time
}

// Execute は指定されたフィールドのみを更新する。
// SetCategoryID=true かつ CategoryID が non-nil の時のみ、本人所有のカテゴリかを検証する
// (NULL に更新する場合は検証不要)。
// 対象タスクが存在しない/他人のものなら domaintask.ErrNotFound を返す。
func (u *PatchTaskUsecase) Execute(ctx context.Context, in PatchInput) error {
	if in.SetCategoryID && in.CategoryID != nil {
		ok, err := u.categoryRepo.ExistsForUser(ctx, in.UserID, *in.CategoryID)
		if err != nil {
			return err
		}
		if !ok {
			return domaintask.ErrCategoryNotAccessible
		}
	}
	return u.taskRepo.Patch(ctx, domaintask.PatchParams{
		UserID: in.UserID,
		TaskID: in.TaskID,

		TaskTitle:  in.TaskTitle,
		TaskStatus: in.TaskStatus,

		SetCategoryID:      in.SetCategoryID,
		CategoryID:         in.CategoryID,
		SetTaskDescription: in.SetTaskDescription,
		TaskDescription:    in.TaskDescription,
		SetTaskDueAt:       in.SetTaskDueAt,
		TaskDueAt:          in.TaskDueAt,
	})
}
