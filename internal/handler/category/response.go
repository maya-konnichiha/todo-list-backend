package category

import (
	"time"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
)

// ListItemResponse は GET /categories のレスポンス要素 DTO。
type ListItemResponse struct {
	CategoryID   int64     `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ToListResponse はドメインモデルのスライスをレスポンス DTO のスライスに変換する。
// 件数が 0 でも nil ではなく空スライスを返し、JSON 化時に [] になるようにする。
func ToListResponse(categories []*domaincategory.Category) []ListItemResponse {
	resp := make([]ListItemResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, ListItemResponse{
			CategoryID:   c.CategoryID,
			CategoryName: c.CategoryName,
			CreatedAt:    c.CreatedAt,
		})
	}
	return resp
}
