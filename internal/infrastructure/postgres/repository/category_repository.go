package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
)

// categoryRow は categories テーブルから取得した 1 行を表す内部表現。
// SQL の Scan 先として使い、toDomainModel で domain モデルに変換する。
type categoryRow struct {
	CategoryID   int64
	UserID       int64
	CategoryName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CategoryRepository は domaincategory.CategoryRepository の PostgreSQL 実装。
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository は CategoryRepository を生成する。
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// Create はカテゴリを INSERT し、DB で採番された行を返す。
func (r *CategoryRepository) Create(ctx context.Context, params domaincategory.CreateParams) (*domaincategory.Category, error) {
	const query = `
		INSERT INTO categories (user_id, category_name)
		VALUES ($1, $2)
		RETURNING category_id, user_id, category_name, created_at, updated_at
	`
	var row categoryRow
	err := r.pool.QueryRow(ctx, query, params.UserID, params.CategoryName).Scan(
		&row.CategoryID,
		&row.UserID,
		&row.CategoryName,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return r.toDomainModel(row), nil
}

// ListByUserID は本人(userID)が持つアクティブなカテゴリの一覧を created_at 昇順で返す。
// 件数が 0 でも nil ではなく空スライスを返すのは「JSON 化したとき null ではなく [] にする」ため。
func (r *CategoryRepository) ListByUserID(ctx context.Context, userID int64) ([]*domaincategory.Category, error) {
	const query = `
		SELECT category_id, user_id, category_name, created_at, updated_at
		FROM categories
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC, category_id ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*domaincategory.Category, 0)
	for rows.Next() {
		var row categoryRow
		if err := rows.Scan(
			&row.CategoryID,
			&row.UserID,
			&row.CategoryName,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, r.toDomainModel(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

// SoftDelete は対象カテゴリの deleted_at に現在時刻をセットする。
// 対象が見つからない(他人のもの・既に削除済み・存在しない)場合は domaincategory.ErrNotFound を返す。
func (r *CategoryRepository) SoftDelete(ctx context.Context, userID int64, categoryID int64) error {
	const query = `
		UPDATE categories
		SET deleted_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE category_id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, categoryID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domaincategory.ErrNotFound
	}
	return nil
}

// ExistsForUser は (categoryID, userID) のアクティブなカテゴリが存在するかを返す。
// task 作成/更新時に「指定された categoryId が本人のものか」を検証するために使う。
func (r *CategoryRepository) ExistsForUser(ctx context.Context, userID int64, categoryID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM categories
			WHERE category_id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, categoryID, userID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// toDomainModel は categories テーブル行を domaincategory.Category に変換する。
func (r *CategoryRepository) toDomainModel(row categoryRow) *domaincategory.Category {
	return domaincategory.NewCategory(domaincategory.NewCategoryParams{
		CategoryID:   row.CategoryID,
		UserID:       row.UserID,
		CategoryName: row.CategoryName,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	})
}
