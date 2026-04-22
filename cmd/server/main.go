// Package main はアプリのエントリポイント。
//
// ここで行うこと:
//  1. 環境変数の読み込み（.env）
//  2. DB 接続プール（pgxpool）の作成と疎通確認
//  3. 依存の組み立て（Repository → Usecase → handler.Deps）
//  4. http.ServeMux にルート登録
//  5. HTTP サーバー起動
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
	// --------------------------------------------------------------------
	// 1. .env の読み込み
	//
	// 本番（Cloud Run 等）では .env を置かずに環境変数を直接セットするので、
	// ファイルが無くても続行する（Fatal にはしない）。
	// --------------------------------------------------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using OS environment variables")
	}

	// --------------------------------------------------------------------
	// 2. 設定値の取得
	//
	// os.Getenv は未設定でも空文字を返すだけ。必須のものはここでゼロチェック。
	// --------------------------------------------------------------------
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// --------------------------------------------------------------------
	// 3. DB 接続プールの作成
	//
	// pgxpool:
	//   - DB との TCP 接続をプールして使い回す仕組み
	//   - 並行リクエストで接続を張り直さずに済む
	//
	// タイムアウト付き ctx:
	//   - DB が無応答でも 5 秒で諦める（fail-fast）
	//
	// defer pool.Close():
	//   - main 終了時にプールを閉じる = 保持している DB 接続を全部切る
	// --------------------------------------------------------------------
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

	// --------------------------------------------------------------------
	// 4. 依存の組み立て（DI 配線）
	//
	// pool → Repository → Usecase → handler.Deps の順に組み上げる。
	// 各層は interface で依存を受け取るので、ここで具象を詰める。
	// --------------------------------------------------------------------
	userRepo := postgres.NewUserRepository(pool)
	createUserUC := userusecase.NewCreateUser(userRepo)

	deps := handler.Deps{
		CreateUser: createUserUC,
	}

	// --------------------------------------------------------------------
	// 5. http.ServeMux にルート登録
	//
	// Gin 時代は gin.Default() + handler.RegisterRoutes で 2 行だった。
	// 標準ではミドルウェア（ログ / panic 回復）が付かない分、自分で足す必要がある。
	// 今回は学習段階なので省略。必要になったら以下で足す:
	//
	//   var handler http.Handler = mux
	//   handler = loggingMiddleware(handler)
	//   handler = recoveryMiddleware(handler)
	// --------------------------------------------------------------------
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, deps)

	// --------------------------------------------------------------------
	// 6. HTTP サーバー起動
	//
	// http.Server を自前で構築する理由:
	//   - ReadHeaderTimeout を明示できる（slowloris 攻撃対策）
	//     net/http の素の ListenAndServe はタイムアウト無しで、時間をかけてヘッダを
	//     送るだけで接続を占有する攻撃が成り立つ
	//   - 将来 graceful shutdown（server.Shutdown）を入れる時の下地になる
	//
	// Gin の r.Run() は内部で http.Server を組んでいる。今回はそれを自前で書いているだけ。
	// --------------------------------------------------------------------
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
