package user

import (
	"net/http"

	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

func RegisterRoutes(mux *http.ServeMux, prefix string, createUser *userusecase.CreateUser) {
	h := NewCreateHandler(createUser)
	mux.HandleFunc("POST "+prefix+"/users", h.Handle)
}
