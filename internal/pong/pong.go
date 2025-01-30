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
	return &Pong{db: db}
}

func (p *Pong) Record(leaderboardName, username1, username2, score string) (matchResult *MatchResult, err error) {
	matchScore, err := parseScore(score)
	if err != nil {
		return nil, err
	}

	if username1 == username2 {
		return nil, fmt.Errorf("player can't play against himself")
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
	SELECT id, name FROM leaderboards
	WHERE name = $1
	`

	leaderboard := &Leaderboard{}
	err = tx.QueryRow(query, leaderboardName).Scan(
		&leaderboard.ID,
		&leaderboard.Name,
	)
	if err != nil {
		return nil, err
	}

	query = `
	SELECT * FROM players
	WHERE leaderboard_id = $1 AND username = $2
	`
	player1 := &Player{}
	err = tx.QueryRow(query, leaderboard.ID, username1).Scan(
		&player1.ID,
		&player1.LeaderboardID,
		&player1.Username,
		&player1.MatchesWon,
		&player1.MatchesDrawn,
		&player1.MatchesLost,
		&player1.TotalGamesWon,
		&player1.TotalGamesLost,
		&player1.CurrentStreak,
		&player1.Elo,
		&player1.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	player2 := &Player{}
	err = tx.QueryRow(query, leaderboard.ID, username2).Scan(
		&player2.ID,
		&player2.LeaderboardID,
		&player2.Username,
		&player2.MatchesWon,
		&player2.MatchesDrawn,
		&player2.MatchesLost,
		&player2.TotalGamesWon,
		&player2.TotalGamesLost,
		&player2.CurrentStreak,
		&player2.Elo,
		&player2.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	player1.TotalGamesWon += matchScore.P1
	player1.TotalGamesLost += matchScore.P2

	player2.TotalGamesWon += matchScore.P2
	player2.TotalGamesLost += matchScore.P1

	winner, loser := player1, player2
	if matchScore.P1 > matchScore.P2 {
		player1.MatchesWon++
		player1.CurrentStreak++
		player2.MatchesLost++
		player2.CurrentStreak = 0
	} else if matchScore.P2 > matchScore.P1 {
		winner = player2
		loser = player1
		player1.MatchesLost++
		player1.CurrentStreak = 0
		player2.MatchesWon++
		player2.CurrentStreak++
	} else {
		player1.MatchesDrawn++
		player1.CurrentStreak = 0
		player2.MatchesDrawn++
		player2.CurrentStreak = 0
	}

	p.updateElo(winner, loser, matchScore.P1 == matchScore.P2)

	query = `
	UPDATE players
	SET
		matches_won = $1,
		matches_drawn = $2,
		matches_lost = $3,
		total_games_won = $4,
		total_games_lost = $5,
		current_streak = $6,
		elo = $7
	WHERE id = $8 
	`
	_, err = tx.Exec(query,
		player1.MatchesWon,
		player1.MatchesDrawn,
		player1.MatchesLost,
		player1.TotalGamesWon,
		player1.TotalGamesLost,
		player1.CurrentStreak,
		player1.Elo,
		player1.ID,
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
		player2.CurrentStreak,
		player2.Elo,
		player2.ID,
	)
	if err != nil {
		return nil, err
	}

	query = `
	INSERT INTO matches (leaderboard_id, player1_id, player2_id, score)
	VALUES ($1, $2, $3, $4);
	`

	_, err = tx.Exec(query,
		leaderboard.ID,
		player1.ID,
		player2.ID,
		score,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert match details: %w", err)
	}

	matchResult = &MatchResult{player1, player2, matchScore}
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

func (p *Pong) updateElo(winner, loser *Player, isDraw bool) {
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

func parseScore(score string) (*MatchScore, error) {
	if !strings.Contains(score, "-") {
		return nil, fmt.Errorf("match score %s needs to have '-' separator", score)
	}

	splitScore := strings.Split(score, "-")

	if len(splitScore) != 2 {
		return nil, fmt.Errorf("match invalid score format: %s", score)
	}

	p1Score, err := strconv.Atoi(splitScore[0])
	if err != nil {
		return nil, fmt.Errorf("player1 score needs to be a number")
	}

	p2Score, err := strconv.Atoi(splitScore[1])
	if err != nil {
		return nil, fmt.Errorf("player2 score needs to be a number")
	}

	return &MatchScore{p1Score, p2Score}, nil
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
