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
	s.Router.Post("/command", s.command)
	s.Router.Post("/event", s.event)
}

func (s *Server) command(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	request := &CommandRequest{
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

	var err error
	var commandResponse string

	switch request.command {
	case "/record":
		result, err := s.pong.Record(request.channelID, request.teamID, request.text)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		commandResponse = formatRecordResponse(result)

	case "/stats":
		player, err := s.pong.Stats(request.channelID, request.teamID, request.text)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		commandResponse = formatStatsResponse(player)

	case "/leaderboard":
		leaderboard, err := s.pong.Leaderboard(request.channelID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		commandResponse = s.formatLeaderboard(leaderboard)

	default:
		http.Error(w, "Unsupported command", http.StatusBadRequest)
		return
	}

	response, err := json.Marshal(&CommandResponse{"in_channel", commandResponse})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func formatRecordResponse(res *pong.MatchResult) string {
	var gamesDetails string
	for i, g := range res.Games {
		gamesDetails += fmt.Sprintf("- Game %d: %s\n", i+1, g)
	}

	var response string
	if res.Winner.GamesWon != res.Loser.GamesWon {
		response = fmt.Sprintf(
			"Match recorded successfully:\n<@%s> vs <@%s>\n%s:trophy: Winner: <@%s> (%d-%d in games)",
			res.Winner.UserID,
			res.Loser.UserID,
			gamesDetails,
			res.Winner.UserID,
			res.Winner.GamesWon,
			res.Loser.GamesWon,
		)
	} else {
		response = fmt.Sprintf(
			"Match recorded succesfully:\n<@%s> vs <@%s>\n%sDraw",
			res.Winner.UserID,
			res.Loser.UserID,
			gamesDetails,
		)
	}

	return response
}

func formatStatsResponse(player *pong.Player) string {
	r := `Stats for <@%s>:
	- Matches played: %d
	- Matches won: %d
	- Matches lost: %d
	- Matches drawn: %d
	- Games won: %d
	- Games lost: %d
	- Points won: %d
	- Win ratio: %.2f%%
	- Current streak: %d
	`

	return fmt.Sprintf(
		r,
		player.UserID,
		player.MatchesWon+player.MatchesLost+player.MatchesDrawn,
		player.MatchesWon,
		player.MatchesLost,
		player.MatchesDrawn,
		player.GamesWon,
		player.GamesLost,
		player.PointsWon,
		float32(player.MatchesWon)/float32(player.MatchesWon+player.MatchesLost+player.MatchesDrawn)*100,
		player.CurrentStreak,
	)
}

func (s *Server) event(w http.ResponseWriter, r *http.Request) {
	var outerEvent EventRequest
	if err := json.NewDecoder(r.Body).Decode(&outerEvent); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	switch outerEvent.Type {
	case "url_verification":
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(outerEvent.Challenge))
	default:
		return
	}
}

func (s *Server) formatLeaderboard(leaderboard []pong.Player) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio"})
	for rank, player := range leaderboard {
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
	text := fmt.Sprintf(":table_tennis_paddle_and_ball: *Current Leaderboard*:\n```%s```", t.Render())

	return text
}
