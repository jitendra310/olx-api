package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/jitendra310/olx-api/internal/httpx"
	"github.com/jitendra310/olx-api/internal/middleware"
)

type listing struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
}

// to store all the dependency
type ListingHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

// the below function will construct the above struct, thats why we can can call this
// function as a constructor
func NewListingHandler(db *sql.DB, logger *slog.Logger) *ListingHandler {
	return &ListingHandler{
		db:     db,
		logger: logger,
	}
}

// attaching method to a struct
// Now, we can say it's a method of ListingHandler
// method attacher (lh ListingHandler)
func (lh ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	//request scoped context
	ctx := r.Context()
	// when client cancel the request but, DB continuously fetching the resources the we call it zombie query
	// to wait some time in the query we can use for test pg_sleep(20)
	rows, err := lh.db.QueryContext(ctx,
		`SELECT id, title, description, price, city, created_at
			FROM listings
			ORDER BY created_at DESC
			LIMIT 100`)
	if err != nil {
		lh.logger.Error("listing query error", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
		return
	}
	defer rows.Close()

	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			lh.logger.Error("row scan error", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
			return

		}
		lh.logger.Info("listing feached", "total", len(listings))

		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows.err: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(listings)
}

func (lh ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	id := r.PathValue("id")

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		lh.logger.Error("no userid found in context")
		httpx.Error(w, http.StatusInternalServerError, "something went wrong", httpx.CodeInternalError)
		return
	}

	// lh.logger.Debug("debug log", "listing_id", id)
	// lh.logger.Info("starting query", "listing_id", id)
	// lh.logger.Warn("warn log", "listing_id", id)

	_, err := lh.db.ExecContext(ctx,
		`DELETE FROM listings WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		lh.logger.Error("delete failed", "listing_id", id, "requestId", requestId, "err", err)
		// http.Error(w, "internal server error", http.StatusInternalServerError)
		httpx.Error(w, http.StatusInternalServerError, "somthing went wrong", httpx.CodeInternalError)
		return
	}
	// result.RowsAffected() can get the affected/deleted row

	w.WriteHeader(http.StatusNoContent)
}

func (lh ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		lh.logger.Error("no userid found in context")
		httpx.Error(w, http.StatusInternalServerError, "something went wrong", httpx.CodeInternalError)
		return
	}
	var req CreateListingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.logger.Error("failed to decode", "request_id", requestId, "err", err)
		httpx.Error(w, http.StatusBadRequest, "invalid_body", httpx.CodeMalformedJSON)
		return
	}

	if err := req.Validate(); err != nil {
		var verr *ValidationError
		errors.As(err, &verr)

		httpx.ValidationError(w, http.StatusUnprocessableEntity, err.Error(), httpx.CodeValidationFailed, verr.Field)
		return
	}

	row := lh.db.QueryRowContext(ctx, `
	INSERT INTO listings (user_id, title, description, price, city) VALUES ($1, $2, $3, $4, $5) RETURNING id, title, created_at`, userID, req.Title, req.Description, req.Price, req.City)

	// var id string
	var out CreatelistingResponse
	if err := row.Scan(&out.ID, &out.Title, &out.CreatedAt); err != nil {
		lh.logger.Error("failed to decodinserte", "request_id", requestId, "err", err)
		httpx.Error(w, http.StatusInternalServerError, "something went wrong", httpx.CodeInternalError)
		return
	}

	lh.logger.Info("listing created", "request_id", requestId, "listing_id", out.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(out)
}
