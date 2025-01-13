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
	} else {
		return "", fmt.Errorf("tag user %s", commandParts[0])
	}

	id2 := re.FindString(commandParts[1])
	if id2 != "" {
		p2.userID = strings.Split(strings.TrimPrefix(commandParts[1], "<@"), "|")[0]
	} else {
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
		p1.pointsWon)

	if err != nil {
		return "", err
	}

	_, err = p.db.Exec(query, p2.userID, p2.channelID,
		p2.matchesWon,
		p2.matchesLost,
		p2.matchesDrawn,
		p2.gamesWon,
		p2.gamesLost,
		p2.pointsWon)

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
	r := `[0-9]+\-[0-9]+`
	re := regexp.MustCompile(r)

	totalGames1, totalGames2 := 0, 0
	totalScore1, totalScore2 := 0, 0

	for _, game := range games {
		s := re.FindString(game)

		if s == "" {
			return fmt.Errorf("invalid game format %s", game)
		}

		score := strings.Split(game, "-")

		if len(score) != 2 {
			return fmt.Errorf("invalid set format: %s", game)
		}

		firstPlayerScore, err := strconv.Atoi(score[0])
		if err != nil {
			return fmt.Errorf("invalid player1 score format")
		}
		totalScore1 += firstPlayerScore

		secondPlayerScore, err := strconv.Atoi(score[1])
		if err != nil {
			return fmt.Errorf("invalid player2 score format")
		}
		totalScore2 += secondPlayerScore

		if firstPlayerScore > secondPlayerScore {
			totalGames1++
		} else if firstPlayerScore < secondPlayerScore {
			totalGames2++
		}
	}

	p1.gamesWon = totalGames1
	p1.gamesLost = totalGames2

	p1.pointsWon = totalScore1

	p2.gamesWon = totalGames2
	p2.gamesLost = totalGames1

	p2.pointsWon = totalScore2

	switch {
	case totalGames1 > totalGames2:
		p1.matchesWon++
		p2.matchesLost++

	case totalGames1 < totalGames2:
		p2.matchesWon++
		p1.matchesLost++

	default:
		p1.matchesDrawn++
		p2.matchesDrawn++
	}

	log.Println("OK")

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
