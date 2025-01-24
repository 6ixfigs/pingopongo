package types

type Player struct {
	Id            int
	UserID        string
	ChannelID     string
	TeamID        string
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
