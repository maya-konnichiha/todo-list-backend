package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainuser "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// PostgreSQL の unique_violation エラーコード。
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgCodeUniqueViolation = "23505"

// Repository は domainuser.UserRepository の PostgreSQL 実装。
type Repository struct {
	pool *pgxpool.Pool
}

// New は Repository を生成する。
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create はユーザーを INSERT し、DB で採番された行を返す。
// email の UNIQUE 衝突は domainuser.ErrEmailAlreadyRegistered に変換する。
func (r *Repository) Create(ctx context.Context, params domainuser.CreateParams) (*domainuser.User, error) {
	const query = `
		INSERT INTO users (user_name, user_email)
		VALUES ($1, $2)
		RETURNING user_id, user_name, user_email, created_at, updated_at
	`
	var (
		userID    int64
		userName  string
		userEmail string
		createdAt time.Time
		updatedAt time.Time
	)
	err := r.pool.QueryRow(ctx, query, params.UserName, params.UserEmail).Scan(
		&userID,
		&userName,
		&userEmail,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCodeUniqueViolation {
			return nil, domainuser.ErrEmailAlreadyRegistered
		}
		return nil, err
	}
	return domainuser.NewUser(domainuser.NewUserParams{
		UserID:    userID,
		UserName:  userName,
		UserEmail: userEmail,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}), nil
}
