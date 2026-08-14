package handlers

import (
	"net/mail"
	"strings"
	"time"
)

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (req SignupRequest) Validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return &ValidationError{Field: "title", msg: "must not be empty"}
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return &ValidationError{Field: "email", msg: "must be a valid email address"}
	}
	if len(req.Password) < 8 {
		return &ValidationError{Field: "password", msg: "must be grater then 8 charecter"}
	}

	return nil
}
