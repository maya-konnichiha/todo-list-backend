package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/maya-konnichiha/todo-list-backend/internal/handler"
	"github.com/maya-konnichiha/todo-list-backend/internal/infrastructure/postgres"
	"github.com/maya-konnichiha/todo-list-backend/internal/registry"
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
	pool, err := postgres.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to PostgreSQL")

	// DI 配線は registry に集約
	deps := registry.NewDeps(registry.NewDepsParams{
		DB:     pool,
		Logger: slog.Default(),
	})

	router := handler.NewRouter(deps)

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
