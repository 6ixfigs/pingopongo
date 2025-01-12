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

	switch request.command {
	case "/record":
		s.record(w, request)
	default:

	}
}

func (s *Server) record(w http.ResponseWriter, r *SlackRequest) {

	query := `
	UPDATE players
	SET
		games_won 	= games_won + $3,
		games_lost 	= games_lost + $4,
		games_drawn	= games_drawn + $5,
		sets_won	= sets_won + $6,
		sets_lost 	= sets_lost + $7,
		points_won 	= points_won + $8,
		points_lost = points_lost + $9
	WHERE slack_id 	= $1 AND channel_id = $2;
	`

	commandParts := strings.Split(r.text, " ")
	if len(commandParts) < 3 {
		sendResponse(w, "Invalid command format")
		return
	}

	var firstPlayer, secondPlayer Player

	firstPlayer.userID = strings.Split(strings.TrimPrefix(commandParts[0], "<@"), "|")[0]
	secondPlayer.userID = strings.Split(strings.TrimPrefix(commandParts[1], "<@"), "|")[0]

	firstPlayer.channelID = r.channelID
	secondPlayer.channelID = r.channelID

	games := commandParts[2:]

	err := getMatchResult(games, &firstPlayer, &secondPlayer)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Println(firstPlayer.userID)
	log.Println(firstPlayer.channelID)
	log.Println(firstPlayer.matchesWon)
	log.Println(firstPlayer.matchesLost)
	log.Println(firstPlayer.matchesDrawn)
	log.Println(firstPlayer.gamesWon)
	log.Println(firstPlayer.gamesLost)
	log.Println(firstPlayer.pointsWon)
	log.Println(firstPlayer.pointsLost)

	_, err = s.db.Exec(query, firstPlayer.userID, firstPlayer.channelID,
		firstPlayer.matchesWon,
		firstPlayer.matchesLost,
		firstPlayer.matchesDrawn,
		firstPlayer.gamesWon,
		firstPlayer.gamesLost,
		firstPlayer.pointsWon,
		firstPlayer.pointsLost)

	if err != nil {
		sendResponse(w, "Error updating player1 stats")
		return
	}

	_, err = s.db.Exec(query, secondPlayer.userID, secondPlayer.channelID,
		secondPlayer.matchesWon,
		secondPlayer.matchesLost,
		secondPlayer.matchesDrawn,
		secondPlayer.gamesWon,
		secondPlayer.gamesLost,
		secondPlayer.pointsWon,
		secondPlayer.gamesLost)

	if err != nil {
		sendResponse(w, "Error updating player2 stats")
		return
	}

	winner := firstPlayer.userID
	if secondPlayer.gamesWon > firstPlayer.gamesWon {
		winner = secondPlayer.userID
	}

	responseText := formatMatchResponse(
		firstPlayer.userID,
		secondPlayer.userID,
		games,
		winner,
		firstPlayer.gamesWon,
		secondPlayer.gamesWon,
	)

	sendResponse(w, responseText)
}

func getMatchResult(games []string, p1, p2 *Player) error {
	firstPlayerGamesWon, secondPlayerGamesWon := 0, 0
	totalFirstPlayerScore, totalSecondPlayerScore := 0, 0

	for _, game := range games {
		score := strings.Split(game, "-")

		if len(score) != 2 {
			return fmt.Errorf("invalid set format: %s", game)
		}

		firstPlayerScore, err := strconv.Atoi(score[0])
		if err != nil {
			return fmt.Errorf("invalid player1 score format")
		}
		totalFirstPlayerScore += firstPlayerScore

		secondPlayerScore, err := strconv.Atoi(score[1])
		if err != nil {
			return fmt.Errorf("invalid player2 score format")
		}
		totalSecondPlayerScore += secondPlayerScore

		if firstPlayerScore > secondPlayerScore {
			firstPlayerGamesWon++
		} else if firstPlayerScore < secondPlayerScore {
			secondPlayerGamesWon++
		}
	}

	p1.gamesWon = firstPlayerGamesWon
	p1.gamesLost = secondPlayerGamesWon

	p1.pointsWon = totalFirstPlayerScore
	p1.pointsLost = totalSecondPlayerScore

	p2.gamesWon = secondPlayerGamesWon
	p2.gamesLost = firstPlayerGamesWon

	p2.pointsWon = totalSecondPlayerScore
	p2.pointsLost = totalFirstPlayerScore

	switch {
	case firstPlayerGamesWon > secondPlayerGamesWon:
		p1.matchesWon++
		p2.matchesLost++

	case firstPlayerGamesWon < secondPlayerGamesWon:
		p2.matchesWon++
		p1.matchesLost++

	default:
		p1.matchesDrawn++
		p2.matchesDrawn++
	}

	return nil
}

func sendResponse(w http.ResponseWriter, responseText string) {
	response, err := json.Marshal(&SlackResponse{"in_channel", responseText})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)

}

func formatMatchResponse(firstPlayer, secondPlayer string, sgames []string, winner string, firstPlayerGamesWon, secondPlayerGamesWon int) string {
	var setsDetails string
	for i, set := range sgames {
		setsDetails += fmt.Sprintf("- Set %d: %s\n", i+1, set)
	}

	var response string
	if firstPlayerGamesWon != secondPlayerGamesWon {
		response = fmt.Sprintf(
			"Match recorded successfully:\n<@%s> vs <@%s>\n%s🎉 Winner: <@%s> (%d-%d in sets)",
			firstPlayer,
			secondPlayer,
			setsDetails,
			winner,
			firstPlayerGamesWon,
			secondPlayerGamesWon,
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
