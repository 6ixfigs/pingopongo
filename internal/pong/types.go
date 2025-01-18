package pong

type Player struct {
	id            int
	UserID        string
	channelID     string
	teamID        string
	matchesWon    int
	matchesDrawn  int
	matchesLost   int
	currentStreak int
	GamesWon      int
	gamesLost     int
	pointsWon     int
}

type MatchResult struct {
	Winner *Player
	Loser  *Player
	Games  []string
}
