package category

import "time"

// Category はカテゴリエンティティ。domain 層は外部ライブラリに依存させない。
type Category struct {
	CategoryID   int64
	UserID       int64
	CategoryName string
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewCategoryParams は Category 生成時のパラメータ。
//
// 主に repository 層で DB から取得した行を Category に復元する際に使用する。
type NewCategoryParams struct {
	CategoryID   int64
	UserID       int64
	CategoryName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CategoryOption は Category のオプション設定を行う関数型。
type CategoryOption func(*Category)

// NewCategory はカテゴリを生成するコンストラクタ。
// 必須パラメータは Params 構造体で、オプショナルなパラメータ(DeletedAt 等)は
// Functional Options で受け取る。
func NewCategory(params NewCategoryParams, opts ...CategoryOption) *Category {
	c := &Category{
		CategoryID:   params.CategoryID,
		UserID:       params.UserID,
		CategoryName: params.CategoryName,
		CreatedAt:    params.CreatedAt,
		UpdatedAt:    params.UpdatedAt,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithDeletedAt は soft delete 時刻を設定するオプション。
func WithDeletedAt(deletedAt *time.Time) CategoryOption {
	return func(c *Category) {
		c.DeletedAt = deletedAt
	}
}
