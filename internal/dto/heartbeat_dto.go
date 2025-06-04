package dto

type HeartbeatDto struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	Interval      int      `json:"interval"`
	IntervalUnit  string   `json:"intervalUnit"`
	Enabled       bool     `json:"enabled"`
	Status        string   `json:"status,omitempty"`
	OwnerTeamId   string   `json:"ownerTeamId,omitempty"`
	AlertMessage  string   `json:"alertMessage,omitempty"`
	AlertTags     []string `json:"alertTags,omitempty"`
	AlertPriority string   `json:"alertPriority,omitempty"`
}

type HeartbeatPaginatedResponseDto struct {
	Values []HeartbeatDto `json:"values"`
	Links  struct {
		Next string `json:"next,omitempty"`
	} `json:"links"`
}
