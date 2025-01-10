package rest

import (
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
	s.Router.Post("/slack/events", s.record) // change back to /record after development
	s.Router.Post("/leaderboard", s.showLeaderboard)
}

func (s *Server) record(w http.ResponseWriter, r *http.Request) {

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

	commandText := r.FormValue("text")

	parts := strings.Fields(commandText)

	if len(parts) < 3 {
		http.Error(w, "Invalid command format", http.StatusBadRequest)
		return
	}

	firstPlayerName := strings.TrimPrefix(parts[firstPlayer], "@")
	secondPlayerName := strings.TrimPrefix(parts[secondPlayer], "@")

	sets := parts[2:]
	firstPlayerSetsWon, secondPlayerSetsWon := 0, 0

	for _, set := range sets {
		score := strings.Split(set, "-")

		if len(score) != 2 {
			http.Error(w, "Invalid score format", http.StatusBadRequest)
			return
		}

		firstPlayerScore, err := strconv.Atoi(score[firstPlayer])
		if err != nil {
			http.Error(w, "Invalid score for first player", http.StatusBadRequest)
			return
		}

		secondPlayerScore, err := strconv.Atoi(score[secondPlayer])
		if err != nil {
			http.Error(w, "Invalid score for second player", http.StatusBadRequest)
			return
		}

		if firstPlayerScore > secondPlayerScore {
			firstPlayerSetsWon++
		} else {
			secondPlayerSetsWon++
		}
	}

	firstPlayerStats, secondPlayerStats := getGameResult(firstPlayerSetsWon, secondPlayerSetsWon)

	err := s.doQuery(queryUpdateUser, firstPlayerName, firstPlayerStats)
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

func getGameResult(firstPlayerSetsWon, secondPlayerSetsWon int) (PlayerStats, PlayerStats) {
	switch {
	case firstPlayerSetsWon > secondPlayerSetsWon:
		return PlayerStats{1, 0, 0, firstPlayerSetsWon, secondPlayerSetsWon}, PlayerStats{0, 1, 0, secondPlayerSetsWon, firstPlayerSetsWon}
	case firstPlayerSetsWon < secondPlayerSetsWon:
		return PlayerStats{0, 1, 0, firstPlayerSetsWon, secondPlayerSetsWon}, PlayerStats{1, 0, 0, secondPlayerSetsWon, firstPlayerSetsWon}
	default:
		return PlayerStats{0, 0, 1, firstPlayerSetsWon, secondPlayerSetsWon}, PlayerStats{0, 0, 1, secondPlayerSetsWon, firstPlayerSetsWon}
	}

}
