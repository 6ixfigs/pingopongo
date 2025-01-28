package pong

import (
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/6ixfigs/pingypongy/internal/slack"
)

type Pong struct {
	db *sql.DB
}

func New(db *sql.DB) *Pong {
	return &Pong{db: db}
}

func (p *Pong) Record(channelID, teamID, commandText string) (matchResult *MatchResult, err error) {
	args := strings.Split(commandText, " ")
	if len(args) < 3 {
		return nil, fmt.Errorf("not enough arguments in command")
	}

	for _, userMention := range args[:2] {
		err := validateUserMention(userMention)
		if err != nil {
			return nil, err
		}
	}

	err = validateUserMention(args[0])
	if err != nil {
		return nil, err
	}

	err = validateUserMention(args[1])
	if err != nil {
		return nil, err
	}

	err = validateGames(args[2:])
	if err != nil {
		return nil, err
	}

	user1ID := extractUserID(args[0])
	user2ID := extractUserID(args[1])

	if user1ID == user2ID {
		return nil, fmt.Errorf("player can't play against himself")
	}

	user1Name, err := slack.GetUserInfo(user1ID)
	if err != nil {
		return nil, err
	}

	user2Name, err := slack.GetUserInfo(user2ID)
	if err != nil {
		return nil, err
	}

	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				matchResult = nil
			}
		} else {
			if err = tx.Commit(); err != nil {
				matchResult = nil
			}
		}
	}()

	query := `
	WITH inserted AS (
		INSERT INTO players (user_id, channel_id, team_id, full_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, channel_id, team_id) DO NOTHING
		RETURNING *
	)
	SELECT * FROM inserted
	UNION
	SELECT * FROM players
	WHERE user_id = $1 AND channel_id = $2 AND team_id = $3;
	`
	player1 := &Player{}
	err = tx.QueryRow(query, user1ID, channelID, teamID, user1Name).Scan(
		&player1.id,
		&player1.UserID,
		&player1.channelID,
		&player1.teamID,
		&player1.FullName,
		&player1.MatchesWon,
		&player1.MatchesLost,
		&player1.MatchesDrawn,
		&player1.TotalGamesWon,
		&player1.TotalGamesLost,
		&player1.TotalPointsWon,
		&player1.CurrentStreak,
		&player1.Elo,
	)
	if err != nil {
		return nil, err
	}

	player2 := &Player{}
	err = tx.QueryRow(query, user2ID, channelID, teamID, user2Name).Scan(
		&player2.id,
		&player2.UserID,
		&player2.channelID,
		&player2.teamID,
		&player2.FullName,
		&player2.MatchesWon,
		&player2.MatchesLost,
		&player2.MatchesDrawn,
		&player2.TotalGamesWon,
		&player2.TotalGamesLost,
		&player2.TotalPointsWon,
		&player2.CurrentStreak,
		&player2.Elo,
	)
	if err != nil {
		return nil, err
	}

	games := args[2:]
	results := determineGameResults(games, player1, player2)

	matchResult = &MatchResult{}
	matchResult.P1 = player1
	matchResult.P2 = player2
	for _, result := range results {
		player1.TotalPointsWon += result.P1PointsWon
		player2.TotalPointsWon += result.P2PointsWon

		if player1 == result.Winner {
			matchResult.P1GamesWon++
		} else {
			matchResult.P2GamesWon++
		}
	}

	player1.TotalGamesWon += matchResult.P1GamesWon
	player1.TotalGamesLost += matchResult.P2GamesWon

	player2.TotalGamesWon += matchResult.P2GamesWon
	player2.TotalGamesLost += matchResult.P1GamesWon

	matchResult.Games = results

	if matchResult.P1GamesWon > matchResult.P2GamesWon {
		matchResult.Winner = player1
		player1.MatchesWon++
		player1.CurrentStreak++
		player2.MatchesLost++
		player2.CurrentStreak = 0
	} else if matchResult.P2GamesWon > matchResult.P1GamesWon {
		matchResult.Winner = player2
		player1.MatchesLost++
		player1.CurrentStreak = 0
		player2.MatchesWon++
		player2.CurrentStreak++
	} else {
		matchResult.IsDraw = true
		player1.MatchesDrawn++
		player1.CurrentStreak = 0
		player2.MatchesDrawn++
		player2.CurrentStreak = 0
	}

	p.updateElo(matchResult)

	query = `
	UPDATE players
	SET
		matches_won = $1,
		matches_lost = $2,
		matches_drawn = $3,
		total_games_won = $4,
		total_games_lost = $5,
		total_points_won = $6,
		current_streak = $7,
		elo = $8
	WHERE user_id = $9 AND channel_id = $10 AND team_id = $11
	`
	_, err = tx.Exec(query,
		player1.MatchesWon,
		player1.MatchesDrawn,
		player1.MatchesLost,
		player1.TotalGamesWon,
		player1.TotalGamesLost,
		player1.TotalPointsWon,
		player1.CurrentStreak,
		player1.Elo,
		player1.UserID,
		player1.channelID,
		player1.teamID,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(query,
		player2.MatchesWon,
		player2.MatchesDrawn,
		player2.MatchesLost,
		player2.TotalGamesWon,
		player2.TotalGamesLost,
		player2.TotalPointsWon,
		player2.CurrentStreak,
		player2.Elo,
		player2.UserID,
		player2.channelID,
		player2.teamID,
	)
	if err != nil {
		return nil, err
	}

	query = `
	INSERT INTO match_history (player1_id, player2_id, player1_games_won, player2_games_won)
	VALUES ($1, $2, $3, $4);
	`

	_, err = tx.Exec(query, player1.id, player2.id, player1.TotalGamesWon, player2.TotalGamesWon)
	if err != nil {
		return nil, fmt.Errorf("failed to insert match details: %w", err)
	}

	if err != nil {
		return nil, err
	}

	return matchResult, err
}

