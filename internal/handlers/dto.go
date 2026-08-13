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
	if len(req.Title) > 200 {
		return &ValidationError{Field: "title", msg: "must be at most 200 charaters"}
	}
	if strings.TrimSpace(req.Description) == "" {
		return &ValidationError{Field: "description", msg: "must not be empty"}
	}
	if len(req.Description) > 5000 {
		return &ValidationError{Field: "description", msg: "must be at most 5000 charaters"}
	}
	if req.Price <= 0 {
		return &ValidationError{Field: "price", msg: "must be grater then zero"}
	}
	if strings.TrimSpace(req.City) == "" {
		return &ValidationError{Field: "city", msg: "must not be empty"}
	}

	return nil
}
