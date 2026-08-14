package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jitendra310/olx-api/internal/httpx"
	"github.com/jitendra310/olx-api/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type USER struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

// to store all the dependency
type AuthHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

// the below function will construct the above struct, thats why we can can call this
// function as a constructor
func NewAuthHandler(db *sql.DB, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
	}
}

func (ah AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	log := ah.logger.With("request_id", requestId)

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode", "err", err)
		httpx.Error(w, http.StatusBadRequest, "invalid_body", httpx.CodeMalformedJSON)
		return
	}

	if err := req.Validate(); err != nil {
		var verr *ValidationError
		errors.As(err, &verr)
		httpx.ValidationError(w, http.StatusUnprocessableEntity, err.Error(), httpx.CodeValidationFailed, verr.Field)
		return
	}

	// cost : How slow you want your hashing function - 2 ^ cost 2 ^ 10 = 1024 round (function will run)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		log.Error("hashing failed", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
		return
	}

	row := ah.db.QueryRowContext(ctx,
		`INSERT INTO USERS (name, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`, req.Name, req.Email, hash)

	var u SignupResponse
	err = row.Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httpx.Error(w, http.StatusConflict, "email already taken", httpx.CodeConflict)
			return
		}

		log.Error("scaning  failed", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
		return
	}

	log.Info("new user registered", "user_id", u.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// json to struct - decode
	// struct to json - encode
	_ = json.NewEncoder(w).Encode(u)
}
