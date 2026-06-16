package handler

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	handlercategory "github.com/maya-konnichiha/todo-list-backend/internal/handler/category"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/middleware"
	handlertask "github.com/maya-konnichiha/todo-list-backend/internal/handler/task"
	handleruser "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"
	uccategory "github.com/maya-konnichiha/todo-list-backend/internal/usecase/category"
	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// Deps はアプリケーション全体の依存関係を集約した構造体。
// registry.NewDeps で生成し、NewRouter に渡す。
type Deps struct {
	Logger       *slog.Logger
	DBPool       *pgxpool.Pool
	CreateUserUC *ucuser.CreateUserUsecase
	GetUserUC    *ucuser.GetUserUsecase

	CreateCategoryUC *uccategory.CreateCategoryUsecase
	ListCategoriesUC *uccategory.ListCategoriesUsecase
	DeleteCategoryUC *uccategory.DeleteCategoryUsecase

	CreateTaskUC *uctask.CreateTaskUsecase
	ListTasksUC  *uctask.ListTasksUsecase
	GetTaskUC    *uctask.GetTaskUsecase
	UpdateTaskUC *uctask.UpdateTaskUsecase
	PatchTaskUC  *uctask.PatchTaskUsecase
	DeleteTaskUC *uctask.DeleteTaskUsecase
}

// NewRouter はアプリケーションのルーティングを構築する。
// 各エンティティごとの RegisterXxxRoutes を呼び出す。
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	handleruser.RegisterUserRoutes(mux, handleruser.Deps{
		CreateUserUC: d.CreateUserUC,
		GetUserUC:    d.GetUserUC,
	})

	handlercategory.RegisterCategoryRoutes(mux, handlercategory.Deps{
		CreateCategoryUC: d.CreateCategoryUC,
		ListCategoriesUC: d.ListCategoriesUC,
		DeleteCategoryUC: d.DeleteCategoryUC,
	})

	handlertask.RegisterTaskRoutes(mux, handlertask.Deps{
		CreateTaskUC: d.CreateTaskUC,
		ListTasksUC:  d.ListTasksUC,
		GetTaskUC:    d.GetTaskUC,
		UpdateTaskUC: d.UpdateTaskUC,
		PatchTaskUC:  d.PatchTaskUC,
		DeleteTaskUC: d.DeleteTaskUC,
	})

	return middleware.Logging(d.Logger, mux)
}
