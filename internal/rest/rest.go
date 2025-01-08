package rest

import (
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

type GameResults struct {
	firstPlayerGamesWon    int
	firstPlayerGamesLost   int
	firstPlayerGamesDrawn  int
	secondPlayerGamesWon   int
	secondPlayerGamesLost  int
	secondPlayerGamesDrawn int
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
	s.Router.Post("/command", s.parse)
}

func (s *Server) parse(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse data from Slack Client.", http.StatusBadRequest)
		return
	}

	user := r.FormValue("user")
	command := r.FormValue("command")
	text := r.FormValue("text")

	if strings.Compare("record", strings.ToLower(command)) == 0 {
		s.record(user, text)
	}

}

func (s *Server) record(username, commandText string) {

	// if user already in db, do nothing

	queryInsertUpdateUser := `
		INSERT INTO player_stats (username, games_won, games_lost, games_drawn, sets_won, sets_lost)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (username)
		DO UPDATE SET
			games_won = EXCLUDED.games_won + $2,
			games_lost = EXCLUDED.games_lost + $3,
			games_drawn = EXCLUDED.games_drawn + $4,
			sets_won = sets_won + EXCLUDED.sets_won,
			sets_lost = sets_lost + EXCLUDED.sets_lost;
	`

	firstPlayerName := commandText[1]
	secondPlayerName := commandText[2]
	sets := strings.Split(commandText[3:], " ")

	firstPlayerSetsWon, firstPlayerSetsLost, secondPlayerSetsWon, secondPlayerSetsLost := 0, 0, 0, 0
	for setNo := 0; setNo < len(sets); setNo++ {
		score := strings.Split(sets[setNo], "-")

		if score[firstPlayer] > score[secondPlayer] {
			firstPlayerSetsWon++
			secondPlayerSetsLost++
		} else {
			firstPlayerSetsLost++
			secondPlayerSetsWon++
		}
	}

	gameResult := getGameResult(firstPlayerSetsWon, secondPlayerSetsWon)

	_, err := s.db.Exec(queryInsertUpdateUser, username,
		gameResult.firstPlayerGamesWon,
		gameResult.firstPlayerGamesLost,
		gameResult.firstPlayerGamesDrawn,
		firstPlayerSetsWon,
		firstPlayerSetsLost)

	if err != nil {
		return
	}
	_, err = s.db.Exec(queryInsertUpdateUser, username,
		gameResult.secondPlayerGamesWon,
		gameResult.secondPlayerGamesLost,
		gameResult.secondPlayerGamesDrawn,
		secondPlayerSetsWon,
		secondPlayerSetsLost)

}

func getGameResult(firstPlayerSets, secondPlayerSets int) GameResults {
	switch {
	case firstPlayerSets > secondPlayerSets:
		return GameResults{1, 0, 0, 0, 1, 0}
	case firstPlayerSets < secondPlayerSets:
		return GameResults{0, 1, 0, 1, 0, 0}
	default:
		return GameResults{0, 0, 1, 0, 0, 1}
	}

}
