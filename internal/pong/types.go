package pong

type Player struct {
	id            int
	userID        string
	channelID     string
	matchesWon    int
	matchesDrawn  int
	matchesLost   int
	setsWon       int
	setsLost      int
	currentStreak int
	gamesWon      int
	gamesLost     int
	pointsWon     int
	pointsLost    int
}
