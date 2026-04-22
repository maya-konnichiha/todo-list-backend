package user

import (
	"encoding/json"
	"errors"
	"net/http"

	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

type CreateHandler struct {
	usecase *userusecase.CreateUser
}

func NewCreateHandler(uc *userusecase.CreateUser) *CreateHandler {
	return &CreateHandler{usecase: uc}
}

func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	created, err := h.usecase.Execute(r.Context(), userusecase.CreateUserParams{
		UserName:  req.UserName,
		UserEmail: req.UserEmail,
	})
	if err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNameEmpty),
			errors.Is(err, userdomain.ErrUserNameTooLong),
			errors.Is(err, userdomain.ErrUserEmailEmpty),
			errors.Is(err, userdomain.ErrUserEmailInvalid):
			writeErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, userdomain.ErrUserEmailAlreadyExists):
			writeErrorJSON(w, http.StatusConflict, err.Error())
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, newCreateUserResponse(created))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
