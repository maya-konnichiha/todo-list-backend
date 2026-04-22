package user

import (
	"net/http"

	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// RegisterRoutes は /users 配下のルートを mux に登録する。
//
// 引数:
//   - mux        : ルートを登録する先の ServeMux（親が用意する）
//   - prefix     : パスの先頭につける文字列（例: "/api/v1"）
//   - createUser : POST /users で呼ぶ usecase
//
// なぜ prefix を引数で受け取るのか（Gin 時代は RouterGroup に prefix が埋まっていた）:
//   - 標準の http.ServeMux には "Group" 概念が無い。グループ相当のことは
//     「パターン文字列にプレフィックスを含める」ことで実現する。
//   - サブパッケージが "/api/v1" を直接知らずに済むよう、親から文字列で受け取る。
//     これにより、プレフィックス変更時にサブパッケージを触らずに済む。
//
// Go 1.22+ の ServeMux パターン:
//
//	mux.HandleFunc("POST /users", h)   // メソッド限定
//	mux.HandleFunc("GET /users/{id}", h) // パスパラメータ ← r.PathValue("id") で取れる
//
// 以前（1.21 以下）は "POST" を区別できず、自前で `if r.Method != "POST"` を書いていた。
// 1.22 以降は ServeMux がメソッド不一致の時に 405 Method Not Allowed を自動返却してくれる。
func RegisterRoutes(mux *http.ServeMux, prefix string, createUser *userusecase.CreateUser) {
	h := NewCreateHandler(createUser)

	// "POST /api/v1/users" のようなパターンを組み立てて登録する。
	mux.HandleFunc("POST "+prefix+"/users", h.Handle)

	// TODO: mux.HandleFunc("GET " + prefix + "/users/{userId}", getHandler.Handle)
}
