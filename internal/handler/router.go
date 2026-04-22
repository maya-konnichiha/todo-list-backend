package handler

import (
	"net/http"

	handleruser "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"
)

// NewRouter はアプリケーションのルーティングを構築する。
// Go 1.22+ の http.ServeMux パターン機能(メソッド + パス)を使用。
func NewRouter(userHandler *handleruser.Handler) http.Handler {
	mux := http.NewServeMux()

	// 認証不要
	mux.HandleFunc("POST /users", userHandler.Create)

	return mux
}
