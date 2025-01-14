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

	var text string
	var err error

	switch request.command {
	case "/record":
		winner, results, err = s.pong.Record(request.channelID, request.text)

	default:
		http.Error(w, "Received invalid command", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(&SlackResponse{"in_channel", text})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func formatRecordResponse(p1, p2 *Player, games []string, winner string) string {
	var gamesDetails string
	for i, g := range games {
		gamesDetails += fmt.Sprintf("- Game %d: %s\n", i+1, g)
	}

	var response string
	if p1.gamesWon != p2.gamesWon {
		response = fmt.Sprintf(
			"Match recorded successfully:\n<@%s> vs <@%s>\n%s:trophy: Winner: <@%s> (%d-%d in games)",
			p1.userID,
			p2.userID,
			gamesDetails,
			winner,
			p1.gamesWon,
			p2.gamesWon,
		)
	} else {
		response = fmt.Sprintf(
			"Match recorded succesfully:\n<@%s> vs <@%s>\n%sDraw",
			p1.userID,
			p2.userID,
			gamesDetails,
		)
	}

	return response
}
