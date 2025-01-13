package pong

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Pong struct {
	db *sql.DB
}

func New(db *sql.DB) *Pong {
	return &Pong{db}
}

func (p *Pong) Record(channelID, commandText string) (string, error) {

	query := `
	UPDATE players
	SET
		matches_won 	= matches_won + $3,
		matches_lost 	= matches_lost + $4,
		matches_drawn	= matches_drawn + $5,
		games_won		= games_won + $6,
		games_lost 		= games_lost + $7,
		points_won 		= points_won + $8,
	WHERE slack_id 		= $1 AND channel_id = $2;
	`
	var p1, p2 Player
	regex := `<@([A-Z0-9]+)\|([a-zA-Z0-9._-]+)>`
	re := regexp.MustCompile(regex)

	commandParts := strings.Split(commandText, " ")
	if len(commandParts) < 3 {
		return "", fmt.Errorf("not enough arguments in command")
	}

	id1 := re.FindString(commandParts[0])
	if id1 != "" {
		p1.userID = strings.Split(strings.TrimPrefix(commandParts[0], "<@"), "|")[0]
		log.Println(p1.userID)
		return "", fmt.Errorf("tag user %s", commandParts[0])
	}

	id2 := re.FindString(commandParts[0])
	if id2 != "" {
		p2.userID = strings.Split(strings.TrimPrefix(commandParts[1], "<@"), "|")[0]
		return "", fmt.Errorf("tag user %s", commandParts[1])
	}

	p1.channelID = channelID
	p2.channelID = channelID

	games := commandParts[2:]
	err := processMatchResult(games, &p1, &p2)

	if err != nil {
		return "", err
	}

	_, err = p.db.Exec(query, p1.userID, p1.channelID,
		p1.matchesWon,
		p1.matchesLost,
		p1.matchesDrawn,
		p1.gamesWon,
		p1.gamesLost,
		p1.pointsWon,
		p1.pointsLost)

	if err != nil {
		return "", err
	}

	_, err = p.db.Exec(query, p2.userID, p2.channelID,
		p2.matchesWon,
		p2.matchesLost,
		p2.matchesDrawn,
		p2.gamesWon,
		p2.gamesLost,
		p2.pointsWon,
		p2.gamesLost)

	if err != nil {
		return "", err
	}

	winner := p1.userID
	if p2.gamesWon > p1.gamesWon {
		winner = p2.userID
	}

	responseText := formatMatchResponse(
		p1.userID,
		p2.userID,
		games,
		winner,
		p1.gamesWon,
		p2.gamesWon,
	)

	return responseText, nil
}

func processMatchResult(games []string, p1, p2 *Player) error {
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

func formatMatchResponse(firstPlayer, secondPlayer string, sgames []string, winner string, firstPlayerGamesWon, secondPlayerGamesWon int) string {
	var gamesDetails string
	for i, set := range sgames {
		gamesDetails += fmt.Sprintf("- Set %d: %s\n", i+1, set)
	}

	var response string
	if firstPlayerGamesWon != secondPlayerGamesWon {
		response = fmt.Sprintf(
			"Match recorded successfully:\n<@%s> vs <@%s>\n%s:trophy: Winner: <@%s> (%d-%d in games)",
			firstPlayer,
			secondPlayer,
			gamesDetails,
			winner,
			firstPlayerGamesWon,
			secondPlayerGamesWon,
		)
	} else {
		response = fmt.Sprintf(
			"Match recorder succesfully:\n<@%s> vs <@%s>\n%sDraw",
			firstPlayer,
			secondPlayer,
			gamesDetails,
		)
	}

	return response
}
