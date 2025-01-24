package repository

import (
	"database/sql"
	"fmt"

	"github.com/6ixfigs/pingypongy/internal/types"
)

type SQLRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

type Repository interface {
	Begin() (*SQLRepository, error)
	Commit() error
	Rollback() error

	GetOrElseAddPlayer(userID, channelID, teamID string) (*types.Player, error)
	UpdateChannelID(oldID, newID string) error
	AddMatchToHistory(p1, p2 *types.Player) error
	UpdatePlayer(player *types.Player) error
	GetLeaderboardData(channelID string) (*sql.Rows, error)
}

func (repo *SQLRepository) Begin() (*SQLRepository, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	return &SQLRepository{db: repo.db, tx: tx}, nil
}

func (repo *SQLRepository) Commit() error {
	if repo.tx != nil {
		return repo.tx.Commit()
	}

	return fmt.Errorf("no transaction to commit")
}

func (repo *SQLRepository) Rollback() error {
	if repo.tx != nil {
		return repo.tx.Rollback()
	}

	return fmt.Errorf("no transaction to rollback")
}

func (repo *SQLRepository) GetOrElseAddPlayer(userID, channelID, teamID string) (*types.Player, error) {
	query := `
	WITH inserted AS (
		INSERT INTO players (user_id, channel_id, team_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id, team_id) DO NOTHING
		RETURNING *
	)
	SELECT * FROM inserted
	UNION
	SELECT * FROM players
	WHERE user_id = $1 AND channel_id = $2 AND team_id = $3;
	`
	player := &types.Player{}
	err := repo.db.QueryRow(query, userID, channelID, teamID).Scan(
		&player.Id,
		&player.UserID,
		&player.ChannelID,
		&player.TeamID,
		&player.FullName,
		&player.MatchesWon,
		&player.MatchesLost,
		&player.MatchesDrawn,
		&player.GamesWon,
		&player.GamesLost,
		&player.PointsWon,
		&player.CurrentStreak,
		&player.Elo,
	)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (repo *SQLRepository) UpdateChannelID(oldID, newID string) error {
	query := `
	UPDATE players
	SET channel_id = $1
	WHERE channel_id = $2
	`

	_, err := repo.db.Exec(query, newID, oldID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SQLRepository) AddMatchToHistory(p1, p2 *types.Player) error {
	query := `
	INSERT INTO match_history (player1_id, player2_id, player1_games_won, player2_games_won)
	VALUES ($1, $2, $3, $4);
	`

	_, err := repo.db.Exec(query, p1.Id, p2.Id, p1.GamesWon, p2.GamesWon)
	if err != nil {
		return fmt.Errorf("failed to insert match details: %w", err)
	}

	return nil
}

func (repo *SQLRepository) UpdatePlayer(player *types.Player) error {
	query := `
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
	_, err := repo.db.Exec(query,
		player.MatchesWon,
		player.MatchesDrawn,
		player.MatchesLost,
		player.GamesWon,
		player.GamesLost,
		player.PointsWon,
		player.CurrentStreak,
		player.Elo,
		player.UserID,
		player.ChannelID,
		player.TeamID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SQLRepository) GetLeaderboardData(channelID string) (*sql.Rows, error) {
	query := `
		SELECT full_name, matches_won, matches_drawn, matches_lost, elo
		FROM players
		WHERE channel_id = $1
		ORDER BY elo DESC
		LIMIT 15
	`

	return repo.db.Query(query, channelID)
}
