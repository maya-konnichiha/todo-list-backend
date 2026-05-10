package category

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uccategory "github.com/maya-konnichiha/todo-list-backend/internal/usecase/category"
)

// ListHandler は GET /categories を処理するハンドラ。
type ListHandler struct {
	uc *uccategory.ListCategoriesUsecase
}

// NewListHandler は ListHandler を生成する。
func NewListHandler(uc *uccategory.ListCategoriesUsecase) *ListHandler {
	return &ListHandler{uc: uc}
}

// Handle は GET /categories を処理する。
// X-User-Id ヘッダーで本人を特定し、本人が持つカテゴリ一覧を返す。
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromHeader(r)
	if err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "X-User-Id ヘッダーは正の整数で指定してください")
		return
	}

	categories, err := h.uc.Execute(r.Context(), uccategory.ListInput{UserID: userID})
	if err != nil {
		errhandler.Internal(w, err)
		return
	}

	resp := ToListResponse(categories)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode list categories response", slog.Any("error", err))
	}
}
