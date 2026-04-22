package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/maya-konnichiha/todo-list-backend/internal/handler"
	"github.com/maya-konnichiha/todo-list-backend/internal/infrastructure/postgres"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using OS environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to PostgreSQL successfully")

	userRepo := postgres.NewUserRepository(pool)
	createUserUC := userusecase.NewCreateUser(userRepo)

	deps := handler.Deps{
		CreateUser: createUserUC,
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, deps)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("server starting on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
