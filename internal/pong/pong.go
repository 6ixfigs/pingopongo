package pong

import (
	"database/sql"
)

type Pong struct {
	db *sql.DB
}

func New(db *sql.DB) *Pong {
	return &Pong{db}
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
