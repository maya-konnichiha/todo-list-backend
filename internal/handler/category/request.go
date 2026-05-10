package category

import "unicode/utf8"

// CreateCategoryRequest は POST /categories のリクエスト DTO。
// userId は X-User-Id ヘッダーで受け取るためボディには含めない。
type CreateCategoryRequest struct {
	CategoryName string `json:"categoryName"`
}

// validateCreateRequest はリクエストの形式を検証し、違反があれば人間向けメッセージを返す。
// 形式チェックのみを行い、ビジネスルールは usecase/repository 層で処理する。
func validateCreateRequest(req CreateCategoryRequest) string {
	if req.CategoryName == "" {
		return "categoryName は必須です"
	}
	if utf8.RuneCountInString(req.CategoryName) > 50 {
		return "categoryName は 50 文字以内にしてください"
	}
	return ""
}
