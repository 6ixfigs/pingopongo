package matches

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/6ixfigs/pingypongy/internal/models"
	"github.com/6ixfigs/pingypongy/internal/webhooks"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Rtr chi.Router
	db  *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		Rtr: chi.NewRouter(),
		db:  db,
	}
}

func (h *Handler) MountRoutes() {
	h.Rtr.Post("/", h.Record)
}

func (h *Handler) Record(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := chi.URLParam(r, "leaderboard_name")
	username1 := r.FormValue("player1")
	username2 := r.FormValue("player2")
	score := r.FormValue("score")

	if username1 == username2 {
		http.Error(w, "player can't play against himself", http.StatusInternalServerError)
		return
	}

	matchScore, err := parseScore(score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	query := `
	SELECT id, name FROM leaderboards
	WHERE name = $1
	`

	leaderboard := &models.Leaderboard{}
	err = tx.QueryRow(query, name).Scan(
		&leaderboard.ID,
		&leaderboard.Name,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query = `
	SELECT * FROM players
	WHERE leaderboard_id = $1 AND username = $2
	`
	player1 := &models.Player{}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	player2 := &models.Player{}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	updateElo(winner, loser, matchScore.P1 == matchScore.P2)

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := &models.MatchResult{
		P1:        player1,
		P2:        player2,
		P1EloDiff: player1.Elo - p1OldElo,
		P2EloDiff: player2.Elo - p2OldElo,
		Score:     matchScore,
	}

	response := fmt.Sprintf("Match recorded: (%+d) %s %d - %d %s (%+d) !\n",
		result.P1EloDiff,
		result.P1.Username,
		result.Score.P1,
		result.Score.P2,
		result.P2.Username,
		result.P2EloDiff,
	)

	urls, err := webhooks.All(h.db, name)
	if err == nil {
		webhooks.Notify(urls, response)
	}

	w.Write([]byte(response))
}

func parseScore(score string) (*models.MatchScore, error) {
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

	return &models.MatchScore{
		P1: p1Score,
		P2: p2Score,
	}, nil
}

func updateElo(winner, loser *models.Player, isDraw bool) {
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
