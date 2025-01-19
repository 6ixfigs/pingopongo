package pong

import (
	"database/sql"
	"fmt"
	"log"
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

func (p *Pong) Record(channelID, teamID, commandText string) (*MatchResult, error) {
	queryInsert := `
	INSERT INTO players (user_id, channel_id, team_id, matches_won, matches_lost, matches_drawn, games_won, games_lost, points_won, current_streak)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`

	queryUpdate := `
	UPDATE players
	SET
		matches_won 	= matches_won + $4,
		matches_lost 	= matches_lost + $5,
		matches_drawn	= matches_drawn + $6,
		games_won		= games_won + $7,
		games_lost 		= games_lost + $8,
		points_won 		= points_won + $9,
		current_streak	= $10
	WHERE user_id 		= $1 AND channel_id = $2 AND team_id = $3;
	`

	p1, p2 := &Player{}, &Player{}
	result := &MatchResult{}

	args := strings.Split(commandText, " ")
	if len(args) < 3 {
		return result, fmt.Errorf("not enough arguments in command")
	}

	getUserID(p1, args[1])
	getUserID(p2, args[0])

	if p1.UserID == "" || p2.UserID == "" {
		return result, fmt.Errorf("invalid player tags %s, %s", p1.UserID, p2.UserID)
	}

	p1.channelID = channelID
	p2.channelID = channelID

	p1.teamID = teamID
	p2.teamID = teamID

	// check if user1 exists, if not INSERT into db
	exists, err := p.checkUserExists(p1)
	if err != nil {
		return result, err
	}
	if !exists {
		_, err = p.db.Exec(queryInsert, p1.UserID, p1.channelID, p1.teamID, 0, 0, 0, 0, 0, 0, 0)

		if err != nil {
			return result, err
		}

	}

	// check if user2 exists, if not INSERT into db
	exists, err = p.checkUserExists(p2)
	if err != nil {
		return result, err
	}
	if !exists {
		_, err = p.db.Exec(queryInsert, p2.UserID, p2.channelID, p2.teamID, 0, 0, 0, 0, 0, 0, 0)
		if err != nil {
			return result, err
		}
	}

	v, err := p.getPlayerValue(p1, "current_streak")
	if err != nil {
		return result, err
	}
	// unfortunatelly have to check for int64 because of SQL ...
	if streak, ok := v.(int64); ok {
		p1.CurrentStreak = int(streak)
	} else {
		return result, fmt.Errorf("current_streak of player2 isn't an int")
	}

	v, err = p.getPlayerValue(p2, "current_streak")
	if err != nil {
		return result, err
	}
	if streak, ok := v.(int64); ok {
		p2.CurrentStreak = int(streak)
	} else {
		return result, fmt.Errorf("current_streak of player2 isn't an int")
	}

	games := args[2:]
	err = processGameResults(games, p1, p2)
	if err != nil {
		return result, err
	}

	_, err = p.db.Exec(queryUpdate, p1.UserID, p1.channelID, p1.teamID,
		p1.MatchesWon,
		p1.MatchesLost,
		p1.MatchesDrawn,
		p1.GamesWon,
		p1.GamesLost,
		p1.PointsWon,
		p1.CurrentStreak)

	if err != nil {
		return result, err
	}

	_, err = p.db.Exec(queryUpdate, p2.UserID, p2.channelID, p2.teamID,
		p2.MatchesWon,
		p2.MatchesLost,
		p2.MatchesDrawn,
		p2.GamesWon,
		p2.GamesLost,
		p2.PointsWon,
		p2.CurrentStreak)

	if err != nil {
		return result, err
	}

	result.Winner = p1
	result.Loser = p2
	if p1.GamesWon < p2.GamesWon {
		result.Winner = p2
		result.Loser = p1
	}

	err = p.storeMatchDetails(p1, p2)
	if err != nil {
		return result, err
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
			if score1 != 11 && score2 != 11 {
				return fmt.Errorf("one of the scores in the game %s should be 11", game)
			}
		}

		p1.PointsWon += score1
		p2.PointsWon += score2

		if score1 > score2 {
			p1.GamesWon++
			p2.GamesLost++
		} else if score1 < score2 {
			p2.GamesWon++
			p1.GamesLost++
		}
	}

	switch {
	case p1.GamesWon > p2.GamesWon:
		p1.MatchesWon++
		p2.MatchesLost++
		p1.CurrentStreak++
		p2.CurrentStreak = 0

	case p1.GamesWon < p2.GamesWon:
		p2.MatchesWon++
		p1.MatchesLost++
		p1.CurrentStreak = 0
		p2.CurrentStreak++

	default:
		p1.MatchesDrawn++
		p2.MatchesDrawn++
		p1.CurrentStreak = 0
		p2.CurrentStreak = 0
	}

	return nil
}

