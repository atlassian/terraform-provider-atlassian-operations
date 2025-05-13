package dto

type AlertPolicyDto struct {
	ID                     string                 `json:"id,omitempty"`
	Type                   string                 `json:"type"`
	Name                   string                 `json:"name"`
	Description            string                 `json:"description,omitempty"`
	TeamID                 string                 `json:"teamId,omitempty"`
	Enabled                bool                   `json:"enabled"`
	Filter                 *AlertFilterDto        `json:"filter,omitempty"`
	TimeRestriction        *TimeRestrictionDto    `json:"timeRestriction,omitempty"`
	Alias                  string                 `json:"alias,omitempty"`
	Message                string                 `json:"message"`
	AlertDescription       string                 `json:"alertDescription,omitempty"`
	Source                 string                 `json:"source,omitempty"`
	Entity                 string                 `json:"entity,omitempty"`
	Responders             []ResponderDto         `json:"responders,omitempty"`
	Actions                []string               `json:"actions,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Details                map[string]interface{} `json:"details,omitempty"`
	Continue               bool                   `json:"continue"`
	UpdatePriority         bool                   `json:"updatePriority"`
	PriorityValue          string                 `json:"priorityValue,omitempty"`
	KeepOriginalResponders bool                   `json:"keepOriginalResponders"`
	KeepOriginalDetails    bool                   `json:"keepOriginalDetails"`
	KeepOriginalActions    bool                   `json:"keepOriginalActions"`
	KeepOriginalTags       bool                   `json:"keepOriginalTags"`
}

type AlertFilterDto struct {
	Type       string              `json:"type"`
	Conditions []AlertConditionDto `json:"conditions"`
}

type AlertConditionDto struct {
	Field         string `json:"field"`
	Key           string `json:"key,omitempty"`
	Not           bool   `json:"not,omitempty"`
	Operation     string `json:"operation"`
	ExpectedValue string `json:"expectedValue"`
	Order         int    `json:"order,omitempty"`
}

type TimeRestrictionDto struct {
	Enabled          bool                       `json:"enabled"`
	TimeRestrictions []TimeRestrictionPeriodDto `json:"timeRestrictions"`
}

type TimeRestrictionPeriodDto struct {
	StartHour   int `json:"startHour"`
	StartMinute int `json:"startMinute"`
	EndHour     int `json:"endHour"`
	EndMinute   int `json:"endMinute"`
}

type ResponderDto struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

type ActionDto struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}
