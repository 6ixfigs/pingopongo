package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/6ixfigs/pingypongy/internal/slack"
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

	request := &slack.CommandRequest{
		TeamID:         r.FormValue("team_id"),
		TeamDomain:     r.FormValue("team_domain"),
		EnterpriseID:   r.FormValue("enterprise_id"),
		EnterpriseName: r.FormValue("enterprise_name"),
		ChannelID:      r.FormValue("channel_id"),
		ChannelName:    r.FormValue("channel_name"),
		UserID:         r.FormValue("user_id"),
		Command:        r.FormValue("command"),
		Text:           r.FormValue("text"),
		ResponseUrl:    r.FormValue("response_url"),
		TriggerID:      r.FormValue("trigger_id"),
		ApiAppID:       r.FormValue("api_app_id"),
	}

	var err error
	var pongErr *pong.PongError
	var result *pong.MatchResult
	var player *pong.Player
	var leaderboard []pong.Player
	var responseText string

	switch request.Command {
	case "/record":
		result, pongErr = s.pong.Record(request.ChannelID, request.TeamID, request.Text)
		if pongErr != nil {
			responseText = formatError(pongErr)
		} else {
			responseText = formatMatchResult(result)
		}

	case "/stats":
		player, err = s.pong.Stats(request.ChannelID, request.TeamID, request.Text)
		if err != nil {
			responseText = formatError(err.Error())
		} else {
			responseText = formatStats(player)
		}

	case "/leaderboard":
		leaderboard, err = s.pong.Leaderboard(request.ChannelID)
		if err != nil {
			responseText = formatError(err.Error())
		} else {
			responseText = formatLeaderboard(leaderboard)
		}

	default:
		http.Error(w, "Unsupported command", http.StatusBadRequest)
		return
	}

	sendPrivately := false
	if err != nil {
		sendPrivately = true
	}

	w.WriteHeader(http.StatusOK)
	sendSlackResponse(request.ResponseUrl, responseText, sendPrivately)
}

func sendSlackResponse(responseURL, responseText string, private bool) {
	responseBody := &slack.CommandResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	}

	if private {
		responseBody.ResponseType = "ephemeral"
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
	var request slack.EventRequest
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

func formatMatchResult(result *pong.MatchResult) string {
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

func formatStats(player *pong.Player) string {
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
		player.TotalGamesWon,
		player.TotalGamesLost,
		player.TotalPointsWon,
		float64(player.MatchesWon)/float64(matchesPlayed)*100,
		player.CurrentStreak,
		player.Elo,
	)
}

func formatLeaderboard(leaderboard []pong.Player) string {
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

func formatError(e error) string {

	switch e.Code {
	case pong.UserError:
		return ":exclamation:" + e.Info
	case pong.InternalError:
		return "Something went wrong!"

	default:
		return "Something went wrong!"
	}
}
