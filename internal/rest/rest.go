package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router chi.Router
	pong   *pong.Pong
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
		pong:   pong.New(db),
		Config: cfg,
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/slack/events", s.parse)
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

	var result *pong.MatchResult
	var err error
	var commandResponse string
	player := &pong.Player{}

	switch request.command {
	case "/record":
		result, err = s.pong.Record(request.channelID, request.teamID, request.text)
		commandResponse = formatRecordResponse(result)

	case "/stats":
		player, err = s.pong.Stats(request.channelID, request.teamID, request.text)
		commandResponse = formatStatsResponse(player)
	default:
		http.Error(w, "Received invalid command", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(&SlackResponse{"in_channel", commandResponse})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
