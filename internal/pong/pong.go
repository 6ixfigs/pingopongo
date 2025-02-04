package pong

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Pong struct {
	db *sql.DB
}

func New(db *sql.DB) *Pong {
	return &Pong{db: db}
}

func (p *Pong) CreateLeaderboard(leaderboardName string) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query := `
	INSERT INTO leaderboards (name) 
	VALUES ($1)
	`
	_, err = tx.Exec(query, leaderboardName)

	return err
}

func (p *Pong) RegisterWebhook(leaderboardName, url string) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
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
		return err
	}

	query = `
	INSERT INTO webhooks (leaderboard_id, url)
	VALUES ($1, $2)
	`
	_, err = tx.Exec(query, leaderboard.ID, url)

	return err
}

func (p *Pong) ListWebhooks(leaderboardName string) (webhooks []string, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				webhooks = nil
			}
		} else {
			if err := tx.Commit(); err != nil {
				webhooks = nil
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
	SELECT url FROM webhooks
	WHERE leaderboard_id = $1
	`
	rows, err := tx.Query(query, leaderboard.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (p *Pong) DeleteWebhooks(leaderboardName string) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
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
		return err
	}

	query = `
	DELETE FROM webhooks
	WHERE leaderboard_id = $1
	`
	_, err = tx.Exec(query, leaderboard.ID)

	return err
}

func (p *Pong) CreatePlayer(leaderboardName, username string) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
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
		return err
	}

	query = `
	INSERT INTO players (leaderboard_id, username)
	VALUES ($1, $2)
	`

	_, err = tx.Exec(query, leaderboard.ID, username)

	return err
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

	p1OldElo, p2OldElo := player1.Elo, player2.Elo
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

	matchResult = &MatchResult{
		player1,
		player2,
		player1.Elo - p1OldElo,
		player2.Elo - p2OldElo,
		matchScore,
	}
	return matchResult, nil
}

func (p *Pong) Leaderboard(leaderboardName string) (rankings []Player, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				rankings = nil
			}
		} else {
			if err = tx.Commit(); err != nil {
				rankings = nil
			}
		}
	}()

	query := `
	SELECT FROM leaderboards
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
		SELECT username, matches_won, matches_drawn, matches_lost, elo
		FROM players
		WHERE leaderboard_id = $1
		ORDER BY elo DESC
	`

	rows, err := p.db.Query(query, leaderboard.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var player Player
		err = rows.Scan(
			&player.Username,
			&player.MatchesWon,
			&player.MatchesDrawn,
			&player.MatchesLost,
			&player.Elo,
		)
		if err != nil {
			return nil, err
		}

		rankings = append(rankings, player)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rankings, nil
}

func (p *Pong) Stats(leaderboardName, username string) (player *Player, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				player = nil
			}
		} else {
			if err = tx.Commit(); err != nil {
				player = nil
			}
		}
	}()

	query := `
	SELECT FROM leaderboards
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
	WHERE 	leaderboard_id = $1 AND username = $2
	`

	err = p.db.QueryRow(query, leaderboard.ID, username).Scan(
		&player.ID,
		&player.LeaderboardID,
		&player.Username,
		&player.MatchesWon,
		&player.MatchesDrawn,
		&player.MatchesLost,
		&player.TotalGamesWon,
		&player.TotalGamesLost,
		&player.CurrentStreak,
		&player.Elo,
		&player.CreatedAt,
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
