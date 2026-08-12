package handlers

import (
	"fmt"
	"strings"
	"time"
)

type CreateListingsRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	City        string `json:"city"`
}

type CreatelistingResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type ValidationError struct {
	Field string
	msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s:%s", e.Field, e.msg)
}

func (req CreateListingsRequest) Validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return &ValidationError{Field: "title", msg: "must not be empty"}
	}
	return nil
}
