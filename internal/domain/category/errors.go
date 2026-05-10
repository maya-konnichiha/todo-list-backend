package category

import "errors"

var (
	// ErrNotFound は指定されたカテゴリが存在しない(または soft delete 済み)場合に返る。
	// 「他人のカテゴリにアクセスしようとした」ケースもここに集約し、
	// 権限エラー情報を露出させない。
	ErrNotFound = errors.New("category not found")
)
