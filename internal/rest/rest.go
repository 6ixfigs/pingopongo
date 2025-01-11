package rest

import (
	"fmt"
	"log"
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
type PlayerStats struct {
	gamesWon   int
	gamesLost  int
	gamesDrawn int
	setsWon    int
	setsLost   int
	pointsWon  int
	pointsLost int
}

type RecordCommand struct {
	channelID   string
	commandText []string
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

	log.Printf("Connected to DB: %s\n", cfg.DBConn)

	return &Server{
		Router: chi.NewRouter(),
		db:     db,
		Config: cfg,
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/slack/events", s.parse)
	s.Router.Post("/leaderboard", s.showLeaderboard)
}

func (s *Server) parse(w http.ResponseWriter, r *http.Request) {
	commandName := r.FormValue("command")
	commandText := r.FormValue("text")
	commandParts := strings.Fields(commandText)

	// slackID := r.FormValue("user_id")
	channelID := r.FormValue("channel_id")
	channelID = strings.TrimPrefix(channelID, "<#")
	channelID = strings.TrimSuffix(channelID, ">")

	switch strings.ToLower(commandName) {
	case "/record":
		recordCommand := RecordCommand{channelID, commandParts}
		s.record(w, recordCommand)
	case "/leaderboard":
		s.showLeaderboard(w, r)
	default:

	}

}

func (s *Server) record(w http.ResponseWriter, recordCommand RecordCommand) {

	queryUpdateUser := `
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

	if len(recordCommand.commandText) < 3 {
		sendResponse(w, "Invalid command format", http.StatusBadRequest)
		return
	}

	firstPlayerSlackID, secondPlayerSlackID := recordCommand.commandText[0], recordCommand.commandText[1]

	firstPlayerSlackID = strings.TrimPrefix(firstPlayerSlackID, "<@")
	firstPlayerSlackID = strings.Split(strings.TrimSuffix(firstPlayerSlackID, ">"), "|")[0]

	secondPlayerSlackID = strings.TrimPrefix(secondPlayerSlackID, "<@")
	secondPlayerSlackID = strings.Split(strings.TrimSuffix(secondPlayerSlackID, ">"), "|")[0]

	sets := recordCommand.commandText[2:]

	firstPlayerStats, secondPlayerStats, err := getGameResult(sets)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = s.doQuery(queryUpdateUser, firstPlayerSlackID, recordCommand.channelID, firstPlayerStats)

	if err != nil {
		sendResponse(w, "Error updating player1 stats", http.StatusInternalServerError)
		return
	}

	err = s.doQuery(queryUpdateUser, secondPlayerSlackID, recordCommand.channelID, secondPlayerStats)
	if err != nil {
		sendResponse(w, "Error updating player2 stats", http.StatusInternalServerError)
		return
	}

	sendResponse(w, "Command processed sucessfully!", http.StatusOK)
}

func (s *Server) showLeaderboard(w http.ResponseWriter, r *http.Request) {

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

func sendResponse(w http.ResponseWriter, responseText string, status int) {
	response := map[string]string{
		"response_type": "in_channel",
		"text":          responseText,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
