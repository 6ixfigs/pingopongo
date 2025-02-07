package players

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/models"
	"github.com/6ixfigs/pingypongy/internal/webhooks"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lib/pq"
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
		log.Printf("err: %v\n", err)
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	name := chi.URLParam(r, "leaderboard_name")
	username := r.FormValue("username")

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
	INSERT INTO players (leaderboard_id, username)
	VALUES ($1, $2)
	`

	_, err = tx.Exec(query, l.ID, username)
	if err != nil {
		log.Printf("err: %v\n", err)
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				http.Error(w, fmt.Sprintf("Player %s already exists on %s leaderboard.", username, name), http.StatusConflict)
				return
			}
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Created player on leaderboard %s: %s\n", name, username)

	urls, err := webhooks.All(h.db, name)
	if err == nil {
		webhooks.Notify(urls, response)
	}

	log.Print(response)

	w.Write([]byte(response))
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "leaderboard_name")
	username := chi.URLParam(r, "username")

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
		log.Printf("err: %v\n", err)
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Player %s does not exist on %s leaderboard.\n", username, name), http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
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
		fmt.Sprintf("%.2f%%", winRatio),
		player.CurrentStreak,
		player.Elo,
	})

	response := fmt.Sprintf("%s's Stats:\n```\n%s\n```\n", player.Username, t.Render())

	urls, err := webhooks.All(h.db, name)
	if err == nil {
		webhooks.Notify(urls, response)
	}

	log.Print(response)

	w.Write([]byte(response))
}
