package handler

import (
	"encoding/json"
	"net/http"

	userhandler "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

type Deps struct {
	CreateUser *userusecase.CreateUser
}

func RegisterRoutes(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	const v1 = "/api/v1"
	userhandler.RegisterRoutes(mux, v1, d.CreateUser)
}
