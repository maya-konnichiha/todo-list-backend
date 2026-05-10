package category

import (
	"net/http"

	uccategory "github.com/maya-konnichiha/todo-list-backend/internal/usecase/category"
)

// Deps は category ハンドラの依存関係。各アクションのユースケースをフィールドとして持つ。
type Deps struct {
	CreateCategoryUC *uccategory.CreateCategoryUsecase
	ListCategoriesUC *uccategory.ListCategoriesUsecase
	DeleteCategoryUC *uccategory.DeleteCategoryUsecase
}

// RegisterCategoryRoutes は category 関連のルートを mux に登録する。
func RegisterCategoryRoutes(mux *http.ServeMux, d Deps) {
	createH := NewCreateHandler(d.CreateCategoryUC)
	listH := NewListHandler(d.ListCategoriesUC)
	deleteH := NewDeleteHandler(d.DeleteCategoryUC)

	mux.HandleFunc("POST /categories", createH.Handle)
	mux.HandleFunc("GET /categories", listH.Handle)
	mux.HandleFunc("DELETE /categories/{categoryId}", deleteH.Handle)
}
