CREATE TABLE tasks (
    task_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories (category_id) ON DELETE SET NULL,
    task_title VARCHAR(255) NOT NULL,
    task_description TEXT,
    task_status VARCHAR(20) NOT NULL DEFAULT 'todo'
        CHECK (task_status IN ('todo', 'in_progress', 'done')),
    task_due_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_user_id
    ON tasks (user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_tasks_category_id
    ON tasks (category_id)
    WHERE deleted_at IS NULL;
