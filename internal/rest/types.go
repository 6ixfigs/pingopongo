package rest

type SlackRequest struct {
	teamID         string
	teamDomain     string
	enterpriseID   string
	enterpriseName string
	channelID      string
	channelName    string
	userID         string
	command        string
	text           string
	responseUrl    string
	triggerID      string
	apiAppID       string
}

type SlackResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

type Player struct {
	id            int
	userID        string
	fullName      string
	channelID     string
	matchesWon    int
	matchesDrawn  int
	matchesLost   int
	setsWon       int
	setsLost      int
	currentStreak int
}
