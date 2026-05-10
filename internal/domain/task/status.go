package task

// Status は task の状態を表す。DB の CHECK 制約と完全に揃える。
// migrations/000003_create_tasks.up.sql の
// `CHECK (task_status IN ('todo', 'in_progress', 'done'))` に対応。
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// ParseStatus は文字列を Status に変換する。
// 既定値以外を弾くため、handler 層のバリデーションから呼ぶ。
func ParseStatus(s string) (Status, bool) {
	switch Status(s) {
	case StatusTodo, StatusInProgress, StatusDone:
		return Status(s), true
	default:
		return "", false
	}
}
