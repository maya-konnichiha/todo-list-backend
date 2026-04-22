package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// PostgreSQL の unique_violation エラーコード。
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgCodeUniqueViolation = "23505"

// Repository は domainuser.Repository の PostgreSQL 実装。
type Repository struct {
	pool *pgxpool.Pool
}

// New は Repository を生成する。
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create はユーザーを INSERT し、DB で採番された行を返す。
// email の UNIQUE 衝突は domainuser.ErrEmailAlreadyRegistered に変換する。
func (r *Repository) Create(ctx context.Context, u *domainuser.User) (*domainuser.User, error) {
	const query = `
		INSERT INTO users (user_name, user_email)
		VALUES ($1, $2)
		RETURNING user_id, user_name, user_email, created_at, updated_at
	`
	var created domainuser.User
	err := r.pool.QueryRow(ctx, query, u.UserName, u.UserEmail).Scan(
		&created.UserID,
		&created.UserName,
		&created.UserEmail,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCodeUniqueViolation {
			return nil, domainuser.ErrEmailAlreadyRegistered
		}
		return nil, err
	}
	return &created, nil
}
