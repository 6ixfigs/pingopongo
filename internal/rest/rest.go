package rest

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/leaderboards"
	"github.com/6ixfigs/pingypongy/internal/matches"
	"github.com/6ixfigs/pingypongy/internal/players"
	"github.com/6ixfigs/pingypongy/internal/webhooks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Rtr *chi.Mux
	Cfg *config.Config
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
		Rtr: chi.NewRouter(),
		Cfg: cfg,
		db:  db,
	}, nil
}

func (s *Server) MountRoutes() {
	s.Rtr.Use(requestLogger)
	s.Rtr.Use(middleware.Recoverer)
	s.Rtr.Use(middleware.CleanPath)
	s.Rtr.Use(middleware.RedirectSlashes)
	s.Rtr.Use(middleware.AllowContentType("application/x-www-form-urlencoded"))
	s.Rtr.Use(middleware.Heartbeat("/ping"))

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

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
