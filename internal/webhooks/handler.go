package webhooks

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("err: %v\n", err)
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	name := chi.URLParam(r, "leaderboard_name")
	url := r.FormValue("url")

	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("err: %v\n", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
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
		log.Printf("err: %v\n", err)
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Leaderboard %s does not exist.\n", name), http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	query = `
	INSERT INTO webhooks (leaderboard_id, url)
	VALUES ($1, $2)
	`
	_, err = tx.Exec(query, l.ID, url)
	if err != nil {
		log.Printf("err: %v\n", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Registered new webhook on leaderboard %s: %s\n", name, url)

	log.Print(response)

	w.Write([]byte(response))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "leaderboard_name")

	webhooks, err := All(h.db, name)
	if err != nil {
		log.Printf("err: %v\n", err)
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Leaderboard %s does not exist.\n", name), http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	var response string
	if len(webhooks) > 0 {
		response = strings.Join(webhooks, "\n") + "\n"
	} else {
		response = "No webhooks registered.\n"
	}

	w.Write([]byte(response))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "leaderboard_name")

	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("err: %v\n", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
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
		log.Printf("err: %v\n", err)
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Leaderboard %s does not exist.\n", name), http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	query = `
	DELETE FROM webhooks
	WHERE leaderboard_id = $1
	`
	_, err = tx.Exec(query, l.ID)
	if err != nil {
		log.Printf("err: %v\n", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Delete all webhooks on leaderboard: %s\n", name)

	log.Print(response)

	w.Write([]byte(response))
}

func Notify(webhooks []string, message string) {
	type payload struct {
		Text string `json:"text"`
	}

	p, err := json.Marshal(payload{message})
	if err != nil {
		log.Printf("err: %v\n", err)
	}

	for _, webhook := range webhooks {
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(p))
		if err != nil {
			log.Printf("err: %v\n", err)
			continue
		}

		if _, err := http.DefaultClient.Do(req); err != nil {
			log.Printf("Unable to notify webhook %s", webhook)
			continue
		}

		log.Printf("Notified %s", webhook)
	}
}

func All(db *sql.DB, leaderboard string) ([]string, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	query := `
	SELECT id, name FROM leaderboards
	WHERE name = $1
	`
	l := &models.Leaderboard{}
	err = tx.QueryRow(query, leaderboard).Scan(
		&l.ID,
		&l.Name,
	)
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil, err
	}

	query = `
	SELECT url FROM webhooks
	WHERE leaderboard_id = $1
	`

	rows, err := tx.Query(query, l.ID)
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var webhooks []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Printf("err: %v\n", err)
			return nil, err
		}
		webhooks = append(webhooks, url)
	}

	if err := rows.Err(); err != nil {
		log.Printf("err: %v\n", err)
		return nil, err
	}

	return webhooks, nil
}
