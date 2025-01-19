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

type MatchResult struct {
	Winner *Player
	Loser  *Player
	Games  []string
}
