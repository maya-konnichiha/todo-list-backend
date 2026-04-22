package repository

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

// userRow は users テーブルから取得した 1 行を表す内部表現。
// SQL の Scan 先として使い、toDomainModel で domain モデルに変換する。
// sqlc を使っていないためこの型を自前で用意している。
type userRow struct {
	UserID    int64
	UserName  string
	UserEmail string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository は domainuser.UserRepository の PostgreSQL 実装。
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository は UserRepository を生成する。
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create はユーザーを INSERT し、DB で採番された行を返す。
// email の UNIQUE 衝突は domainuser.ErrEmailAlreadyRegistered に変換する。
func (r *UserRepository) Create(ctx context.Context, params domainuser.CreateParams) (*domainuser.User, error) {
	const query = `
		INSERT INTO users (user_name, user_email)
		VALUES ($1, $2)
		RETURNING user_id, user_name, user_email, created_at, updated_at
	`
	var row userRow
	err := r.pool.QueryRow(ctx, query, params.UserName, params.UserEmail).Scan(
		&row.UserID,
		&row.UserName,
		&row.UserEmail,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCodeUniqueViolation {
			return nil, domainuser.ErrEmailAlreadyRegistered
		}
		return nil, err
	}
	return r.toDomainModel(row), nil
}

// toDomainModel は users テーブル行を domainuser.User に変換する。
// DB の内部表現(userRow)と domain 層のエンティティの間の境界をこの関数に閉じ込める。
func (r *UserRepository) toDomainModel(row userRow) *domainuser.User {
	return domainuser.NewUser(domainuser.NewUserParams{
		UserID:    row.UserID,
		UserName:  row.UserName,
		UserEmail: row.UserEmail,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	})
}
