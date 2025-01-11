package rest

import (
	"fmt"
	"strconv"
	"strings"

	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/go-chi/chi/v5"
)

const firstPlayer int = 0
const secondPlayer int = 1

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
	s.Router.Post("/slack/events", s.parse)
	s.Router.Post("/leaderboard", s.leaderboard)
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

	switch strings.ToLower(request.command) {
	case "/record":
		s.record(w, request)
	case "/leaderboard":
		s.leaderboard(w, r)
	default:

	}
}

func (s *Server) record(w http.ResponseWriter, recordCommand *SlackRequest) {

	query := `
	UPDATE player_stats
	SET
		gameswon 	= gameswon + $3,
		gameslost 	= gameslost + $4,
		gamesdrawn	= gamesdrawn + $5,
		setswon		= setswon + $6,
		setslost 	= setslost + $7,
		pointswon 	= pointswon + $8,
		pointslost 	= pointslost + $9
	WHERE slackid 	= $1 AND channelid = $2;
	`

	commandParts := strings.Split(recordCommand.text, " ")
	if len(commandParts) < 3 {
		sendResponse(w, "Invalid command format")
		return
	}

	firstPlayerSlackID, secondPlayerSlackID := commandParts[0], commandParts[1]

	firstPlayerSlackID = strings.TrimPrefix(firstPlayerSlackID, "<@")
	firstPlayerSlackID = strings.Split(strings.TrimSuffix(firstPlayerSlackID, ">"), "|")[0]

	secondPlayerSlackID = strings.TrimPrefix(secondPlayerSlackID, "<@")
	secondPlayerSlackID = strings.Split(strings.TrimSuffix(secondPlayerSlackID, ">"), "|")[0]

	sets := commandParts[2:]

	firstPlayerStats, secondPlayerStats, err := getGameResult(sets)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = s.doQuery(query, firstPlayerSlackID, recordCommand.channelID, firstPlayerStats)

	if err != nil {
		sendResponse(w, "Error updating player1 stats")
		return
	}

	err = s.doQuery(query, secondPlayerSlackID, recordCommand.channelID, secondPlayerStats)
	if err != nil {
		sendResponse(w, "Error updating player2 stats")
		return
	}

	winner := firstPlayerSlackID
	if secondPlayerStats.setsWon > firstPlayerStats.setsWon {
		winner = secondPlayerSlackID
	}

	responseText := formatMatchResponse(
		firstPlayerSlackID,
		secondPlayerSlackID,
		sets,
		winner,
		firstPlayerStats.setsWon,
		secondPlayerStats.setsWon,
	)

	sendResponse(w, responseText)
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) doQuery(query, slackID, channelID string, playerStats PlayerStats) error {

	_, err := s.db.Exec(query, slackID, channelID,
		playerStats.gamesWon,
		playerStats.gamesLost,
		playerStats.gamesDrawn,
		playerStats.setsWon,
		playerStats.setsLost,
		playerStats.pointsWon,
		playerStats.pointsLost)

	return err
}

func getGameResult(sets []string) (PlayerStats, PlayerStats, error) {
	firstPlayerSetsWon, secondPlayerSetsWon := 0, 0
	totalFirstPlayerScore, totalSecondPlayerScore := 0, 0

	for _, set := range sets {
		score := strings.Split(set, "-")

		if len(score) != 2 {
			return PlayerStats{}, PlayerStats{}, fmt.Errorf("invalid set format: %s", set)
		}

		firstPlayerScore, err := strconv.Atoi(score[firstPlayer])
		if err != nil {
			return PlayerStats{}, PlayerStats{}, fmt.Errorf("invalid player1 score format")
		}
		totalFirstPlayerScore += firstPlayerScore

		secondPlayerScore, err := strconv.Atoi(score[secondPlayer])
		if err != nil {
			return PlayerStats{}, PlayerStats{}, fmt.Errorf("invalid player2 score format")
		}
		totalSecondPlayerScore += secondPlayerScore

		if firstPlayerScore > secondPlayerScore {
			firstPlayerSetsWon++
		} else if firstPlayerScore < secondPlayerScore {
			secondPlayerSetsWon++
		}
	}

	switch {
	case firstPlayerSetsWon > secondPlayerSetsWon:
		return PlayerStats{1, 0, 0, firstPlayerSetsWon, secondPlayerSetsWon, totalFirstPlayerScore, totalSecondPlayerScore},
			PlayerStats{0, 1, 0, secondPlayerSetsWon, firstPlayerSetsWon, totalSecondPlayerScore, totalFirstPlayerScore},
			nil
	case firstPlayerSetsWon < secondPlayerSetsWon:
		return PlayerStats{0, 1, 0, firstPlayerSetsWon, secondPlayerSetsWon, totalFirstPlayerScore, totalSecondPlayerScore},
			PlayerStats{1, 0, 0, secondPlayerSetsWon, firstPlayerSetsWon, totalSecondPlayerScore, totalFirstPlayerScore},
			nil
	default:
		return PlayerStats{0, 0, 1, firstPlayerSetsWon, secondPlayerSetsWon, totalFirstPlayerScore, totalSecondPlayerScore},
			PlayerStats{0, 0, 1, secondPlayerSetsWon, firstPlayerSetsWon, totalSecondPlayerScore, totalFirstPlayerScore},
			nil
	}

}

func sendResponse(w http.ResponseWriter, responseText string) {
	response, err := json.Marshal(&SlackResponse{"in_channel", responseText})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)

}

func formatMatchResponse(firstPlayer, secondPlayer string, sets []string, winner string, firstPlayerSetsWon, secondPlayerSetsWon int) string {
	var setsDetails string
	for i, set := range sets {
		setsDetails += fmt.Sprintf("- Set %d: %s\n", i+1, set)
	}

	var response string
	if firstPlayerSetsWon != secondPlayerSetsWon {
		response = fmt.Sprintf(
			"Match recorded successfully:\n<@%s> vs <@%s>\n%s🎉 Winner: <@%s> (%d-%d in sets)",
			firstPlayer,
			secondPlayer,
			setsDetails,
			winner,
			firstPlayerSetsWon,
			secondPlayerSetsWon,
		)
	} else {
		response = fmt.Sprintf(
			"Match recorder succesfully:\n<@%s> vs <@%s>\n%sDraw",
			firstPlayer,
			secondPlayer,
			setsDetails,
		)
	}

	return response
}
