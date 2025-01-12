package pong

import (
	"database/sql"
	"fmt"
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
		games_won 	= games_won + $3,
		games_lost 	= games_lost + $4,
		games_drawn	= games_drawn + $5,
		sets_won	= sets_won + $6,
		sets_lost 	= sets_lost + $7,
		points_won 	= points_won + $8,
		points_lost = points_lost + $9
	WHERE slack_id 	= $1 AND channel_id = $2;
	`

	commandParts := strings.Split(commandText, " ")
	if len(commandParts) < 3 {
		return "", fmt.Errorf("invalid command format")
	}

	var firstPlayer, secondPlayer Player

	firstPlayer.userID = strings.Split(strings.TrimPrefix(commandParts[0], "<@"), "|")[0]
	secondPlayer.userID = strings.Split(strings.TrimPrefix(commandParts[1], "<@"), "|")[0]

	firstPlayer.channelID = channelID
	secondPlayer.channelID = channelID

	games := commandParts[2:]

	err := getMatchResult(games, &firstPlayer, &secondPlayer)

	if err != nil {
		return "", err
	}

	_, err = p.db.Exec(query, firstPlayer.userID, firstPlayer.channelID,
		firstPlayer.matchesWon,
		firstPlayer.matchesLost,
		firstPlayer.matchesDrawn,
		firstPlayer.gamesWon,
		firstPlayer.gamesLost,
		firstPlayer.pointsWon,
		firstPlayer.pointsLost)

	if err != nil {
		return "", err
	}

	_, err = p.db.Exec(query, secondPlayer.userID, secondPlayer.channelID,
		secondPlayer.matchesWon,
		secondPlayer.matchesLost,
		secondPlayer.matchesDrawn,
		secondPlayer.gamesWon,
		secondPlayer.gamesLost,
		secondPlayer.pointsWon,
		secondPlayer.gamesLost)

	if err != nil {
		return "", err
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

	return responseText, nil
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

func formatMatchResponse(firstPlayer, secondPlayer string, sgames []string, winner string, firstPlayerGamesWon, secondPlayerGamesWon int) string {
	var setsDetails string
	for i, set := range sgames {
		setsDetails += fmt.Sprintf("- Set %d: %s\n", i+1, set)
	}

	var response string
	if firstPlayerGamesWon != secondPlayerGamesWon {
		response = fmt.Sprintf(
			"Match recorded successfully:\n<@%s> vs <@%s>\n%s:trophy: Winner: <@%s> (%d-%d in sets)",
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
