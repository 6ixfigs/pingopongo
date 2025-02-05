package rest

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/leaderboard"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/6ixfigs/pingypongy/internal/webhooks"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router chi.Router
	Config *config.Config
	db     *sql.DB
	pong   *pong.Pong
}

func NewServer() (*Server, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	db, err := db.Connect(&cfg.DBConn)
	if err != nil {
		return nil, err
	}

	return &Server{
		Router: chi.NewRouter(),
		Config: cfg,
		db:     db,
		pong:   pong.New(db),
	}, nil
}

func (s *Server) MountRoutes() {
	lh := leaderboard.NewHandler(s.db)
	lh.MountRoutes()

	wh := webhooks.NewHandler(s.db)
	wh.MountRoutes()

	s.Router.Mount("/leaderboards", lh.Rtr)
	s.Router.Mount("/leaderboards/{leaderboard_name}/webhooks", wh.Rtr)

	s.Router.Post("/leaderboards/{leaderboard_name}/players", s.createPlayer)
	s.Router.Get("/leaderboards/{leaderboard_name}/players/{username}", s.getPlayerStats)

	s.Router.Post("/leaderboards/{leaderboard_name}/matches", s.recordMatch)
}

func (s *Server) createPlayer(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaderboardName := chi.URLParam(r, "leaderboard_name")
	username := r.FormValue("username")

	err := s.pong.CreatePlayer(leaderboardName, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Created player %s on leaderboard %s", username, leaderboardName)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func (s *Server) recordMatch(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username1 := r.FormValue("player1")
	username2 := r.FormValue("player2")
	score := r.FormValue("score")

	result, err := s.pong.Record(leaderboardName, username1, username2, score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := formatMatchResult(result)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func (s *Server) getPlayerStats(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")
	username := chi.URLParam(r, "username")

	player, err := s.pong.Stats(leaderboardName, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := formatStats(player)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func formatMatchResult(result *pong.MatchResult) string {
	return fmt.Sprintf("Match recorded: (%+d) %s %d - %d %s (%+d)!",
		result.P1EloDiff,
		result.P1.Username,
		result.Score.P1,
		result.Score.P2,
		result.P2.Username,
		result.P2EloDiff,
	)
}

func formatStats(player *pong.Player) string {
	r := `Stats for %s:
	- Matches played: %d
	- Matches won: %d
	- Matches lost: %d
	- Matches drawn: %d
	- Games won: %d
	- Games lost: %d
	- Win ratio: %.2f%%
	- Current streak: %d
	- Elo: %d
	`

	matchesPlayed := player.MatchesWon + player.MatchesLost + player.MatchesDrawn
	winRatio := 0.
	if matchesPlayed > 0 {
		winRatio = float64(player.MatchesWon) / float64(matchesPlayed) * 100
	}
	return fmt.Sprintf(
		r,
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
	)
}
