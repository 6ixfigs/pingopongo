package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Server struct {
	Router chi.Router
	Config *config.Config
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
		pong:   pong.New(db),
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/command", s.parse)
}

func (s *Server) parse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
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

	var text string

	switch request.command {
	case "/leaderboard":
		players, err := s.pong.Leaderboard(request.channelID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		text = s.formatLeaderboardResponse(players)
	default:
		http.Error(w, "Unsupported command", http.StatusBadRequest)
		return
	}

	response, err := json.Marshal(&SlackResponse{"in_channel", text})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (s *Server) formatLeaderboardResponse(players []pong.Player) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio"})
	for rank, player := range players {
		matchesPlayed := player.MatchesWon + player.MatchesDrawn + player.MatchesLost
		t.AppendRow(table.Row{
			rank + 1,
			player.FullName,
			player.MatchesWon,
			player.MatchesDrawn,
			player.MatchesLost,
			matchesPlayed,
			fmt.Sprintf("%.2f", float64(player.MatchesWon)/float64(matchesPlayed)*100),
		})
	}
	leaderboard := fmt.Sprintf(":table_tennis_paddle_and_ball: *Current Leaderboard*:\n```%s```", t.Render())

	return leaderboard
}
