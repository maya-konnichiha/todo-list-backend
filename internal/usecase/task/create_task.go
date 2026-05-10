package task

import (
	"context"
	"time"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// CreateTaskUsecase はタスク作成のユースケース。
// category_id が指定された場合に「本人のカテゴリかどうか」を検証するため
// CategoryRepository にも依存する。
type CreateTaskUsecase struct {
	taskRepo     domaintask.TaskRepository
	categoryRepo domaincategory.CategoryRepository
}

// NewCreateTaskUsecase は CreateTaskUsecase を生成する。
func NewCreateTaskUsecase(taskRepo domaintask.TaskRepository, categoryRepo domaincategory.CategoryRepository) *CreateTaskUsecase {
	return &CreateTaskUsecase{taskRepo: taskRepo, categoryRepo: categoryRepo}
}

// CreateInput は Execute の入力 DTO。
type CreateInput struct {
	UserID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskDueAt       *time.Time
}

// Execute はタスクを作成して返す。
// CategoryID が指定された場合は、それが本人の所有するアクティブなカテゴリかを検証する。
func (u *CreateTaskUsecase) Execute(ctx context.Context, in CreateInput) (*domaintask.Task, error) {
	if in.CategoryID != nil {
		ok, err := u.categoryRepo.ExistsForUser(ctx, in.UserID, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domaintask.ErrCategoryNotAccessible
		}
	}
	return u.taskRepo.Create(ctx, domaintask.CreateParams{
		UserID:          in.UserID,
		CategoryID:      in.CategoryID,
		TaskTitle:       in.TaskTitle,
		TaskDescription: in.TaskDescription,
		TaskDueAt:       in.TaskDueAt,
	})
}
