package task

import (
	"context"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// ListTasksUsecase はタスク一覧取得のユースケース。
type ListTasksUsecase struct {
	repo domaintask.TaskRepository
}

// NewListTasksUsecase は ListTasksUsecase を生成する。
func NewListTasksUsecase(repo domaintask.TaskRepository) *ListTasksUsecase {
	return &ListTasksUsecase{repo: repo}
}

// ListInput は Execute の入力 DTO。
// TaskStatus / CategoryID は nil で「フィルタしない」を意味する。
type ListInput struct {
	UserID     int64
	TaskStatus *domaintask.Status
	CategoryID *int64
}

// Execute はフィルタに合うタスクの一覧を返す。
func (u *ListTasksUsecase) Execute(ctx context.Context, in ListInput) ([]*domaintask.Task, error) {
	return u.repo.ListByFilter(ctx, domaintask.ListFilter{
		UserID:     in.UserID,
		TaskStatus: in.TaskStatus,
		CategoryID: in.CategoryID,
	})
}
