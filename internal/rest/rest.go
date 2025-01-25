package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/6ixfigs/pingypongy/internal/repository"
	"github.com/6ixfigs/pingypongy/internal/types"
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
		pong:   pong.New(repository.NewSQLRepository(db)),
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

	var responseText string

	switch request.command {
	case "/record":
		result, err := s.pong.Record(request.channelID, request.teamID, request.text)
		if err != nil {
			responseText = formatError(err.Error())
		} else {
			responseText = formatMatchResult(result)
		}

	case "/stats":
		player, err := s.pong.Stats(request.channelID, request.teamID, request.text)
		if err != nil {
			responseText = formatError(err.Error())
		} else {
			responseText = formatStats(player)
		}

	case "/leaderboard":
		leaderboard, err := s.pong.Leaderboard(request.channelID)
		if err != nil {
			responseText = formatError(err.Error())
		} else {
			responseText = formatLeaderboard(leaderboard)
		}

	default:
		// code won't reach this point because Slack will automatically return an error message for invalid slash commands
		return
	}

	w.WriteHeader(http.StatusOK)
	sendSlackResponse(request.responseUrl, responseText)
}

func sendSlackResponse(responseURL, responseText string) {
	responseBody := &CommandResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		fmt.Printf("Failed to marshal response: %v\n", err)
		return
	}

	resp, err := http.Post(responseURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Failed to send response to Slack: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non-OK response from Slack: %v\n", resp.Status)
	}
}

func (s *Server) event(w http.ResponseWriter, r *http.Request) {
	var request EventRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	if request.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(request.Challenge))
		return
	}

	innerEvent := request.Event
	var err error

	switch innerEvent["type"] {
	case "channel_id_changed":
		err = s.pong.UpdateChannelID(innerEvent["old_channel_id"], innerEvent["new_channel_id"])
	default:
		err = fmt.Errorf("unrecognized event")
	}

	fmt.Printf("err: %v\n", err)
}

func formatMatchResult(result *types.MatchResult) string {
	players := fmt.Sprintf("<@%s> vs <@%s>", result.P1.UserID, result.P2.UserID)

	var gameResults string
	for i, g := range result.Games {
		gameResults += fmt.Sprintf("- Game %d: %d-%d\n", i+1, g.P1PointsWon, g.P2PointsWon)
	}

	var conclusion string
	if result.IsDraw {
		conclusion = "Draw!"
	} else {
		conclusion = fmt.Sprintf(":trophy: Winner: <@%s> %d-%d",
			result.Winner.UserID,
			result.P1GamesWon,
			result.P2GamesWon,
		)
	}

	response := fmt.Sprintf("Match recorded:\n%s\n%s\n%s", players, gameResults, conclusion)
	return response
}

func formatStats(player *types.Player) string {
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
	- Elo: %d
	`

	matchesPlayed := player.MatchesWon + player.MatchesLost + player.MatchesDrawn
	return fmt.Sprintf(
		r,
		player.UserID,
		matchesPlayed,
		player.MatchesWon,
		player.MatchesLost,
		player.MatchesDrawn,
		player.GamesWon,
		player.GamesLost,
		player.PointsWon,
		float64(player.MatchesWon)/float64(matchesPlayed)*100,
		player.CurrentStreak,
		player.Elo,
	)
}

func formatLeaderboard(leaderboard []types.Player) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio", "Elo"})
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
			player.Elo,
		})
	}
	text := fmt.Sprintf(":table_tennis_paddle_and_ball: *Current Leaderboard*:\n```%s```", t.Render())

	return text
}

func formatError(err string) string {
	return ":exclamation: " + err
}
