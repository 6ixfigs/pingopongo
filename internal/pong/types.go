package pong

type Player struct {
	id            int
	UserID        string
	channelID     string
	teamID        string
	FullName      string
	MatchesWon    int
	MatchesDrawn  int
	MatchesLost   int
	GamesWon      int
	GamesLost     int
	PointsWon     int
	CurrentStreak int
	Elo           int
}

type GameResult struct {
	Winner   *Player
	P1Points int
	P2Points int
}

type MatchResult struct {
	Winner *Player
	Loser  *Player
	IsDraw bool
	Games  []GameResult
}
