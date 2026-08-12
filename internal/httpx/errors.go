package httpx

import (
	"encoding/json"
	"net/http"
)

type Code string

const (
	CodeInvalidID        Code = "invalid_id"
	CodeNotFound         Code = "not_found"
	CodeInternalError    Code = "internal_error"
	CodeMalformedJSON    Code = "malformed_json"
	CodeValidationFailed Code = "validation_failed"
	CodeUnauthenticated  Code = "unauthenticated"
	CodeForbidden        Code = "forbidden"
	CodeConflict         Code = "conflict"
)

type ErrorEnvelope struct {
	Error ErrorPayload
}

type ErrorPayload struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func Error(w http.ResponseWriter, status int, message string, code Code) {
	w.Header().Set("Content-Type", "applaiction/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorEnvelope{ErrorPayload{
		Code:    code,
		Message: message,
	}})
}

func ValidationError(w http.ResponseWriter, status int, message string, code Code, field string) {
	w.Header().Set("Content-Type", "applaiction/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorEnvelope{ErrorPayload{
		Code:    code,
		Message: message,
		Field:   field,
	}})
}
