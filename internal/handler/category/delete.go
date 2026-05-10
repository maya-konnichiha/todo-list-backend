package category

import (
	"errors"
	"net/http"
	"strconv"

	domaincategory "github.com/maya-konnichiha/todo-list-backend/internal/domain/category"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uccategory "github.com/maya-konnichiha/todo-list-backend/internal/usecase/category"
)

// DeleteHandler は DELETE /categories/{categoryId} を処理するハンドラ。
type DeleteHandler struct {
	uc *uccategory.DeleteCategoryUsecase
}

// NewDeleteHandler は DeleteHandler を生成する。
func NewDeleteHandler(uc *uccategory.DeleteCategoryUsecase) *DeleteHandler {
	return &DeleteHandler{uc: uc}
}

// Handle は DELETE /categories/{categoryId} を処理する。
// 他人のカテゴリへの削除も「存在しない」として 404 を返す(権限漏洩防止)。
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromHeader(r)
	if err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "X-User-Id ヘッダーは正の整数で指定してください")
		return
	}

	categoryID, err := strconv.ParseInt(r.PathValue("categoryId"), 10, 64)
	if err != nil || categoryID <= 0 {
		errhandler.BadRequest(w, "INVALID_REQUEST", "categoryId は正の整数で指定してください")
		return
	}

	if err := h.uc.Execute(r.Context(), uccategory.DeleteInput{
		UserID:     userID,
		CategoryID: categoryID,
	}); err != nil {
		if errors.Is(err, domaincategory.ErrNotFound) {
			errhandler.NotFound(w, "CATEGORY_NOT_FOUND", "カテゴリが見つかりません")
			return
		}
		errhandler.Internal(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
