package rest

import (
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
	"github.com/6ixfigs/pingypongy/internal/db"
	"github.com/6ixfigs/pingypongy/internal/pong"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Server struct {
	Router chi.Router
	Config *config.Config
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
		pong:   pong.New(db),
	}, nil
}

func (s *Server) MountRoutes() {
	s.Router.Post("/leaderboards", s.createLeaderboard)
	s.Router.Get("/leaderboards/{leaderboard_name}", s.getLeaderboard)

	s.Router.Post("/leaderboards/{leaderboard_name}/webhooks", s.registerWebhook)
	s.Router.Get("/leaderboards/{leaderboard_name}/webhooks", s.registerWebhook)
	s.Router.Delete("/leaderboards/{leaderboard_name}/webhooks", s.deleteWebhooks)

	s.Router.Post("/leaderbaords/{leaderboard_name}/players", s.createPlayer)
	s.Router.Get("/leaderboards/{leaderboard_name}/players/{username}", s.getPlayerStats)

	s.Router.Post("/leaderboards/{leaderboard_name}/matches", s.recordMatch)
}

func (s *Server) createLeaderboard(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) registerWebhook(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) listWebhooks(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) deleteWebhooks(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) createPlayer(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) recordMatch(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) getPlayerStats(w http.ResponseWriter, r *http.Request) {

}

func formatMatchResult(result *pong.MatchResult) string {
	players := fmt.Sprintf("<@%s> vs <@%s>", result.P1.UserID, result.P2.UserID)

	var gameResults string
	for i, g := range result.Games {
		gameResults += fmt.Sprintf("- Game %d: %d-%d\n", i+1, g.P1PointsWon, g.P2PointsWon)
	}

	var conclusion string
	if result.IsDraw {
		conclusion = "Draw!"
	} else {
		conclusion = fmt.Sprintf(":trophy: Winner: <@%s> %d-%d",
			result.Winner.UserID,
			result.P1GamesWon,
			result.P2GamesWon,
		)
	}

	response := fmt.Sprintf("Match recorded:\n%s\n%s\n%s", players, gameResults, conclusion)
	return response
}

func formatStats(player *pong.Player) string {
	r := `Stats for <@%s>:
	- Matches played: %d
	- Matches won: %d
	- Matches lost: %d
	- Matches drawn: %d
	- Games won: %d
	- Games lost: %d
	- Points won: %d
	- Win ratio: %.2f%%
	- Current streak: %d
	- Elo: %d
	`

	matchesPlayed := player.MatchesWon + player.MatchesLost + player.MatchesDrawn
	return fmt.Sprintf(
		r,
		player.UserID,
		matchesPlayed,
		player.MatchesWon,
		player.MatchesLost,
		player.MatchesDrawn,
		player.TotalGamesWon,
		player.TotalGamesLost,
		player.TotalPointsWon,
		float64(player.MatchesWon)/float64(matchesPlayed)*100,
		player.CurrentStreak,
		player.Elo,
	)
}

func formatLeaderboard(leaderboard []pong.Player) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "player", "W", "D", "L", "P", "Win Ratio", "Elo"})
	for rank, player := range leaderboard {
		matchesPlayed := player.MatchesWon + player.MatchesDrawn + player.MatchesLost
		t.AppendRow(table.Row{
			rank + 1,
			player.FullName,
			player.MatchesWon,
			player.MatchesDrawn,
			player.MatchesLost,
			matchesPlayed,
			fmt.Sprintf("%.2f", float64(player.MatchesWon)/float64(matchesPlayed)*100),
			player.Elo,
		})
	}
	text := fmt.Sprintf(":table_tennis_paddle_and_ball: *Current Leaderboard*:\n```%s```", t.Render())

	return text
}
