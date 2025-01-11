package pong

import (
	"database/sql"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Pong struct {
	db *sql.DB
}

func New(db *sql.DB) *Pong {
	return &Pong{db}
}

func (p *Pong) Leaderboard(channelID string) (string, error) {
	query := `
		SELECT full_name, matches_won, matches_drawn, matches_lost
		FROM players
		WHERE channel_id = $1
		ORDER BY matches_won DESC
		LIMIT 15
	`

	rows, err := p.db.Query(query, channelID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var players []player

	for rows.Next() {
		var p player
		err = rows.Scan(
			&p.fullName,
			&p.matchesWon,
			&p.matchesDrawn,
			&p.matchesLost,
		)
		if err != nil {
			return "", err
		}

		players = append(players, p)
	}

	if err = rows.Err(); err != nil {
		return "", err
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio"})
	for rank, player := range players {
		matchesPlayed := player.matchesWon + player.matchesDrawn + player.matchesLost
		t.AppendRow(table.Row{
			rank + 1,
			player.fullName,
			player.matchesWon,
			player.matchesDrawn,
			player.matchesLost,
			matchesPlayed,
			fmt.Sprintf("%.2f", float64(player.matchesWon)/float64(matchesPlayed)*100),
		})
	}
	leaderboard := fmt.Sprintf("```%s```", t.Render())

	return leaderboard, nil
}
