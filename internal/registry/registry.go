package registry

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maya-konnichiha/todo-list-backend/internal/handler"
	"github.com/maya-konnichiha/todo-list-backend/internal/infrastructure/postgres/repository"
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
