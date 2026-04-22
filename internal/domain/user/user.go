package user

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

var (
	ErrUserNameEmpty    = errors.New("user_name is required")
	ErrUserNameTooLong  = errors.New("user_name must be 50 characters or less")
	ErrUserEmailEmpty   = errors.New("user_email is required")
	ErrUserEmailInvalid = errors.New("user_email is invalid")
)

const userNameMaxLen = 50

type User struct {
	UserID    int64
	UserName  string
	UserEmail string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

type NewUserParams struct {
	UserName  string
	UserEmail string
}

func NewUser(params NewUserParams) (*User, error) {
	name, email, err := normalizeAndValidate(params.UserName, params.UserEmail)
	if err != nil {
		return nil, err
	}
	return &User{
		UserName:  name,
		UserEmail: email,
	}, nil
}

type ReconstructParams struct {
	UserID    int64
	UserName  string
	UserEmail string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func Reconstruct(params ReconstructParams) (*User, error) {
	name, email, err := normalizeAndValidate(params.UserName, params.UserEmail)
	if err != nil {
		return nil, err
	}
	return &User{
		UserID:    params.UserID,
		UserName:  name,
		UserEmail: email,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt,
		DeletedAt: params.DeletedAt,
	}, nil
}

func normalizeAndValidate(rawName, rawEmail string) (string, string, error) {
	name := strings.TrimSpace(rawName)
	email := strings.TrimSpace(rawEmail)

	if name == "" {
		return "", "", ErrUserNameEmpty
	}
	if len([]rune(name)) > userNameMaxLen {
		return "", "", ErrUserNameTooLong
	}
	if email == "" {
		return "", "", ErrUserEmailEmpty
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", ErrUserEmailInvalid
	}
	return name, email, nil
}
