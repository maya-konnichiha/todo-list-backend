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

// NewUserParams は User 生成時のパラメータ。
//
// このパラメータ構造体は 2 つの用途を兼ねている:
//
//  1. 新規作成前の一時オブジェクトを作る場合（usecase 層で使用）
//     → UserName / UserEmail のみ指定する。
//     UserID / CreatedAt / UpdatedAt は INSERT 時に DB 側で採番・設定されるため、
//     ゼロ値のままで良い。
//
//  2. DB から取得した行を User に復元する場合（repository 層で使用）
//     → DB から返ってきた全フィールドを指定する。
//
// Go ではフィールドを省略すると自動でゼロ値になる（int64 は 0、time.Time は
// time.Time{}）ため、用途 1 の呼び出しでも UserID/CreatedAt/UpdatedAt を
// 明示的に書く必要はない。
type NewUserParams struct {
	UserID    int64
	UserName  string
	UserEmail string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserOption は User のオプション設定を行う関数型。
type UserOption func(*User)

// NewUser はユーザーを生成するコンストラクタ。
// 必須パラメータは Params 構造体で、オプショナルなパラメータ(DeletedAt 等)は
// Functional Options で受け取る。
func NewUser(params NewUserParams, opts ...UserOption) *User {
	u := &User{
		UserID:    params.UserID,
		UserName:  params.UserName,
		UserEmail: params.UserEmail,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// WithDeletedAt は soft delete 時刻を設定するオプション。
func WithDeletedAt(deletedAt *time.Time) UserOption {
	return func(u *User) {
		u.DeletedAt = deletedAt
	}
}
