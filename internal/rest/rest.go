package rest

import (
	"database/sql"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/leaderboards"
	"github.com/6ixfigs/pingypongy/internal/matches"
	"github.com/6ixfigs/pingypongy/internal/players"
	"github.com/6ixfigs/pingypongy/internal/webhooks"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Rtr chi.Mux
	db  *sql.DB
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
		Rtr: *chi.NewMux(),
		db:  db,
	}, nil
}

func (s *Server) MountRoutes() {
	lh := leaderboards.NewHandler(s.db)
	lh.MountRoutes()

	wh := webhooks.NewHandler(s.db)
	wh.MountRoutes()

	ph := players.NewHandler(s.db)
	ph.MountRoutes()

	mh := matches.NewHandler(s.db)
	mh.MountRoutes()

	s.Rtr.Mount("/leaderboards", lh.Rtr)
	s.Rtr.Mount("/leaderboards/{leaderboard_name}/webhooks", wh.Rtr)
	s.Rtr.Mount("/leaderboards/{leaderboard_name}/players", ph.Rtr)
	s.Rtr.Mount("/leaderboards/{leaderboard_name}/matches", mh.Rtr)
}
