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

	for _, userMention := range args[:2] {
		err := validateUserMention(userMention)
		if err != nil {
			return nil, err
		}
	}

	err := validateGames(args[2:])
	if err != nil {
		return nil, err
	}

	user1ID := extractUserID(args[0])
	user2ID := extractUserID(args[1])

	if user1ID == user2ID {
		return nil, fmt.Errorf("player can't play against himself")
	}

	games := args[2:]

	player1, err := p.getOrElseAddPlayer(user1ID, channelID, teamID)
	if err != nil {
		return nil, err
	}

	player2, err := p.getOrElseAddPlayer(user2ID, channelID, teamID)
	if err != nil {
		return nil, err
	}

	results := determineGameResults(games, player1, player2)

	var p1GamesWon, p2GamesWon int
	for _, result := range results {
		player1.PointsWon += result.P1Points
		player2.PointsWon += result.P2Points

		if player1 == result.Winner {
			p1GamesWon++
		} else {
			p2GamesWon++
		}
	}

	player1.GamesWon += p1GamesWon
	player1.GamesLost += p2GamesWon

	player2.GamesWon += p2GamesWon
	player2.GamesLost += p1GamesWon

	matchResult := &MatchResult{}
	matchResult.Games = results

	if p1GamesWon > p2GamesWon {
		matchResult.Winner = player1
		matchResult.Loser = player2
	} else if p2GamesWon > p1GamesWon {
		matchResult.Winner = player2
		matchResult.Loser = player1
	} else {
		matchResult.IsDraw = true
	}

	if matchResult.IsDraw {
		player1.MatchesDrawn++
		player1.CurrentStreak = 0
		player2.MatchesDrawn++
		player2.CurrentStreak = 0
	} else {
		matchResult.Winner.MatchesWon++
		matchResult.Winner.CurrentStreak++
		matchResult.Loser.MatchesLost++
		matchResult.Loser.CurrentStreak = 0
	}

	p.updateElo(matchResult)

	err = p.updatePlayer(player1)
	if err != nil {
		return nil, err
	}

	err = p.updatePlayer(player2)
	if err != nil {
		return nil, err
	}

	err = p.addMatchToHistory(player1, player2)
	if err != nil {
		return nil, err
	}

	return matchResult, nil
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
	SELECT matches_won, matches_lost, matches_drawn, games_won, games_lost, points_won, current_streak, elo
	FROM players
	WHERE 	user_id		= $1
		AND channel_id 	= $2
		AND team_id	= $3
	`

	player := &Player{}
	err = p.db.QueryRow(
		querySelect,
		userID,
		channelID,
		teamID,
	).Scan(
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

func (p *Pong) updatePlayer(player *Player) error {
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

func (p *Pong) updateElo(matchResult *MatchResult) {
	qW := math.Pow(10, float64(matchResult.Winner.Elo)/400)
	qL := math.Pow(10, float64(matchResult.Loser.Elo)/400)

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

	kW := kFactor(matchResult.Winner.Elo)
	kL := kFactor(matchResult.Loser.Elo)

	sW, sL := 1.0, 0.0
	if matchResult.IsDraw {
		sW, sL = 0.5, 0.5
	}

	matchResult.Winner.Elo = matchResult.Winner.Elo + int(math.Round(kW*(sW-eW)))
	matchResult.Loser.Elo = matchResult.Loser.Elo + int(math.Round(kL*(sL-eL)))
}

func (p *Pong) getOrElseAddPlayer(userID, channelID, teamID string) (*Player, error) {
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
	player := &Player{}
	err := p.db.QueryRow(query, userID, channelID, teamID).Scan(
		&player.id,
		&player.UserID,
		&player.channelID,
		&player.teamID,
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
		scores := strings.Split(game, "-")

		result.P1Points, _ = strconv.Atoi(scores[0])
		result.P2Points, _ = strconv.Atoi(scores[1])

		if result.P1Points > result.P2Points {
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
