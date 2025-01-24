package pong

import "fmt"

func (p *Pong) GetOrElseAddPlayer(userID, channelID, teamID string) (*Player, error) {
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
	player := &Player{}
	err := p.db.QueryRow(query, userID, channelID, teamID).Scan(
		&player.id,
		&player.UserID,
		&player.channelID,
		&player.teamID,
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

func (p *Pong) AddMatchToHistory(p1, p2 *Player) error {
	query := `
	INSERT INTO match_history (player1_id, player2_id, player1_games_won, player2_games_won)
	VALUES ($1, $2, $3, $4);
	`

	_, err := p.db.Exec(query, p1.id, p2.id, p1.GamesWon, p2.GamesWon)
	if err != nil {
		return fmt.Errorf("failed to insert match details: %w", err)
	}

	return nil
}

func (p *Pong) UpdatePlayer(player *Player) error {
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
	_, err := p.db.Exec(query,
		player.MatchesWon,
		player.MatchesDrawn,
		player.MatchesLost,
		player.GamesWon,
		player.GamesLost,
		player.PointsWon,
		player.CurrentStreak,
		player.Elo,
		player.UserID,
		player.channelID,
		player.teamID,
	)
	if err != nil {
		return err
	}

	return nil
}
