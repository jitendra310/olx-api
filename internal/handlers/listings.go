package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/jitendra310/olx-api/internal/middleware"
)

type listing struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       string    `json:"price"`
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			lh.logger.Error("row scan error", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		lh.logger.Info("listing feached", "total", len(listings))

		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows.err: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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

	// lh.logger.Debug("debug log", "listing_id", id)
	// lh.logger.Info("starting query", "listing_id", id)
	// lh.logger.Warn("warn log", "listing_id", id)

	_, err := lh.db.ExecContext(ctx,
		`DELETE FROM listing WHERE id=$1`, id)
	if err != nil {
		lh.logger.Error("delete failed", "listing_id", id, "requestId", requestId, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// result.RowsAffected() can get the affected/deleted row

	w.WriteHeader(http.StatusNoContent)
}
