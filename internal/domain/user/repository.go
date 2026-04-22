package user

import "context"

// Repository はユーザー永続化層の振る舞いを宣言する。
// 実装は internal/repository/user にあり、usecase はこの interface 経由で触る。
type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
}
