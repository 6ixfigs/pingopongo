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

	matchResult := &MatchResult{}
	matchResult.P1 = player1
	matchResult.P2 = player2
	for _, result := range results {
		player1.PointsWon += result.P1PointsWon
		player2.PointsWon += result.P2PointsWon

		if player1 == result.Winner {
			matchResult.P1GamesWon++
		} else {
			matchResult.P2GamesWon++
		}
	}

	player1.GamesWon += matchResult.P1GamesWon
	player1.GamesLost += matchResult.P2GamesWon

	player2.GamesWon += matchResult.P2GamesWon
	player2.GamesLost += matchResult.P1GamesWon

	matchResult.Games = results

	if matchResult.P1GamesWon > matchResult.P2GamesWon {
		matchResult.Winner = player1
		player1.MatchesWon++
		player1.CurrentStreak++
		player2.MatchesLost++
		player2.CurrentStreak = 0
	} else if matchResult.P2GamesWon > matchResult.P1GamesWon {
		matchResult.Winner = player2
		player1.MatchesLost--
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
	SELECT matches_won, matches_lost, matches_drawn, games_won, games_lost, points_won, current_streak, elo
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
	INSERT INTO match_history (player1_id, player2_id, player1_games_won, player2_games_won)
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

func (p *Pong) getOrElseAddPlayer(userID, channelID, teamID string) (*Player, error) {
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
