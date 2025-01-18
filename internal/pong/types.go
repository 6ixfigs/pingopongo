package pong

type Player struct {
	id            int
	UserID        string
	channelID     string
	teamID        string
	MatchesWon    int
	MatchesDrawn  int
	MatchesLost   int
	CurrentStreak int
	GamesWon      int
	GamesLost     int
	PointsWon     int
}

type MatchResult struct {
	Winner *Player
	Loser  *Player
	Games  []string
}