func (p *Pong) checkUserExists(player *Player) (bool, error) {

	v, err := p.getPlayerValue(player, "id")
	if err != nil {
		return false, fmt.Errorf("error accessing player")
	}

	if _, ok := v.(int64); ok {
		return true, nil
	}

	return false, nil
}

func (p *Pong) getPlayerValue(player *Player, columnName string) (interface{}, error) {
	allowedColumns := map[string]bool{
		"id":             true,
		"matches_won":    true,
		"matches_lost":   true,
		"matches_drawn":  true,
		"games_won":      true,
		"games_lost":     true,
		"points_won":     true,
		"current_streak": true,
	}

	if !allowedColumns[columnName] {
		return nil, fmt.Errorf("invalid column name: %s", columnName)
	}

	querySelect := fmt.Sprintf(`
		SELECT %s
		FROM players
		WHERE user_id 		= $1 
		  AND channel_id 	= $2 
		  AND team_id 		= $3;
	`, columnName)

	var v interface{}
	err := p.db.QueryRow(querySelect, player.UserID, player.channelID, player.teamID).Scan(&v)
	if err != sql.ErrNoRows {
		return v, nil
	}

	return nil, nil
}

func validateUserTag(tag string) string {
	regex := `<@([A-Z0-9]+)\|([a-zA-Z0-9._-]+)>`
	re := regexp.MustCompile(regex)

	return re.FindString(tag)
}

func getUserID(p *Player, tag string) {
	v := validateUserTag(tag)

	if v != "" {
		p.UserID = strings.Split(strings.TrimPrefix(v, "<@"), "|")[0]
	}

}

func (p *Pong) Stats(channelID, teamID, commandText string) (*Player, error) {

	querySelect := `
	SELECT matches_won, matches_lost, matches_drawn, games_won, games_lost, points_won, current_streak
	FROM players
	WHERE 	user_id		= $1
		AND channel_id 	= $2
		AND team_id		= $3
	`

	player := &Player{channelID: channelID, teamID: teamID}
	args := strings.Split(commandText, " ")
	log.Println(commandText)

	if len(args) != 1 {
		return nil, fmt.Errorf("/stats command should have exactly 1 argument, the player tag")
	}

	log.Println(args[0])

	getUserID(player, args[0])

	row := p.db.QueryRow(
		querySelect,
		player.UserID,
		player.channelID,
		player.teamID,
	).Scan(
		&player.MatchesWon,
		&player.MatchesLost,
		&player.MatchesDrawn,
		&player.GamesWon,
		&player.GamesLost,
		&player.PointsWon,
		&player.CurrentStreak,
	)

	if row != sql.ErrNoRows {
		return player, nil
	}

	return &Player{}, nil

}

func (p *Pong) storeMatchDetails(p1, p2 *Player) error {
	v, err := p.getPlayerValue(p1, "id")
	if err != nil {
		return err
	}
	id1, ok := v.(int64)
	if !ok {
		return fmt.Errorf("id from p1 isn't an int")
	}

	v, err = p.getPlayerValue(p2, "id")
	if err != nil {
		return err
	}
	id2, ok := v.(int64)
	if !ok {
		return fmt.Errorf("id from p2 isn't an int")
	}

	queryInsertMatch := `
		INSERT INTO match_history (player_id_1, player_id_2, games_won_1, games_won_2)
		VALUES ($1, $2, $3, $4);
	`

	_, err = p.db.Exec(queryInsertMatch, id1, id2, p1.GamesWon, p2.GamesWon)
	if err != nil {
		return fmt.Errorf("failed to insert match details: %w", err)
	}

	return nil

func (p *Pong) Leaderboard(channelID string) ([]Player, error) {
	query := `
		SELECT full_name, matches_won, matches_drawn, matches_lost
		FROM players
		WHERE channel_id = $1
		ORDER BY matches_won DESC
		LIMIT 15
	`

	rows, err := p.db.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []Player

	for rows.Next() {
		var player Player
		err = rows.Scan(
			&player.FullName,
			&player.MatchesWon,
			&player.MatchesDrawn,
			&player.MatchesLost,
		)
		if err != nil {
			return nil, err
		}

		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}
