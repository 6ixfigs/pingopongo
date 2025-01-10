package rest

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"database/sql"
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
	s.Router.Post("/commands", s.parse)
	s.Router.Post("/leaderboard", s.showLeaderboard)
}

func (s *Server) parse(w http.ResponseWriter, r *http.Request) {
	commandName := r.FormValue("command")
	commandText := r.FormValue("text")
	commandParts := strings.Fields(commandText)

	switch strings.ToLower(commandName) {
	case "record":
		s.record(w, commandParts)
	case "leaderboard":
		s.showLeaderboard(w, r)
	default:
	}

}

func (s *Server) record(w http.ResponseWriter, commandParts []string) {

	queryUpdateUser := `
	UPDATE player_stats
	SET
		GamesWon 	= GamesWon + $2,
		GamesLost 	= GamesLost + $3,
		GamesDrawn	= GamesDrawn + $4,
		SetsWon 	= SetsWon + $5,
		SetsLost 	= SetsLost + $6
		PointsWon 	= PointsWon + $7
		PointsLost 	= PointsLost + $8
	WHERE SlackID 	= $1;
	`

	if len(commandParts) < 3 {
		http.Error(w, "Invalid command format", http.StatusBadRequest)
		return
	}

	firstPlayerName := strings.TrimPrefix(commandParts[firstPlayer], "@")
	secondPlayerName := strings.TrimPrefix(commandParts[secondPlayer], "@")

	sets := commandParts[2:]

	firstPlayerStats, secondPlayerStats, err := getGameResult(sets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = s.doQuery(queryUpdateUser, firstPlayerName, firstPlayerStats)

	if err != nil {
		http.Error(w, "Error updating player1 stats", http.StatusInternalServerError)
		return
	}

	err = s.doQuery(queryUpdateUser, secondPlayerName, secondPlayerStats)
	if err != nil {
		http.Error(w, "Error updating player2 stats", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command processed successfully!"))
}

func (s *Server) showLeaderboard(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) doQuery(query, slackID string, playerStats PlayerStats) error {

	_, err := s.db.Exec(query, slackID,
		playerStats.gamesWon,
		playerStats.gamesLost,
		playerStats.gamesDrawn,
		playerStats.setsWon,
		playerStats.setsLost)

	return err
}

func getGameResult(sets []string) (PlayerStats, PlayerStats, error) {
	firstPlayerSetsWon, secondPlayerSetsWon := 0, 0
	firstPlayerScore, secondPlayerScore := 0, 0

	for _, set := range sets {
		score := strings.Split(set, "-")

		if len(score) != 2 {
			return PlayerStats{}, PlayerStats{}, fmt.Errorf("invalid set format: %s", set)
		}

		firstPlayerScore, err := strconv.Atoi(score[firstPlayer])
		if err != nil {
			return PlayerStats{}, PlayerStats{}, fmt.Errorf("invalid player1 score format")
		}

		secondPlayerScore, err := strconv.Atoi(score[secondPlayer])
		if err != nil {
			return PlayerStats{}, PlayerStats{}, fmt.Errorf("invalid player2 score format")
		}

		if firstPlayerScore > secondPlayerScore {
			firstPlayerSetsWon++
		} else {
			secondPlayerSetsWon++
		}
	}

	switch {
	case firstPlayerSetsWon > secondPlayerSetsWon:
		return PlayerStats{1, 0, 0, firstPlayerSetsWon, secondPlayerSetsWon, firstPlayerScore, secondPlayerScore},
			PlayerStats{0, 1, 0, secondPlayerSetsWon, firstPlayerSetsWon, secondPlayerScore, firstPlayerScore},
			nil
	case firstPlayerSetsWon < secondPlayerSetsWon:
		return PlayerStats{0, 1, 0, firstPlayerSetsWon, secondPlayerSetsWon, firstPlayerScore, secondPlayerScore},
			PlayerStats{1, 0, 0, secondPlayerSetsWon, firstPlayerSetsWon, secondPlayerScore, firstPlayerScore},
			nil
	default:
		return PlayerStats{0, 0, 1, firstPlayerSetsWon, secondPlayerSetsWon, firstPlayerScore, secondPlayerScore},
			PlayerStats{0, 0, 1, secondPlayerSetsWon, firstPlayerSetsWon, secondPlayerScore, firstPlayerScore},
			nil
	}

}
