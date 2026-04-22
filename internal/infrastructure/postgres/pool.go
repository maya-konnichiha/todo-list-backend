package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool は DATABASE_URL 文字列から pgxpool を生成し、疎通確認(Ping)まで行う。
// Ping 失敗時はプールをクローズしてからエラーを返すため、呼び出し側は返ってきた
// pool を defer Close する責務だけ持てばよい。
func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return pool, nil
}
