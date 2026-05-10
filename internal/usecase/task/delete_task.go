package task

import (
	"context"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// DeleteTaskUsecase はタスク削除のユースケース。
// 削除は soft delete(deleted_at に時刻をセット)で行う。
type DeleteTaskUsecase struct {
	repo domaintask.TaskRepository
}

// NewDeleteTaskUsecase は DeleteTaskUsecase を生成する。
func NewDeleteTaskUsecase(repo domaintask.TaskRepository) *DeleteTaskUsecase {
	return &DeleteTaskUsecase{repo: repo}
}

// DeleteInput は Execute の入力 DTO。
type DeleteInput struct {
	UserID int64
	TaskID int64
}

// Execute は対象タスクを soft delete する。
// 対象が存在しない/他人のものなら domaintask.ErrNotFound を返す。
func (u *DeleteTaskUsecase) Execute(ctx context.Context, in DeleteInput) error {
	return u.repo.SoftDelete(ctx, in.UserID, in.TaskID)
}
