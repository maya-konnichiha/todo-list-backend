package task

import (
	"context"
	"time"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// UpdateTaskUsecase はタスク全体更新(PUT)のユースケース。
// PATCH と違って全フィールド送りつけて上書きする。
type UpdateTaskUsecase struct {
	taskRepo     domaintask.TaskRepository
	categoryRepo domaincategory.CategoryRepository
}

// NewUpdateTaskUsecase は UpdateTaskUsecase を生成する。
func NewUpdateTaskUsecase(taskRepo domaintask.TaskRepository, categoryRepo domaincategory.CategoryRepository) *UpdateTaskUsecase {
	return &UpdateTaskUsecase{taskRepo: taskRepo, categoryRepo: categoryRepo}
}

// UpdateInput は Execute の入力 DTO。
type UpdateInput struct {
	UserID          int64
	TaskID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskStatus      domaintask.Status
	TaskDueAt       *time.Time
}

// Execute はタスクを全体更新する。
// CategoryID が non-nil なら本人所有のカテゴリかを検証する。
// 対象タスクが存在しない/他人のものなら domaintask.ErrNotFound を返す。
func (u *UpdateTaskUsecase) Execute(ctx context.Context, in UpdateInput) error {
	if in.CategoryID != nil {
		ok, err := u.categoryRepo.ExistsForUser(ctx, in.UserID, *in.CategoryID)
		if err != nil {
			return err
		}
		if !ok {
			return domaintask.ErrCategoryNotAccessible
		}
	}
	return u.taskRepo.Update(ctx, domaintask.UpdateParams{
		UserID:          in.UserID,
		TaskID:          in.TaskID,
		CategoryID:      in.CategoryID,
		TaskTitle:       in.TaskTitle,
		TaskDescription: in.TaskDescription,
		TaskStatus:      in.TaskStatus,
		TaskDueAt:       in.TaskDueAt,
	})
}
