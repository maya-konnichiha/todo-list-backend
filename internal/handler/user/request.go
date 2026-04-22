package user

type CreateUserRequest struct {
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}
