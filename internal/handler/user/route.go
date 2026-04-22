package user

import (
	"net/http"

	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

func RegisterRoutes(mux *http.ServeMux, prefix string, createUser *userusecase.CreateUser, getUser *userusecase.GetUser) {
	createH := NewCreateHandler(createUser)
	getH := NewGetHandler(getUser)

	mux.HandleFunc("POST "+prefix+"/users", createH.Handle)
	mux.HandleFunc("GET "+prefix+"/users/{userId}", getH.Handle)
}
