package user

import "time"

// User はユーザーエンティティ。domain 層は外部ライブラリに依存させない。
type User struct {
	UserID    int64
	UserName  string
	UserEmail string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
