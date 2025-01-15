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

func (p *Pong) Record(channelID, commandText string) (*MatchResult, error) {

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
	p1, p2 := &Player{}, &Player{}

	result := &MatchResult{}

	args := strings.Split(commandText, " ")
	if len(args) < 3 {
		return result, fmt.Errorf("not enough arguments in command")
	}

	id1 := validateUserTag(args[0])
	id2 := validateUserTag(args[1])

	if id1 == "" || id2 == "" {
		return result, fmt.Errorf("invalid player tags %s, %s", id1, id2)
	}

	p1.UserID = strings.Split(strings.TrimPrefix(args[0], "<@"), "|")[0]
	p2.UserID = strings.Split(strings.TrimPrefix(args[1], "<@"), "|")[0]

	p1.channelID = channelID
	p2.channelID = channelID

	games := args[2:]
	err := processGameResults(games, p1, p2)

	if err != nil {
		return result, err
	}

	_, err = p.db.Exec(query, p1.UserID, p1.channelID,
		p1.matchesWon,
		p1.matchesLost,
		p1.matchesDrawn,
		p1.GamesWon,
		p1.gamesLost,
		p1.pointsWon)

	if err != nil {
		return result, err
	}

	_, err = p.db.Exec(query, p2.UserID, p2.channelID,
		p2.matchesWon,
		p2.matchesLost,
		p2.matchesDrawn,
		p2.GamesWon,
		p2.gamesLost,
		p2.pointsWon)

	if err != nil {
		return result, err
	}

	result.Winner = p1
	result.Loser = p2
	if p1.GamesWon < p2.GamesWon {
		result.Winner = p2
		result.Loser = p1
	}

	result.Games = games

	return result, nil
}

func processGameResults(games []string, p1, p2 *Player) error {
	r := `[0-9]+\-[0-9]+`
	re := regexp.MustCompile(r)

	for _, game := range games {
		score := re.FindString(game)

		if score == "" {
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
		p1.pointsWon += firstPlayerScore

		secondPlayerScore, err := strconv.Atoi(scores[1])
		if err != nil {
			return fmt.Errorf("invalid player2 score format")
		}
		p2.pointsWon += secondPlayerScore

		if firstPlayerScore > secondPlayerScore {
			p1.GamesWon++
			p2.gamesLost++
		} else if firstPlayerScore < secondPlayerScore {
			p2.GamesWon++
			p1.gamesLost++
		}
	}

	switch {
	case p1.GamesWon > p2.GamesWon:
		p1.matchesWon++
		p2.matchesLost++

	case p1.GamesWon < p2.GamesWon:
		p2.matchesWon++
		p1.matchesLost++

	default:
		p1.matchesDrawn++
		p2.matchesDrawn++
	}

	return nil
}

func validateUserTag(tag string) string {
	regex := `<@([A-Z0-9]+)\|([a-zA-Z0-9._-]+)>`
	re := regexp.MustCompile(regex)

	return re.FindString(tag)
}
