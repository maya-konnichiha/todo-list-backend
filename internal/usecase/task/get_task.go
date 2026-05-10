package task

import (
	"context"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// GetTaskUsecase はタスク詳細取得のユースケース。
// categories と JOIN して category_name も一緒に返す。
type GetTaskUsecase struct {
	repo domaintask.TaskRepository
}

// NewGetTaskUsecase は GetTaskUsecase を生成する。
func NewGetTaskUsecase(repo domaintask.TaskRepository) *GetTaskUsecase {
	return &GetTaskUsecase{repo: repo}
}

// GetInput は Execute の入力 DTO。
type GetInput struct {
	UserID int64
	TaskID int64
}

// Execute は taskID で特定したタスクの詳細(category_name 込み)を返す。
// 存在しない/他人のものなら domaintask.ErrNotFound を返す。
func (u *GetTaskUsecase) Execute(ctx context.Context, in GetInput) (*domaintask.Detail, error) {
	return u.repo.FindDetailByID(ctx, in.UserID, in.TaskID)
}
