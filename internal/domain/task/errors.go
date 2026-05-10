package task

import "errors"

var (
	// ErrNotFound は指定されたタスクが存在しない(または soft delete 済み、または他人のもの)
	// 場合に返る。「他人のタスクへのアクセス」も「存在しない」と同等に扱い、
	// 権限情報を露出させない。
	ErrNotFound = errors.New("task not found")

	// ErrCategoryNotAccessible は指定された category_id が存在しない、
	// または本人の所有でない場合に返る。
	// FK 違反は「存在しない」を意味するが、他人所有のカテゴリには FK は通ってしまうので
	// アプリ側でこのエラーを返す必要がある。
	ErrCategoryNotAccessible = errors.New("category not accessible")
)
