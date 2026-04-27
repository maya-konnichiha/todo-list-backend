package handler

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	handleruser "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// Deps はアプリケーション全体の依存関係を集約した構造体。
// registry.NewDeps で生成し、NewRouter に渡す。
type Deps struct {
	Logger       *slog.Logger
	DBPool       *pgxpool.Pool
	CreateUserUC *ucuser.CreateUserUsecase
	GetUserUC    *ucuser.GetUserUsecase
}

// NewRouter はアプリケーションのルーティングを構築する。
// 各エンティティごとの RegisterXxxRoutes を呼び出す。
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	handleruser.RegisterUserRoutes(mux, handleruser.Deps{
		CreateUserUC: d.CreateUserUC,
		GetUserUC:    d.GetUserUC,
	})

	return mux
}
