package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"database/sql"
	"net/http"
	"net/url"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/go-chi/chi/v5"
)

const firstPlayer int = 0
const secondPlayer int = 1

type Server struct {
	Router chi.Router
	db     *sql.DB
	Config *config.Config
}

type GameResults struct {
	firstPlayerGamesWon    int
	firstPlayerGamesLost   int
	firstPlayerGamesDrawn  int
	secondPlayerGamesWon   int
	secondPlayerGamesLost  int
	secondPlayerGamesDrawn int
}

const (
	clientID     = "8148123983154.8265447105907"      // Replace with your Slack App's Client ID
	clientSecret = "de22b4b89ff61957fdd98f48ed61ce82" // Replace with your Slack App's Client Secret
	redirectURI  = "http://localhost:8080/auth"       // Your Redirect URL
)

func NewServer() (*Server, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	db, err := db.Connect(&cfg.DBConn)
	if err != nil {
		return nil, err
	}

	return &Server{
		Router: chi.NewRouter(),
		db:     db,
		Config: cfg,
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/record", s.record)
	s.Router.Post("/leaderboard", s.showLeaderboard)
	s.Router.Post("/auth", s.handleOAuthRedirect)
}

func (s *Server) handleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange the code for an access token
	tokenURL := "https://slack.com/api/oauth.v2.access"
	resp, err := http.PostForm(tokenURL, url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		log.Println("Error during token exchange:", err)
		return
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		log.Println("Error decoding response:", err)
		return
	}

	// Check for errors in Slack's response
	if !result["ok"].(bool) {
		http.Error(w, fmt.Sprintf("Slack API error: %v", result["error"]), http.StatusBadRequest)
		return
	}

	// Access token
	accessToken := result["access_token"].(string)
	fmt.Fprintf(w, "Authorization successful! Access token: %s", accessToken)

	// Store the token securely (e.g., database or encrypted storage)
	log.Printf("Access Token: %s", accessToken)
}

func (s *Server) record(w http.ResponseWriter, r *http.Request) {
	// Insert or update player stats
	queryInsertUpdateUser := `
		INSERT INTO player_stats (username, games_won, games_lost, games_drawn, sets_won, sets_lost)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (username)
		DO UPDATE SET
			games_won = EXCLUDED.games_won + $2,
			games_lost = EXCLUDED.games_lost + $3,
			games_drawn = EXCLUDED.games_drawn + $4,
			sets_won = sets_won + EXCLUDED.sets_won,
			sets_lost = sets_lost + EXCLUDED.sets_lost;
	`

	// Extract user and command text from Slack request
	username := strings.ReplaceAll(r.FormValue("user"), "@", "")
	commandText := r.FormValue("text")

	// Split the command text into words
	parts := strings.Fields(commandText)

	if len(parts) < 4 {
		http.Error(w, "Invalid command format", http.StatusBadRequest)
		return
	}

	// Extract player names (skip the @ symbol)
	firstPlayer := strings.TrimPrefix(parts[0], "@")
	secondPlayer := strings.TrimPrefix(parts[1], "@")

	// Extract the scores
	sets := parts[2:]

	// Initialize the set counters
	firstPlayerSetsWon, firstPlayerSetsLost, secondPlayerSetsWon, secondPlayerSetsLost := 0, 0, 0, 0

	// Loop through sets and calculate wins/losses
	for _, set := range sets {
		score := strings.Split(set, "-")

		if len(score) != 2 {
			http.Error(w, "Invalid score format", http.StatusBadRequest)
			return
		}

		firstPlayerScore, err := strconv.Atoi(score[0])
		if err != nil {
			http.Error(w, "Invalid score for first player", http.StatusBadRequest)
			return
		}

		secondPlayerScore, err := strconv.Atoi(score[1])
		if err != nil {
			http.Error(w, "Invalid score for second player", http.StatusBadRequest)
			return
		}

		// Determine who won the set
		if firstPlayerScore > secondPlayerScore {
			firstPlayerSetsWon++
			secondPlayerSetsLost++
		} else {
			firstPlayerSetsLost++
			secondPlayerSetsWon++
		}
	}

	// Determine the game result
	gameResult := getGameResult(firstPlayerSetsWon, secondPlayerSetsWon)

	// Update or insert the first player's stats
	_, err := s.db.Exec(queryInsertUpdateUser, firstPlayer,
		gameResult.firstPlayerGamesWon,
		gameResult.firstPlayerGamesLost,
		gameResult.firstPlayerGamesDrawn,
		firstPlayerSetsWon,
		firstPlayerSetsLost)

	if err != nil {
		http.Error(w, "Error updating player stats", http.StatusInternalServerError)
		return
	}

	// Update or insert the second player's stats
	_, err = s.db.Exec(queryInsertUpdateUser, secondPlayer,
		gameResult.secondPlayerGamesWon,
		gameResult.secondPlayerGamesLost,
		gameResult.secondPlayerGamesDrawn,
		secondPlayerSetsWon,
		secondPlayerSetsLost)

	if err != nil {
		http.Error(w, "Error updating player stats", http.StatusInternalServerError)
		return
	}

	// Send success response to Slack
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command processed successfully!"))
}

func (s *Server) showLeaderboard(w http.ResponseWriter, r *http.Request) {

}

func getGameResult(firstPlayerSets, secondPlayerSets int) GameResults {
	switch {
	case firstPlayerSets > secondPlayerSets:
		return GameResults{1, 0, 0, 0, 1, 0}
	case firstPlayerSets < secondPlayerSets:
		return GameResults{0, 1, 0, 1, 0, 0}
	default:
		return GameResults{0, 0, 1, 0, 0, 1}
	}

}
