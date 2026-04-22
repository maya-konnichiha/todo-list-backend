package user

import "context"

// CreateParams はユーザー作成時のパラメータ。
// usecase 層が handler から受け取った入力をこの形式に詰め替えて repository に渡す。
type CreateParams struct {
	UserName  string
	UserEmail string
}

// UserRepository はユーザー永続化層の振る舞いを宣言する。
// 実装は internal/infrastructure/postgres/repository/user にあり、
// usecase はこの interface 経由で触る。
type UserRepository interface {
	Create(ctx context.Context, params CreateParams) (*User, error)
}
