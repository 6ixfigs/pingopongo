package rest

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Server struct {
	Router chi.Router
	Config *config.Config
	db     *sql.DB
	pong   *pong.Pong
}

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
		Config: cfg,
		db:     db,
		pong:   pong.New(db),
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/leaderboards", s.createLeaderboard)
	s.Router.Get("/leaderboards/{leaderboard_name}", s.getLeaderboard)

	s.Router.Post("/leaderboards/{leaderboard_name}/webhooks", s.registerWebhook)
	s.Router.Get("/leaderboards/{leaderboard_name}/webhooks", s.listWebhooks)
	s.Router.Delete("/leaderboards/{leaderboard_name}/webhooks", s.deleteWebhooks)

	s.Router.Post("/leaderboards/{leaderboard_name}/players", s.createPlayer)
	s.Router.Get("/leaderboards/{leaderboard_name}/players/{username}", s.getPlayerStats)

	s.Router.Post("/leaderboards/{leaderboard_name}/matches", s.recordMatch)
}

func (s *Server) createLeaderboard(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaderboardName := r.FormValue("name")

	err := s.pong.CreateLeaderboard(leaderboardName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Leaderboard %s created!", leaderboardName)))
}

func (s *Server) registerWebhook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaderboardName := chi.URLParam(r, "leaderboard_name")
	url := r.FormValue("url")

	err := s.pong.RegisterWebhook(leaderboardName, url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Registered webhook %s on leaderboard %s", url, leaderboardName)))
}

func (s *Server) listWebhooks(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")

	webhooks, err := s.pong.ListWebhooks(leaderboardName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(strings.Join(webhooks, "\n")))
}

func (s *Server) deleteWebhooks(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")

	err := s.pong.DeleteWebhooks(leaderboardName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Deleted all webhooks in %s leaderboard", leaderboardName)))
}

func (s *Server) createPlayer(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaderboardName := chi.URLParam(r, "leaderboard_name")
	username := r.FormValue("username")

	err := s.pong.CreatePlayer(leaderboardName, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Created player %s on leaderboard %s", username, leaderboardName)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func (s *Server) recordMatch(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username1 := r.FormValue("player1")
	username2 := r.FormValue("player2")
	score := r.FormValue("score")

	result, err := s.pong.Record(leaderboardName, username1, username2, score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := formatMatchResult(result)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")

	rankings, err := s.pong.Leaderboard(leaderboardName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := formatLeaderboard(leaderboardName, rankings)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func (s *Server) getPlayerStats(w http.ResponseWriter, r *http.Request) {
	leaderboardName := chi.URLParam(r, "leaderboard_name")
	username := chi.URLParam(r, "username")

	player, err := s.pong.Stats(leaderboardName, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := formatStats(player)

	_, err = s.notifyWebhooks(leaderboardName, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(response))
}

func formatMatchResult(result *pong.MatchResult) string {
	return fmt.Sprintf("Match recorded: (%+d) %s %d - %d %s (%+d)!",
		result.P1EloDiff,
		result.P1.Username,
		result.Score.P1,
		result.Score.P2,
		result.P2.Username,
		result.P2EloDiff,
	)
}

func formatStats(player *pong.Player) string {
	r := `Stats for %s:
	- Matches played: %d
	- Matches won: %d
	- Matches lost: %d
	- Matches drawn: %d
	- Games won: %d
	- Games lost: %d
	- Win ratio: %.2f%%
	- Current streak: %d
	- Elo: %d
	`

	matchesPlayed := player.MatchesWon + player.MatchesLost + player.MatchesDrawn
	winRatio := 0.
	if matchesPlayed > 0 {
		winRatio = float64(player.MatchesWon) / float64(matchesPlayed) * 100
	}
	return fmt.Sprintf(
		r,
		player.Username,
		matchesPlayed,
		player.MatchesWon,
		player.MatchesLost,
		player.MatchesDrawn,
		player.TotalGamesWon,
		player.TotalGamesLost,
		winRatio,
		player.CurrentStreak,
		player.Elo,
	)
}

func formatLeaderboard(leaderboardName string, leaderboard []pong.Player) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio", "Elo"})
	for rank, player := range leaderboard {
		matchesPlayed := player.MatchesWon + player.MatchesDrawn + player.MatchesLost
		winRatio := 0.
		if matchesPlayed > 0 {
			winRatio = float64(player.MatchesWon) / float64(matchesPlayed) * 100
		}
		t.AppendRow(table.Row{
			rank + 1,
			player.Username,
			player.MatchesWon,
			player.MatchesDrawn,
			player.MatchesLost,
			matchesPlayed,
			fmt.Sprintf("%.2f%%", winRatio),
			player.Elo,
		})
	}
	text := fmt.Sprintf("%s leaderboard:\n```\n%s\n```", leaderboardName, t.Render())

	return text
}

func (s *Server) notifyWebhooks(leaderboardName, text string) ([]string, error) {
	webhooks, err := s.pong.ListWebhooks(leaderboardName)
	if err != nil {
		return nil, err
	}

	var failed []string
	payload := []byte(fmt.Sprintf(`{"text": "%s"}`, text))
	for _, webhook := range webhooks {
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(payload))
		if err != nil {
			failed = append(failed, webhook)
			continue
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			failed = append(failed, webhook)
			continue
		}
	}

	return failed, nil
}
