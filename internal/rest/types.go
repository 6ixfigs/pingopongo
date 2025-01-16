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

type EventRequest struct {
	Token          string            `json:"token"`
	Challenge      string            `json:"challenge"`
	TeamID         string            `json:"team_id"`
	ApiAppID       string            `json:"api_app_id"`
	Event          map[string]string `json:"event"`
	Type           string            `json:"type"`
	Authorizations interface{}       `json:"authorizations"`
	EventContext   string            `json:"event_context"`
	EventID        string            `json:"event_id"`
	EventTime      string            `json:"event_time"`
}