func (p *Pong) Leaderboard(channelID string) ([]Player, error) {
	query := `
		SELECT full_name, matches_won, matches_drawn, matches_lost, elo
		FROM players
		WHERE channel_id = $1
		ORDER BY elo DESC
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

func (p *Pong) Stats(channelID, teamID, commandText string) (*Player, error) {
	args := strings.Split(commandText, " ")
	if len(args) != 1 {
		return nil, fmt.Errorf("/stats command should have exactly 1 argument, the player tag")
	}

	err := validateUserMention(args[0])
	if err != nil {
		return nil, err
	}

	userID := extractUserID(args[0])

	querySelect := `
	SELECT matches_won, matches_lost, matches_drawn, total_games_won, total_games_lost, total_points_won, current_streak, elo
	FROM players
	WHERE 	user_id		= $1
		AND channel_id 	= $2
		AND team_id	= $3
	`

	player := &Player{UserID: userID, channelID: channelID, teamID: teamID}
	err = p.db.QueryRow(
		querySelect,
		userID,
		channelID,
		teamID,
	).Scan(
		&player.MatchesWon,
		&player.MatchesLost,
		&player.MatchesDrawn,
		&player.TotalGamesWon,
		&player.TotalGamesLost,
		&player.TotalPointsWon,
		&player.CurrentStreak,
		&player.Elo,
	)

	if err != nil {
		return nil, err
	}

	return player, nil

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

func (p *Pong) updateElo(result *MatchResult) {
	q1 := math.Pow(10, float64(result.P1.Elo)/400)
	q2 := math.Pow(10, float64(result.P2.Elo)/400)

	e1 := q1 / (q1 + q2)
	e2 := q2 / (q1 + q2)

	kFactor := func(rating int) float64 {
		if rating < 2100 {
			return 32
		}
		if rating >= 2100 && rating < 2400 {
			return 24
		}
		return 16
	}

	k1 := kFactor(result.P1.Elo)
	k2 := kFactor(result.P2.Elo)

	s1, s2 := 1.0, 0.0
	if result.Winner == result.P2 {
		s1, s2 = 0.0, 1.0
	} else if result.IsDraw {
		s1, s2 = 0.5, 0.5
	}

	result.P1.Elo = result.P1.Elo + int(math.Round(k1*(s1-e1)))
	result.P2.Elo = result.P2.Elo + int(math.Round(k2*(s2-e2)))
}

func validateUserMention(rawUserMention string) error {
	regex := `<@([A-Z0-9]+)\|([a-zA-Z0-9._-]+)>`
	re := regexp.MustCompile(regex)

	if re.FindString(rawUserMention) == "" {
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
		result.P1 = p1
		result.P2 = p2

		scores := strings.Split(game, "-")

		result.P1PointsWon, _ = strconv.Atoi(scores[0])
		result.P2PointsWon, _ = strconv.Atoi(scores[1])

		if result.P1PointsWon > result.P2PointsWon {
			result.Winner = p1
		} else {
			result.Winner = p2
		}

		results = append(results, result)
	}

	return results
}

func extractUserID(rawMention string) string {
	return strings.Split(strings.TrimPrefix(rawMention, "<@"), "|")[0]
}
