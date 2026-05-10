package task

import (
	"time"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// CreateResponse は POST /tasks のレスポンス DTO。
type CreateResponse struct {
	TaskID int64 `json:"taskId"`
}

// ToCreateResponse はドメインモデルをレスポンス DTO に変換する。
func ToCreateResponse(t *domaintask.Task) CreateResponse {
	return CreateResponse{TaskID: t.TaskID}
}

// ListItemResponse は GET /tasks のレスポンス要素 DTO。
type ListItemResponse struct {
	TaskID          int64      `json:"taskId"`
	CategoryID      *int64     `json:"categoryId"`
	TaskTitle       string     `json:"taskTitle"`
	TaskDescription *string    `json:"taskDescription"`
	TaskStatus      string     `json:"taskStatus"`
	TaskDueAt       *time.Time `json:"taskDueAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ToListResponse はドメインモデルのスライスをレスポンス DTO のスライスに変換する。
func ToListResponse(tasks []*domaintask.Task) []ListItemResponse {
	resp := make([]ListItemResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, ListItemResponse{
			TaskID:          t.TaskID,
			CategoryID:      t.CategoryID,
			TaskTitle:       t.TaskTitle,
			TaskDescription: t.TaskDescription,
			TaskStatus:      string(t.TaskStatus),
			TaskDueAt:       t.TaskDueAt,
			CreatedAt:       t.CreatedAt,
			UpdatedAt:       t.UpdatedAt,
		})
	}
	return resp
}

// GetResponse は GET /tasks/{taskId} のレスポンス DTO。
// 一覧と違い、categories と JOIN した categoryName を返す(categoryId ではない)。
type GetResponse struct {
	TaskID          int64      `json:"taskId"`
	CategoryName    *string    `json:"categoryName"`
	TaskTitle       string     `json:"taskTitle"`
	TaskDescription *string    `json:"taskDescription"`
	TaskStatus      string     `json:"taskStatus"`
	TaskDueAt       *time.Time `json:"taskDueAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ToGetResponse はドメインモデルをレスポンス DTO に変換する。
func ToGetResponse(d *domaintask.Detail) GetResponse {
	return GetResponse{
		TaskID:          d.TaskID,
		CategoryName:    d.CategoryName,
		TaskTitle:       d.TaskTitle,
		TaskDescription: d.TaskDescription,
		TaskStatus:      string(d.TaskStatus),
		TaskDueAt:       d.TaskDueAt,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}
