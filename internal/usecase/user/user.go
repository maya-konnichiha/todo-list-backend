package user

import (
	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// Usecase はユーザー関連のビジネスロジックを束ねる。
// repository は interface 型で受け取り、実装に依存しない。
type Usecase struct {
	repo domainuser.UserRepository
}

// New は Usecase を生成する。
func New(repo domainuser.UserRepository) *Usecase {
	return &Usecase{repo: repo}
}
