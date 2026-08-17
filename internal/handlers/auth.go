package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jitendra310/olx-api/internal/config"
	"github.com/jitendra310/olx-api/internal/httpx"
	"github.com/jitendra310/olx-api/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

var dummyHash = []byte("$2a$10$8BB6nPEzyd3NP73iMKrcXeT1WX2ENDW9KXo971qTy4sz8X8rUP0Jy")

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
	cfg    config.Config
}

// the below function will construct the above struct, thats why we can can call this
// function as a constructor
func NewAuthHandler(db *sql.DB, logger *slog.Logger, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
		cfg:    cfg,
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

func (ah AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	log := ah.logger.With("request_id", requestId)

	var req SigninRequest
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

	var u USER
	row := ah.db.QueryRowContext(ctx,
		`SELECT id, email, password FROM users WHERE email = $1`, req.Email)

	err := row.Scan(&u.ID, &u.Email, &u.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			httpx.Error(w, http.StatusUnauthorized, "email or password don't match", httpx.CodeUnauthenticated)
			return
		}
		log.Error("find user by email failed", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong ", httpx.CodeInternalError)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		log.Warn("password mismatch", "user_id", u.ID)
		httpx.Error(w, http.StatusUnauthorized, "email or password don't match", httpx.CodeUnauthenticated)
		return
	}

	tokenTTL := 24 * time.Hour
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   u.ID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}

	//gen rand key: openssl rand -base64 32 ZwQfR6OMAP9kk1WoE15U8p8oExZqfXsQESso8rrChTU=
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(ah.cfg.JwtKet))
	if err != nil {
		log.Error("jwt sign failed", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
		return
	}

	out := SigninResponse{
		Token:     signed,
		ExpiresIn: int(tokenTTL.Seconds()),
	}

	log.Info("user logged in", "user_id", u.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}
