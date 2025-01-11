package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Server struct {
	Router chi.Router
	db     *sql.DB
	Config *config.Config
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
		db:     db,
		Config: cfg,
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/command", s.parse)
}

func (s *Server) parse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Something went wrong", http.StatusOK)
		return
	}

	request := &SlackRequest{
		r.FormValue("team_id"),
		r.FormValue("team_domain"),
		r.FormValue("enterprise_id"),
		r.FormValue("enterprise_name"),
		r.FormValue("channel_id"),
		r.FormValue("channel_name"),
		r.FormValue("user_id"),
		r.FormValue("command"),
		r.FormValue("text"),
		r.FormValue("response_url"),
		r.FormValue("trigger_id"),
		r.FormValue("api_app_id"),
	}

	switch request.command {
	case "/leaderboard":
		s.leaderboard(w, request)
	default:
		http.Error(w, "Received invalid command", http.StatusOK)
	}
}

func (s *Server) leaderboard(w http.ResponseWriter, r *SlackRequest) {
	query := `
		SELECT user_id, matches_won, matches_drawn, matches_lost
		FROM players
		WHERE channel_id = $1
		ORDER BY matches_won DESC
		LIMIT 15
	`

	rows, err := s.db.Query(query, r.channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player

	for rows.Next() {
		var player Player
		err = rows.Scan(
			&player.userID,
			&player.matchesWon,
			&player.matchesDrawn,
			&player.matchesLost,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "Player", "W", "D", "L", "P", "Win Ratio"})
	for rank, player := range players {
		matchesPlayed := player.matchesWon + player.matchesDrawn + player.matchesLost
		t.AppendRow(table.Row{
			rank + 1,
			fmt.Sprintf("<%s>", player.userID),
			player.matchesWon,
			player.matchesDrawn,
			player.matchesLost,
			matchesPlayed,
			fmt.Sprintf("%.2f", float64(player.matchesWon)/float64(matchesPlayed)*100),
		})
	}
	leaderboard := fmt.Sprintf("```%s```", t.Render())

	response, err := json.Marshal(&SlackResponse{"in_channel", leaderboard})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
