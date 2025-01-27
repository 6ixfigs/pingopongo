package slack

type CommandRequest struct {
	TeamID         string
	TeamDomain     string
	EnterpriseID   string
	EnterpriseName string
	ChannelID      string
	ChannelName    string
	UserID         string
	Command        string
	Text           string
	ResponseUrl    string
	TriggerID      string
	ApiAppID       string
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

type UserInfoResponse struct {
	Ok    bool     `json:"ok"`
	Error string   `json:"error,omitempty"`
	User  UserInfo `json:"user,omitempty"`
}

type UserInfo struct {
	RealName string `json:"real_name"`
}
