package pong

import (
	"database/sql"
	"fmt"
	"math"
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
	queryInsert := `
	INSERT INTO players (slack_id, channel_id, matches_won, matches_lost, matches_drawn, games_won, games_lost, points_won, points_lost)
	VALUES ($1, $2, $3, $3, $4, $5, $6, $7, $8, $9)
	`

	queryUpdate := `
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

	exists, err := p.checkUserExists(p1)

	if err != nil {
		return result, err
	}

	if !exists {
		err = p.doQuery(queryInsert, p1)

		if err != nil {
			return result, err
		}

	}

	exists, err = p.checkUserExists(p2)

	if err != nil {
		return result, err
	}

	if !exists {
		err = p.doQuery(queryInsert, p2)

		if err != nil {
			return result, err
		}
	}

	games := args[2:]
	err = processGameResults(games, p1, p2)

	if err != nil {
		return result, err
	}

	_, err = p.db.Exec(queryUpdate, p1.UserID, p1.channelID,
		p1.matchesWon,
		p1.matchesLost,
		p1.matchesDrawn,
		p1.GamesWon,
		p1.gamesLost,
		p1.pointsWon)

	if err != nil {
		return result, err
	}

	_, err = p.db.Exec(queryUpdate, p2.UserID, p2.channelID,
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

	for _, game := range games {

		// check that score has a valid separator
		if !strings.Contains(game, "-") {
			return fmt.Errorf("game %s needs to have '-' separator", game)
		}

		scores := strings.Split(game, "-")

		// check that both players have a score
		if len(scores) != 2 {
			return fmt.Errorf("invalid game format: %s", game)
		}

		// check that both scores are numbers
		score1, err := strconv.Atoi(scores[0])
		if err != nil {
			return fmt.Errorf("player1 score needs to be a number")
		}

		score2, err := strconv.Atoi(scores[1])
		if err != nil {
			return fmt.Errorf("player2 score needs to be a number")
		}

		if score1 > 11 || score2 > 11 {
			if !(math.Abs(float64(score1-score2)) == 2) {
				return fmt.Errorf("the difference in scores of the game %s should be 2", game)
			}
		} else {
			if score1 != 11 || score2 != 11 {
				return fmt.Errorf("one of the scores in the game %s should be 11", game)
			}
		}
		p1.pointsWon += score1
		p2.pointsWon += score2

		if score1 > score2 {
			p1.GamesWon++
			p2.gamesLost++
		} else if score1 < score2 {
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

func (p *Pong) checkUserExists(player *Player) (bool, error) {

	querySelect := `
	SELECT id FROM players
	WHERE slack_id = $1
	`

	var id int
	err := p.db.QueryRow(querySelect, player.UserID).Scan(&id)

	if err != nil {
		return false, err
	}

	if err != sql.ErrNoRows {
		return true, nil
	}

	return false, nil
}

func (p *Pong) doQuery(q string, player *Player) error {

	_, err := p.db.Exec(q, player.UserID, player.channelID,
		player.matchesWon,
		player.matchesLost,
		player.matchesDrawn,
		player.GamesWon,
		player.gamesLost,
		player.pointsWon)

	return err
}

func validateUserTag(tag string) string {
	regex := `<@([A-Z0-9]+)\|([a-zA-Z0-9._-]+)>`
	re := regexp.MustCompile(regex)

	return re.FindString(tag)
}
