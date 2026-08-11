package httpx

import (
	"encoding/json"
	"net/http"
)

type Code string

const (
	CodeInvalidId     Code = "invalid_id"
	CodeInternalError Code = "internal_error"
	CodeMalFormedJSON Code = "malformed_json"
)

type ErrorEnvelope struct {
	Error ErrorPayload
}

type ErrorPayload struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func Error(w http.ResponseWriter, status int, message string, code Code) {
	w.Header().Set("Content-Type", "applaiction/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorEnvelope{ErrorPayload{
		Code:    code,
		Message: message,
	}})
}
