package user

import "errors"

var (
	// ErrNotFound は指定されたユーザーが存在しない(または soft delete 済み)場合に返る。
	ErrNotFound = errors.New("user not found")

	// ErrEmailAlreadyRegistered は既にアクティブなユーザーで同じ email が使われている場合に返る。
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)
