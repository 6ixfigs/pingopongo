package leaderboards

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
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
	h.Rtr.Post("/", h.Create)
	h.Rtr.Get("/{leaderboard_name}", h.Get)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")

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
	INSERT INTO leaderboards (name)
	VALUES ($1)
	`

	_, err = tx.Exec(query, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Created leaderboard: %s !\n", name)

	w.Write([]byte(response))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
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
	SELECT username, matches_won, matches_drawn, matches_lost, elo
	FROM players
	WHERE leaderboard_id = $1
	ORDER BY elo DESC
	`

	rows, err := tx.Query(query, l.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var rankings []models.Player
	for rows.Next() {
		player := models.Player{}
		err = rows.Scan(
			&player.Username,
			&player.MatchesWon,
			&player.MatchesDrawn,
			&player.MatchesLost,
			&player.Elo,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rankings = append(rankings, player)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio", "Elo"})
	for rank, player := range rankings {
		matchesPlayed := player.MatchesWon + player.MatchesDrawn + player.MatchesLost
		winRatio := 0.
		if matchesPlayed > 0 {
			winRatio = float64(player.MatchesWon) / float64(matchesPlayed) * 100
		}
		t.AppendRow(table.Row{
			rank + 1,
			player.Username,
			player.MatchesWon,
			player.MatchesDrawn,
			player.MatchesLost,
			matchesPlayed,
			fmt.Sprintf("%.2f%%", winRatio),
			player.Elo,
		})
	}

	response := fmt.Sprintf("Leaderboard %s:\n```\n%s\n```\n", l.Name, t.Render())

	w.Write([]byte(response))
}
