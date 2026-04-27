package user

import (
	"time"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// CreateResponse は POST /users のレスポンス DTO。
type CreateResponse struct {
	UserID    int64     `json:"userId"`
	UserName  string    `json:"userName"`
	UserEmail string    `json:"userEmail"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToCreateResponse はドメインモデルをレスポンス DTO に変換する。
func ToCreateResponse(u *domainuser.User) CreateResponse {
	return CreateResponse{
		UserID:    u.UserID,
		UserName:  u.UserName,
		UserEmail: u.UserEmail,
		CreatedAt: u.CreatedAt,
	}
}

// GetResponse は GET /users/{userId} のレスポンス DTO。
type GetResponse struct {
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
}

// ToGetResponse はドメインモデルをレスポンス DTO に変換する。
func ToGetResponse(u *domainuser.User) GetResponse {
	return GetResponse{
		UserName:  u.UserName,
		UserEmail: u.UserEmail,
	}
}
