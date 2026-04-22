package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	handlerpkg "github.com/maya-konnichiha/todo-list-backend/internal/handler"
	handleruser "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"
	repouser "github.com/maya-konnichiha/todo-list-backend/internal/repository/user"
	ucuser "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to create connection pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("connected to PostgreSQL")

	// DI 配線: repository -> usecase -> handler -> router
	userRepo := repouser.New(pool)
	userUC := ucuser.New(userRepo)
	userH := handleruser.New(userUC)

	router := handlerpkg.NewRouter(userH)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("server starting", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, router); err != nil {
		slog.Error("server exited", slog.Any("error", err))
		os.Exit(1)
	}
}
