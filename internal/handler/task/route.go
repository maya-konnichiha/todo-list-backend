package task

import (
	"net/http"

	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
)

// Deps は task ハンドラの依存関係。各アクションのユースケースをフィールドとして持つ。
type Deps struct {
	CreateTaskUC *uctask.CreateTaskUsecase
	ListTasksUC  *uctask.ListTasksUsecase
	GetTaskUC    *uctask.GetTaskUsecase
	UpdateTaskUC *uctask.UpdateTaskUsecase
	PatchTaskUC  *uctask.PatchTaskUsecase
	DeleteTaskUC *uctask.DeleteTaskUsecase
}

// RegisterTaskRoutes は task 関連のルートを mux に登録する。
func RegisterTaskRoutes(mux *http.ServeMux, d Deps) {
	createH := NewCreateHandler(d.CreateTaskUC)
	listH := NewListHandler(d.ListTasksUC)
	getH := NewGetHandler(d.GetTaskUC)
	updateH := NewUpdateHandler(d.UpdateTaskUC)
	patchH := NewPatchHandler(d.PatchTaskUC)
	deleteH := NewDeleteHandler(d.DeleteTaskUC)

	mux.HandleFunc("POST /tasks", createH.Handle)
	mux.HandleFunc("GET /tasks", listH.Handle)
	mux.HandleFunc("GET /tasks/{taskId}", getH.Handle)
	mux.HandleFunc("PUT /tasks/{taskId}", updateH.Handle)
	mux.HandleFunc("PATCH /tasks/{taskId}", patchH.Handle)
	mux.HandleFunc("DELETE /tasks/{taskId}", deleteH.Handle)
}
