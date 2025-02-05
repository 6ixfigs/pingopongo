package webhooks

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/6ixfigs/pingypongy/internal/models"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Rtr chi.Router
	db  *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		Rtr: chi.NewRouter(),
		db:  db,
	}
}

func (h *Handler) MountRoutes() {
	h.Rtr.Post("/", h.Register)
	h.Rtr.Get("/", h.List)
	h.Rtr.Delete("/", h.Delete)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := chi.URLParam(r, "leaderboard_name")
	url := r.FormValue("url")

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	query := `
	SELECT id, name FROM leaderboards
	WHERE name = $1
	`

	l := &models.Leaderboard{}
	err = tx.QueryRow(query, name).Scan(
		&l.ID,
		&l.Name,
	)

	query = `
	INSERT INTO webhooks (leaderboard_id, url)
	VALUES ($1, $2)
	`

	_, err = tx.Exec(query, l.ID, url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Registered new webhook on leaderboard %s: %s", name, url)

	w.Write([]byte(response))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "leaderboard_name")

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	query := `
	SELECT id, name FROM leaderboards
	WHERE name = $1
	`

	l := &models.Leaderboard{}
	err = tx.QueryRow(query, name).Scan(
		&l.ID,
		&l.Name,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query = `
	SELECT url FROM webhooks
	WHERE leaderboard_id = $1
	`

	rows, err := tx.Query(query, l.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var webhooks []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		webhooks = append(webhooks, url)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := strings.Join(webhooks, "\n")

	w.Write([]byte(response))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "leaderboard_name")

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	query := `
	SELECT id, name FROM leaderboards
	WHERE name = $1
	`

	l := &models.Leaderboard{}
	err = tx.QueryRow(query, name).Scan(
		&l.ID,
		&l.Name,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	query = `
	DELETE FROM webhooks
	WHERE leaderboard_id = $1
	`

	_, err = tx.Exec(query, l.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Delete all webhooks on leaderboard: %s", name)

	w.Write([]byte(response))
}

func Notify(webhooks []string, message string) {
	payload := []byte(fmt.Sprintf(`{ "text": "%s"}`, message))
	for _, webhook := range webhooks {
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(payload))
		if err != nil {
			continue
		}

		http.DefaultClient.Do(req)
	}
}
