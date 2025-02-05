package players

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
	h.Rtr.Get("/{username}", h.Stats)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := chi.URLParam(r, "leaderboard_name")
	username := r.FormValue("username")

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
	INSERT INTO players
	VALUES ($1, $2)
	`

	_, err = tx.Exec(query, l.ID, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Created plyer on leaderboard %s: %s\n", name, username)

	w.Write([]byte(response))
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "leaderboard_name")
	username := chi.URLParam(r, "username")

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
	SELECT * FROM players
	WHERE leaderboard_id = $1 AND username = $2
	`

	player := &models.Player{}
	err = tx.QueryRow(query, l.ID, username).Scan(
		&player.ID,
		&player.LeaderboardID,
		&player.Username,
		&player.MatchesWon,
		&player.MatchesDrawn,
		&player.MatchesLost,
		&player.TotalGamesWon,
		&player.TotalGamesLost,
		&player.CurrentStreak,
		&player.Elo,
		&player.CreatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"player", "W", "D", "L", "P", "GW", "GL", "Win Ratio", "Current Streak", "Elo"})

	matchesPlayed := player.MatchesWon + player.MatchesLost + player.MatchesDrawn
	winRatio := 0.
	if matchesPlayed > 0 {
		winRatio = float64(player.MatchesWon) / float64(matchesPlayed) * 100
	}

	t.AppendRow(table.Row{
		player.Username,
		matchesPlayed,
		player.MatchesWon,
		player.MatchesLost,
		player.MatchesDrawn,
		player.TotalGamesWon,
		player.TotalGamesLost,
		winRatio,
		player.CurrentStreak,
		player.Elo,
	})

	response := fmt.Sprintf("%s's Stats:\n```\n%s\n```\n", l.Name, t.Render())

	w.Write([]byte(response))
}
