package models

type Leaderboard struct {
	ID        int
	Name      string
	CreatedAt string
}

type Player struct {
	ID             int
	LeaderboardID  int
	Username       string
	MatchesWon     int
	MatchesDrawn   int
	MatchesLost    int
	TotalGamesWon  int
	TotalGamesLost int
	CurrentStreak  int
	Elo            int
	CreatedAt      string
}

type MatchScore struct {
	P1 int
	P2 int
}

type MatchResult struct {
	P1        *Player
	P2        *Player
	P1EloDiff int
	P2EloDiff int
	Score     *MatchScore
}
