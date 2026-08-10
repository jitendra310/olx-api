package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
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
	db *sql.DB
}

// the below function will construct the above struct, thats why we can can call this
// function as a constructor
func NewListingHandler(db *sql.DB) *ListingHandler {
	return &ListingHandler{
		db: db,
	}
}

// attaching method to a struct
// Now, we can say it's a method of ListingHandler
// method attacher (lh ListingHandler)
func (lh ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := lh.db.Query(
		`SELECT id, title, description, price, city, created_at
			FROM listings
			ORDER BY created_at DESC
			LIMIT 100`)
	if err != nil {
		log.Printf("query: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			log.Printf("rows.scan: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
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
	id := r.PathValue("id")

	_, err := lh.db.Exec(
		`DELETE FROM listings WHERE id=$1`, id)
	if err != nil {
		log.Printf("delete: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// result.RowsAffected() can get the affected/deleted row

	w.WriteHeader(http.StatusNoContent)
}
