package category

import (
	"encoding/json"
	"net/http"

	"github.com/maya-konnichiha/todo-list-backend/internal/handler/auth"
	"github.com/maya-konnichiha/todo-list-backend/internal/handler/errhandler"
	uccategory "github.com/maya-konnichiha/todo-list-backend/internal/usecase/category"
)

// CreateHandler は POST /categories を処理するハンドラ。
type CreateHandler struct {
	uc *uccategory.CreateCategoryUsecase
}

// NewCreateHandler は CreateHandler を生成する。
func NewCreateHandler(uc *uccategory.CreateCategoryUsecase) *CreateHandler {
	return &CreateHandler{uc: uc}
}

// Handle は POST /categories を処理する。
// X-User-Id ヘッダーで本人を特定する。
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromHeader(r)
	if err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "X-User-Id ヘッダーは正の整数で指定してください")
		return
	}

	var req CreateCategoryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		errhandler.BadRequest(w, "INVALID_REQUEST", "リクエストの形式が不正です")
		return
	}
	if msg := validateCreateRequest(req); msg != "" {
		errhandler.BadRequest(w, "INVALID_REQUEST", msg)
		return
	}

	if _, err := h.uc.Execute(r.Context(), uccategory.CreateInput{
		UserID:       userID,
		CategoryName: req.CategoryName,
	}); err != nil {
		errhandler.Internal(w, err)
		return
	}

	// 仕様: 201 のみ返し、ボディは持たせない。
	w.WriteHeader(http.StatusCreated)
}
