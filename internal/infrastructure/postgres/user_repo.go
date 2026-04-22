package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

var _ user.UserRepository = (*UserRepository)(nil)

const createUserSQL = `
INSERT INTO users (user_name, user_email)
VALUES ($1, $2)
RETURNING user_id, user_name, user_email, created_at, updated_at, deleted_at
`

const findUserByIDSQL = `
SELECT user_id, user_name, user_email, created_at, updated_at, deleted_at
FROM users
WHERE user_id = $1
  AND deleted_at IS NULL
`

func (r *UserRepository) Create(ctx context.Context, params user.CreateParams) (*user.User, error) {
	row := r.pool.QueryRow(ctx, createUserSQL, params.UserName, params.UserEmail)

	var (
		id        int64
		name      string
		email     string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err := row.Scan(&id, &name, &email, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, user.ErrUserEmailAlreadyExists
		}
		return nil, fmt.Errorf("postgres: create user: %w", err)
	}

	return user.Reconstruct(user.ReconstructParams{
		UserID:    id,
		UserName:  name,
		UserEmail: email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	})
}

func (r *UserRepository) FindByID(ctx context.Context, userID int64) (*user.User, error) {
	row := r.pool.QueryRow(ctx, findUserByIDSQL, userID)

	var (
		id        int64
		name      string
		email     string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err := row.Scan(&id, &name, &email, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("postgres: find user by id: %w", err)
	}

	return user.Reconstruct(user.ReconstructParams{
		UserID:    id,
		UserName:  name,
		UserEmail: email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	})
}
