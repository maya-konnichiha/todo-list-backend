package registry

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maya-konnichiha/todo-list-backend/internal/handler"
	"github.com/maya-konnichiha/todo-list-backend/internal/infrastructure/postgres/repository"
	categoryUsecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/category"
	taskUsecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
	userUsecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// NewDepsParams は NewDeps に渡す設定。
type NewDepsParams struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

// NewDeps は全ての依存関係を一箇所で管理し、handler.Deps を生成する。
func NewDeps(params NewDepsParams) handler.Deps {
	return handler.Deps{
		Logger:       params.Logger,
		DBPool:       params.DB,
		CreateUserUC: NewCreateUserUsecase(params.DB),
		GetUserUC:    NewGetUserUsecase(params.DB),

		CreateCategoryUC: NewCreateCategoryUsecase(params.DB),
		ListCategoriesUC: NewListCategoriesUsecase(params.DB),
		DeleteCategoryUC: NewDeleteCategoryUsecase(params.DB),

		CreateTaskUC: NewCreateTaskUsecase(params.DB),
		ListTasksUC:  NewListTasksUsecase(params.DB),
		GetTaskUC:    NewGetTaskUsecase(params.DB),
		UpdateTaskUC: NewUpdateTaskUsecase(params.DB),
		PatchTaskUC:  NewPatchTaskUsecase(params.DB),
		DeleteTaskUC: NewDeleteTaskUsecase(params.DB),
	}
}

// NewCreateUserUsecase はユーザー作成ユースケースを生成する。
func NewCreateUserUsecase(pool *pgxpool.Pool) *userUsecase.CreateUserUsecase {
	repo := repository.NewUserRepository(pool)
	return userUsecase.NewCreateUserUsecase(repo)
}

// NewGetUserUsecase はユーザー取得ユースケースを生成する。
func NewGetUserUsecase(pool *pgxpool.Pool) *userUsecase.GetUserUsecase {
	repo := repository.NewUserRepository(pool)
	return userUsecase.NewGetUserUsecase(repo)
}

// NewCreateCategoryUsecase はカテゴリ作成ユースケースを生成する。
func NewCreateCategoryUsecase(pool *pgxpool.Pool) *categoryUsecase.CreateCategoryUsecase {
	repo := repository.NewCategoryRepository(pool)
	return categoryUsecase.NewCreateCategoryUsecase(repo)
}

// NewListCategoriesUsecase はカテゴリ一覧取得ユースケースを生成する。
func NewListCategoriesUsecase(pool *pgxpool.Pool) *categoryUsecase.ListCategoriesUsecase {
	repo := repository.NewCategoryRepository(pool)
	return categoryUsecase.NewListCategoriesUsecase(repo)
}

// NewDeleteCategoryUsecase はカテゴリ削除ユースケースを生成する。
func NewDeleteCategoryUsecase(pool *pgxpool.Pool) *categoryUsecase.DeleteCategoryUsecase {
	repo := repository.NewCategoryRepository(pool)
	return categoryUsecase.NewDeleteCategoryUsecase(repo)
}

// NewCreateTaskUsecase はタスク作成ユースケースを生成する。
// task UC のうち category 検証が必要なものは TaskRepository と CategoryRepository を両方受け取る。
func NewCreateTaskUsecase(pool *pgxpool.Pool) *taskUsecase.CreateTaskUsecase {
	taskRepo := repository.NewTaskRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	return taskUsecase.NewCreateTaskUsecase(taskRepo, categoryRepo)
}

// NewListTasksUsecase はタスク一覧取得ユースケースを生成する。
func NewListTasksUsecase(pool *pgxpool.Pool) *taskUsecase.ListTasksUsecase {
	repo := repository.NewTaskRepository(pool)
	return taskUsecase.NewListTasksUsecase(repo)
}

// NewGetTaskUsecase はタスク詳細取得ユースケースを生成する。
func NewGetTaskUsecase(pool *pgxpool.Pool) *taskUsecase.GetTaskUsecase {
	repo := repository.NewTaskRepository(pool)
	return taskUsecase.NewGetTaskUsecase(repo)
}

// NewUpdateTaskUsecase はタスク全体更新ユースケースを生成する。
func NewUpdateTaskUsecase(pool *pgxpool.Pool) *taskUsecase.UpdateTaskUsecase {
	taskRepo := repository.NewTaskRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	return taskUsecase.NewUpdateTaskUsecase(taskRepo, categoryRepo)
}

// NewPatchTaskUsecase はタスク部分更新ユースケースを生成する。
func NewPatchTaskUsecase(pool *pgxpool.Pool) *taskUsecase.PatchTaskUsecase {
	taskRepo := repository.NewTaskRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	return taskUsecase.NewPatchTaskUsecase(taskRepo, categoryRepo)
}

// NewDeleteTaskUsecase はタスク削除ユースケースを生成する。
func NewDeleteTaskUsecase(pool *pgxpool.Pool) *taskUsecase.DeleteTaskUsecase {
	repo := repository.NewTaskRepository(pool)
	return taskUsecase.NewDeleteTaskUsecase(repo)
}
