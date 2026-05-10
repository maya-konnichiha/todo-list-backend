package task

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	domaintask "github.com/maya-konnichiha/todo-list-backend/internal/domain/task"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uctask "github.com/maya-konnichiha/todo-list-backend/internal/usecase/task"
)

// ListHandler は GET /tasks?taskStatus=&categoryId= を処理するハンドラ。
type ListHandler struct {
	uc *uctask.ListTasksUsecase
}

// NewListHandler は ListHandler を生成する。
func NewListHandler(uc *uctask.ListTasksUsecase) *ListHandler {
	return &ListHandler{uc: uc}
}

// Handle はタスク一覧を取得する。クエリパラメータでフィルタする。
//   - taskStatus: todo / in_progress / done
//   - categoryId: 正の整数
//
// どちらも省略可。両方無ければ本人の全アクティブタスクを返す。
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromHeader(r)
	if err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "X-User-Id ヘッダーは正の整数で指定してください")
		return
	}

	var statusFilter *domaintask.Status
	if raw := r.URL.Query().Get("taskStatus"); raw != "" {
		s, ok := domaintask.ParseStatus(raw)
		if !ok {
			errhandler.BadRequest(w, "INVALID_REQUEST", "taskStatus は todo / in_progress / done のいずれかで指定してください")
			return
		}
		statusFilter = &s
	}

	var categoryFilter *int64
	if raw := r.URL.Query().Get("categoryId"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			errhandler.BadRequest(w, "INVALID_REQUEST", "categoryId は正の整数で指定してください")
			return
		}
		categoryFilter = &id
	}

	tasks, err := h.uc.Execute(r.Context(), uctask.ListInput{
		UserID:     userID,
		TaskStatus: statusFilter,
		CategoryID: categoryFilter,
	})
	if err != nil {
		errhandler.Internal(w, err)
		return
	}

	resp := ToListResponse(tasks)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode list tasks response", slog.Any("error", err))
	}
}
