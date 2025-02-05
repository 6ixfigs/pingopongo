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
	"github.com/6ixfigs/pingypongy/internal/player"
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

	ph := player.NewHandler(s.db)
	ph.MountRoutes()

	s.Router.Mount("/leaderboards", lh.Rtr)
	s.Router.Mount("/leaderboards/{leaderboard_name}/webhooks", wh.Rtr)
	s.Router.Mount("/leaderboards/{leaderboard_name}/players", ph.Rtr)

	s.Router.Post("/leaderboards/{leaderboard_name}/matches", s.recordMatch)
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
