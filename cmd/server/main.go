// Package main はアプリのエントリポイント。
//
// ここで行うこと:
//  1. 環境変数の読み込み（.env）
//  2. DB 接続プール（pgxpool）の作成と疎通確認
//  3. 依存の組み立て（Repository → Usecase → handler.Deps）
//  4. Gin エンジンの起動 + ルート登録
//  5. HTTP サーバー起動
//
// なぜ main.go に配線を集中させるのか（DI の集約点）:
//   - 依存（DB 接続・ロガー・usecase 等）を作る責務を**一箇所**に閉じ込める
//   - 下位の層（handler / usecase / domain）は「外から渡される」前提で書ける
//   - テストは main.go を経由せず、各層を個別に偽物依存で組み立てて動かせる
//   - プロジェクトが大きくなったら registry/ パッケージに切り出す余地を残しつつ、
//     Todo 規模では 1 ファイルで十分
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
	// godotenv.Load() は「.env ファイルを読んで環境変数にセット」する。
	// 本番（Cloud Run 等）では .env を置かずに環境変数を直接セットするので、
	// ファイルが無くても続行する（Fatal にはしない）。
	// --------------------------------------------------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using OS environment variables")
	}

	// --------------------------------------------------------------------
	// 2. 設定値の取得
	//
	// os.Getenv は未設定でも空文字を返すだけでエラーにはならない。
	// 必須の値は手動でゼロチェックして、無ければ早期失敗させる（fail-fast）。
	// --------------------------------------------------------------------
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080" // 環境変数が無ければデフォルト値
	}

	// --------------------------------------------------------------------
	// 3. DB 接続プールの作成
	//
	// pgxpool とは:
	//   - 「DB との TCP 接続を複数保持して、必要に応じて貸し出す」プール。
	//   - HTTP サーバーは並行（複数ゴルーチン）でリクエストを捌く。
	//     各リクエストが DB 接続を必要とする度に新規接続を張ると、
	//     接続のオープン/クローズが重く、DB 側の接続上限にも当たる。
	//   - プールは接続を**使い回す**ので、両方の問題を解決する。
	//
	// タイムアウト付き context を使う理由:
	//   - DB サーバーが落ちている等で応答が無い時、無限に待たないため。
	//   - 5 秒は実測に基づく値ではなく、学習用の妥当値。
	//     本番ではもっと短く（1〜2 秒）することもある。
	//
	// defer pool.Close() の意義:
	//   - main 終了時にプールを閉じる = 保持している接続を全部切る。
	//   - 閉じないと、プロセスは死んでも DB 側のセッションがしばらく残り、
	//     接続上限を圧迫する。
	//   - Go の defer はスタック LIFO で実行されるので、最初に書いておけば
	//     後でどこで return しても確実に呼ばれる。
	// --------------------------------------------------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // WithTimeout で確保したリソースを解放する（ctx を使い終わったら呼ぶ慣習）

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		// log.Fatalf は「メッセージを出してから os.Exit(1)」する。
		// defer は呼ばれないので、ここまで来る前に確保した defer（今はまだ無い）は失われる点に注意。
		log.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Ping で実際に 1 回接続を確立して、DB に届くかを確認する。
	// プール作成時点では接続は遅延確立されるので、Ping しないと
	// 起動は成功したように見えて最初のリクエストで初めて失敗することになる。
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to PostgreSQL successfully")

	// --------------------------------------------------------------------
	// 4. 依存の組み立て（DI 配線）
	//
	// 依存方向の流れ:
	//   pool → Repository → Usecase → handler.Deps → RegisterRoutes
	//
	// 上の層が下の層を「知る」のではなく、**上が下を組み立てて渡す**。
	// これにより handler は「Repository の実装が Postgres であること」を知らずに済む。
	// --------------------------------------------------------------------
	userRepo := postgres.NewUserRepository(pool)
	createUserUC := userusecase.NewCreateUser(userRepo)

	deps := handler.Deps{
		CreateUser: createUserUC,
		// 将来: GetUser, CreateCategory, CreateTask, ... をここに並べる
	}

	// --------------------------------------------------------------------
	// 5. Gin エンジン起動
	//
	// gin.Default() = gin.New() + Logger + Recovery の 2 つのミドルウェア付き。
	//   - Logger  : リクエストログを標準出力に出す
	//   - Recovery: ハンドラ内 panic を拾って 500 を返す（プロセス死亡を防ぐ）
	//
	// 本番ではロガーを構造化（JSON 出力）したくなるので gin.New() + 自作ミドルウェア
	// が一般的。学習段階では gin.Default() でコンパクトに。
	// --------------------------------------------------------------------
	r := gin.Default()
	handler.RegisterRoutes(r, deps)

	// --------------------------------------------------------------------
	// 6. HTTP サーバー起動
	//
	// r.Run は内部で http.ListenAndServe を呼ぶ。
	// Ctrl+C で止まるとプロセスが即死 → defer pool.Close() は呼ばれる
	// （Go の defer は panic と Fatal では呼ばれないが、シグナルで main が
	//  正常終了する場合は呼ばれる。ただし pool.Close() は「きれいに切りたい」
	//  時のためのもので、緊急停止では省かれても致命的ではない）。
	//
	// より堅牢にするなら: signal.NotifyContext で SIGINT/SIGTERM を受け取り、
	// http.Server.Shutdown で graceful shutdown する。
	// material-hub/cmd/main.go を参照。
	// --------------------------------------------------------------------
	log.Printf("server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
