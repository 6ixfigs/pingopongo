package pong

import (
	"database/sql"
	"fmt"
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
		points_won 		= points_won + $8
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

	responseText := formatMatchResponse(p1, p2, games, winner)

	return responseText, nil
}

func processMatchResult(games []string, p1, p2 *Player) error {
	r := `[0-9]+\-[0-9]+`
	re := regexp.MustCompile(r)

	games1, games2 := 0, 0
	score1, score2 := 0, 0

	for _, game := range games {
		s := re.FindString(game)

		if s == "" {
			return fmt.Errorf("invalid game format %s", game)
		}

		scores := strings.Split(game, "-")

		if len(scores) != 2 {
			return fmt.Errorf("invalid game format: %s", game)
		}

		firstPlayerScore, err := strconv.Atoi(scores[0])
		if err != nil {
			return fmt.Errorf("invalid player1 score format")
		}
		score1 += firstPlayerScore

		secondPlayerScore, err := strconv.Atoi(scores[1])
		if err != nil {
			return fmt.Errorf("invalid player2 score format")
		}
		score2 += secondPlayerScore

		if firstPlayerScore > secondPlayerScore {
			games1++
		} else if firstPlayerScore < secondPlayerScore {
			games2++
		}
	}

	p1.gamesWon = games1
	p1.gamesLost = games2

	p1.pointsWon = score1

	p2.gamesWon = games2
	p2.gamesLost = games1

	p2.pointsWon = score2

	switch {
	case games1 > games2:
		p1.matchesWon++
		p2.matchesLost++

	case games1 < games2:
		p2.matchesWon++
		p1.matchesLost++

	default:
		p1.matchesDrawn++
		p2.matchesDrawn++
	}

	return nil
}

func formatMatchResponse(p1, p2 Player, games []string, winner string) string {
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
			"Match recorder succesfully:\n<@%s> vs <@%s>\n%sDraw",
			p1.userID,
			p2.userID,
			gamesDetails,
		)
	}

	return response
}
