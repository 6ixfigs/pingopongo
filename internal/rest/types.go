package rest

type CommandRequest struct {
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

type CommandResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}
