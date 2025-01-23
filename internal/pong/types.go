package pong

type Player struct {
	id             int
	UserID         string
	channelID      string
	teamID         string
	FullName       string
	MatchesWon     int
	MatchesDrawn   int
	MatchesLost    int
	TotalGamesWon  int
	TotalGamesLost int
	TotalPointsWon int
	CurrentStreak  int
	Elo            int
}

type GameResult struct {
	Winner      *Player
	P1          *Player
	P2          *Player
	P1PointsWon int
	P2PointsWon int
}

type MatchResult struct {
	Winner     *Player
	P1         *Player
	P2         *Player
	P1GamesWon int
	P2GamesWon int
	IsDraw     bool
	Games      []GameResult
}
