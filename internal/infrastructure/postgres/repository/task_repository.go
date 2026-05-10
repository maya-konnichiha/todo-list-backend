package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
)

// taskRow は tasks テーブルから取得した 1 行を表す内部表現。
type taskRow struct {
	TaskID          int64
	UserID          int64
	CategoryID      *int64
	TaskTitle       string
	TaskDescription *string
	TaskStatus      string
	TaskDueAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TaskRepository は domaintask.TaskRepository の PostgreSQL 実装。
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository は TaskRepository を生成する。
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// Create はタスクを INSERT し、DB で採番された行を返す。
// task_status は DB の DEFAULT 'todo' に任せる。
func (r *TaskRepository) Create(ctx context.Context, params domaintask.CreateParams) (*domaintask.Task, error) {
	const query = `
		INSERT INTO tasks (user_id, category_id, task_title, task_description, task_due_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING task_id, user_id, category_id, task_title, task_description, task_status, task_due_at, created_at, updated_at
	`
	var row taskRow
	err := r.pool.QueryRow(ctx, query,
		params.UserID,
		params.CategoryID,
		params.TaskTitle,
		params.TaskDescription,
		params.TaskDueAt,
	).Scan(
		&row.TaskID,
		&row.UserID,
		&row.CategoryID,
		&row.TaskTitle,
		&row.TaskDescription,
		&row.TaskStatus,
		&row.TaskDueAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return r.toDomainModel(row), nil
}

// ListByFilter は本人のアクティブなタスクをフィルタ条件付きで返す。
// フィルタ条件は SQL 内で COALESCE 風に `$N IS NULL OR col = $N` で表現し、
// 動的 SQL を回避する。
func (r *TaskRepository) ListByFilter(ctx context.Context, filter domaintask.ListFilter) ([]*domaintask.Task, error) {
	const query = `
		SELECT task_id, user_id, category_id, task_title, task_description,
		       task_status, task_due_at, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND ($2::varchar IS NULL OR task_status = $2)
		  AND ($3::bigint  IS NULL OR category_id = $3)
		ORDER BY created_at ASC, task_id ASC
	`
	var statusArg *string
	if filter.TaskStatus != nil {
		s := string(*filter.TaskStatus)
		statusArg = &s
	}

	rows, err := r.pool.Query(ctx, query, filter.UserID, statusArg, filter.CategoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*domaintask.Task, 0)
	for rows.Next() {
		var row taskRow
		if err := rows.Scan(
			&row.TaskID,
			&row.UserID,
			&row.CategoryID,
			&row.TaskTitle,
			&row.TaskDescription,
			&row.TaskStatus,
			&row.TaskDueAt,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, r.toDomainModel(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindDetailByID は tasks と categories を LEFT JOIN して詳細を返す。
// 自分のタスクで、かつ deleted_at IS NULL のみが対象。
// category が削除済みの場合や category_id NULL の場合は CategoryName を nil で返す。
func (r *TaskRepository) FindDetailByID(ctx context.Context, userID int64, taskID int64) (*domaintask.Detail, error) {
	const query = `
		SELECT t.task_id, t.user_id, t.category_id, t.task_title, t.task_description,
		       t.task_status, t.task_due_at, t.created_at, t.updated_at,
		       c.category_name
		FROM tasks t
		LEFT JOIN categories c
		       ON c.category_id = t.category_id
		      AND c.deleted_at IS NULL
		WHERE t.task_id  = $1
		  AND t.user_id  = $2
		  AND t.deleted_at IS NULL
	`
	var row taskRow
	var categoryName *string
	err := r.pool.QueryRow(ctx, query, taskID, userID).Scan(
		&row.TaskID,
		&row.UserID,
		&row.CategoryID,
		&row.TaskTitle,
		&row.TaskDescription,
		&row.TaskStatus,
		&row.TaskDueAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&categoryName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domaintask.ErrNotFound
		}
		return nil, err
	}
	return &domaintask.Detail{
		Task:         r.toDomainModel(row),
		CategoryName: categoryName,
	}, nil
}

// Update はタスクを全フィールド一括で上書きする(PUT)。
// 対象が存在しない/他人のものなら domaintask.ErrNotFound。
func (r *TaskRepository) Update(ctx context.Context, params domaintask.UpdateParams) error {
	const query = `
		UPDATE tasks
		SET category_id      = $3,
		    task_title       = $4,
		    task_description = $5,
		    task_status      = $6,
		    task_due_at      = $7,
		    updated_at       = CURRENT_TIMESTAMP
		WHERE task_id   = $1
		  AND user_id   = $2
		  AND deleted_at IS NULL
	`
	tag, err := r.pool.Exec(ctx, query,
		params.TaskID,
		params.UserID,
		params.CategoryID,
		params.TaskTitle,
		params.TaskDescription,
		string(params.TaskStatus),
		params.TaskDueAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domaintask.ErrNotFound
	}
	return nil
}

// Patch は指定されたフィールドのみを更新する(PATCH)。
//
// COALESCE と CASE で「未指定 / NULL / 値」の 3 状態を SQL レベルで表現することで
// 動的 SQL の組み立てを回避している:
//   - 必須カラム(task_title, task_status):
//       COALESCE($n, col)  → $n=NULL なら現在値を維持
//   - nullable カラム(category_id, task_description, task_due_at):
//       CASE WHEN $set THEN $val ELSE col END
//       → $set=false なら現在値を維持、$set=true で $val=NULL なら NULL に更新
func (r *TaskRepository) Patch(ctx context.Context, params domaintask.PatchParams) error {
	const query = `
		UPDATE tasks
		SET task_title       = COALESCE($3::varchar, task_title),
		    task_status      = COALESCE($4::varchar, task_status),
		    category_id      = CASE WHEN $5::bool THEN $6::bigint      ELSE category_id      END,
		    task_description = CASE WHEN $7::bool THEN $8::text        ELSE task_description END,
		    task_due_at      = CASE WHEN $9::bool THEN $10::timestamptz ELSE task_due_at      END,
		    updated_at       = CURRENT_TIMESTAMP
		WHERE task_id   = $1
		  AND user_id   = $2
		  AND deleted_at IS NULL
	`
	var statusArg *string
	if params.TaskStatus != nil {
		s := string(*params.TaskStatus)
		statusArg = &s
	}

	tag, err := r.pool.Exec(ctx, query,
		params.TaskID, params.UserID,
		params.TaskTitle, statusArg,
		params.SetCategoryID, params.CategoryID,
		params.SetTaskDescription, params.TaskDescription,
		params.SetTaskDueAt, params.TaskDueAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domaintask.ErrNotFound
	}
	return nil
}

// SoftDelete は対象タスクの deleted_at に現在時刻をセットする。
func (r *TaskRepository) SoftDelete(ctx context.Context, userID int64, taskID int64) error {
	const query = `
		UPDATE tasks
		SET deleted_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE task_id   = $1
		  AND user_id   = $2
		  AND deleted_at IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, taskID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domaintask.ErrNotFound
	}
	return nil
}

// toDomainModel は tasks テーブル行を domaintask.Task に変換する。
func (r *TaskRepository) toDomainModel(row taskRow) *domaintask.Task {
	return domaintask.NewTask(domaintask.NewTaskParams{
		TaskID:          row.TaskID,
		UserID:          row.UserID,
		CategoryID:      row.CategoryID,
		TaskTitle:       row.TaskTitle,
		TaskDescription: row.TaskDescription,
		TaskStatus:      domaintask.Status(row.TaskStatus),
		TaskDueAt:       row.TaskDueAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	})
}
