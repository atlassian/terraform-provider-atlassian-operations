package dto

type NotificationRuleDto struct {
	ID               string                  `json:"id,omitempty"`
	Name             string                  `json:"name"`
	ActionType       string                  `json:"actionType"`
	NotificationTime []string                `json:"notificationTime,omitempty"`
	TimeRestriction  *TimeRestriction        `json:"timeRestriction,omitempty"`
	Schedules        []string                `json:"schedules,omitempty"`
	Order            int                     `json:"order"`
	Steps            []NotificationRuleStep  `json:"steps,omitempty"`
	Repeat           *NotificationRuleRepeat `json:"repeat,omitempty"`
	Enabled          bool                    `json:"enabled"`
	Criteria         *CriteriaDto            `json:"criteria"`
}

type NotificationRuleStep struct {
	SendAfter int                 `json:"sendAfter"`
	Contact   NotificationContact `json:"contact"`
	Enabled   bool                `json:"enabled"`
}

type NotificationContact struct {
	Method string `json:"method"`
	To     string `json:"to"`
}

type NotificationRuleRepeat struct {
	LoopAfter int  `json:"loopAfter"`
	Enabled   bool `json:"enabled"`
}
