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

func (p *Pong) Record(channelID, teamID, commandText string) (*MatchResult, error) {
	args := strings.Split(commandText, " ")
	if len(args) < 3 {
		return nil, fmt.Errorf("not enough arguments in command")
	}

	err := validateUsers(args[0], args[1])
	if err != nil {
		return nil, err
	}

	err = validateGames(args[2:])
	if err != nil {
		return nil, err
	}

	id1 := extractUserID(args[0])
	id2 := extractUserID(args[1])
	games := args[2:]

	query := `
	WITH inserted AS (
		INSERT INTO players (user_id, channel_id, team_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id, teams_id) DO NOTHING
		RETURNING *
	)
	SELECT * FROM inserted
	UNION
	SELECT * FROM players
	WHERE user_id = $1 AND channel_id = $2 AND team_id = $3;
	`
	player1 := &Player{}
	err = p.db.QueryRow(query, id1, channelID, teamID).Scan(
		&player1.id,
		&player1.UserID,
		&player1.channelID,
		&player1.teamID,
		&player1.MatchesWon,
		&player1.MatchesLost,
		&player1.MatchesDrawn,
		&player1.GamesWon,
		&player1.GamesLost,
		&player1.PointsWon,
		&player1.CurrentStreak,
		&player1.Elo,
	)
	if err != nil {
		return nil, err
	}

	player2 := &Player{}
	err = p.db.QueryRow(query, id2, channelID, teamID).Scan(
		&player2.id,
		&player2.UserID,
		&player2.channelID,
		&player2.teamID,
		&player2.MatchesWon,
		&player2.MatchesLost,
		&player2.MatchesDrawn,
		&player2.GamesWon,
		&player2.GamesLost,
		&player2.PointsWon,
		&player2.CurrentStreak,
		&player2.Elo,
	)
	if err != nil {
		return nil, err
	}

	results := determineGameResults(games, player1, player2)

	var p1GamesWon, p2GamesWon int
	for _, result := range results {
		player1.PointsWon += result.p1Points
		player2.PointsWon += result.p2Points

		if player1 == result.winner {
			p1GamesWon++
		} else {
			p2GamesWon++
		}
	}

	player1.GamesWon += p1GamesWon
	player1.GamesLost += p2GamesWon

	player2.GamesWon += p2GamesWon
	player2.GamesLost += p1GamesWon

	var winner, loser *Player
	if p1GamesWon > p2GamesWon {
		winner = player1
		player1.MatchesWon++
		player1.CurrentStreak++
		loser = player2
		player2.MatchesLost++
		player2.CurrentStreak = 0
	} else if p1GamesWon < p2GamesWon {
		loser = player1
		player1.MatchesLost--
		player1.CurrentStreak = 0
		winner = player2
		player2.MatchesWon++
		player2.CurrentStreak++
	} else {
		player1.MatchesDrawn++
		player1.CurrentStreak = 0
		player2.MatchesDrawn++
		player2.CurrentStreak = 0
	}

	queryUpdate := `
	UPDATE players
	SET
		matches_won = $1,
		matches_lost = $2,
		matches_drawn = $3,
		games_won = $4,
		games_lost = $5,
		points_won = $6,
		current_streak = $7,
		elo = $8
	WHERE user_id = $9 AND channel_id = $10 AND team_id = $11
	`
	_, err = p.db.Exec(queryUpdate,
		player1.MatchesWon,
		player1.MatchesDrawn,
		player1.MatchesLost,
		player1.GamesWon,
		player1.GamesLost,
		player1.PointsWon,
		player1.CurrentStreak,
		player1.Elo,
		player1.UserID,
		player1.channelID,
		player1.teamID,
	)
	if err != nil {
		return nil, err
	}

	_, err = p.db.Exec(queryUpdate,
		player2.MatchesWon,
		player2.MatchesDrawn,
		player2.MatchesLost,
		player2.GamesWon,
		player2.GamesLost,
		player2.PointsWon,
		player2.CurrentStreak,
		player2.Elo,
		player2.UserID,
		player2.channelID,
		player2.teamID,
	)
	if err != nil {
		return nil, err
	}

	err = p.addMatchToHistory(player1, player2)
	if err != nil {
		return nil, err
	}

	matchResult := &MatchResult{winner, loser, results}
	return matchResult, nil
}

func validateUsers(user1, user2 string) error {
	regex := `<@([A-Z0-9]+)\|([a-zA-Z0-9._-]+)>`
	re := regexp.MustCompile(regex)

	if re.FindString(user1) == "" || re.FindString(user2) == "" {
		return fmt.Errorf("not a valid user")
	}

	return nil
}

func validateGames(games []string) error {
	for _, game := range games {
		if !strings.Contains(game, "-") {
			return fmt.Errorf("game %s needs to have '-' separator", game)
		}

		scores := strings.Split(game, "-")

		if len(scores) != 2 {
			return fmt.Errorf("invalid game format: %s", game)
		}

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
		} else if score1 != 11 && score2 != 11 {
			return fmt.Errorf("one of the scores in the game %s should be 11", game)
		}
	}

	return nil
}

func determineGameResults(games []string, p1, p2 *Player) []GameResult {

	var results []GameResult

	for _, game := range games {
		result := GameResult{}
		scores := strings.Split(game, "-")

		result.p1Points, _ = strconv.Atoi(scores[0])
		result.p2Points, _ = strconv.Atoi(scores[1])

		if result.p1Points > result.p2Points {
			result.winner = p1
		} else {
			result.winner = p2
		}

		results = append(results, result)
	}

	return results
}

func extractUserID(rawMention string) string {
	return strings.Split(strings.TrimPrefix(rawMention, "<@"), "|")[0]
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

	if len(args) != 1 {
		return nil, fmt.Errorf("/stats command should have exactly 1 argument, the player tag")
	}

	//	getUserID(player, args[0])

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

func (p *Pong) addMatchToHistory(p1, p2 *Player) error {
	query := `
	INSERT INTO match_history (player_id_1, player_id_2, games_won_1, games_won_2)
	VALUES ($1, $2, $3, $4);
	`

	_, err := p.db.Exec(query, p1.id, p2.id, p1.GamesWon, p2.GamesWon)
	if err != nil {
		return fmt.Errorf("failed to insert match details: %w", err)
	}

	return nil
}

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
			&player.Elo,
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

func (p *Pong) UpdateChannelID(oldID, newID string) error {
	query := `
	UPDATE players
	SET channel_id = $1
	WHERE channel_id = $2
	`

	_, err := p.db.Exec(query, newID, oldID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pong) UpdateElo(winner, loser *Player, isDraw bool) {
	qW := math.Pow(10, float64(winner.Elo)/400)
	qL := math.Pow(10, float64(loser.Elo)/400)

	eW := qW / (qW + qL)
	eL := qL / (qW + qL)

	kFactor := func(rating int) float64 {
		if rating < 2100 {
			return 32
		}
		if rating >= 2100 && rating < 2400 {
			return 24
		}
		return 16
	}

	kW := kFactor(winner.Elo)
	kL := kFactor(loser.Elo)

	sW, sL := 1.0, 0.0
	if isDraw {
		sW, sL = 0.5, 0.5
	}

	winner.Elo = winner.Elo + int(math.Round(kW*(sW-eW)))
	loser.Elo = loser.Elo + int(math.Round(kL*(sL-eL)))
}
