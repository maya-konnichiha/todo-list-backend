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
// 主に repository 層で DB から取得した行を User に復元する際に使用する。
// （ドメインテストで User を明示的に組み立てたい場合にも利用する）
//
// usecase 層で「作成リクエスト」を表現したい場合は、このコンストラクタではなく
// UserRepository.Create に CreateParams を渡す流れを使う。
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
