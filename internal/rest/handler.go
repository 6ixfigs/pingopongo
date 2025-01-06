package auth

import (
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
)

type Handler struct {
	db *sql.DB
}

func newHandler(db *sql.DB) *Handler {
	return &Handler{db}
}

func (h *Handler) parse(w http.ResponseWriter, r *http.Request) {
}
