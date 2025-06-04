package dto

type NotificationPolicyDto struct {
	ID                  string                                `json:"id,omitempty"`
	Type                string                                `json:"type"`
	Name                string                                `json:"name"`
	Description         string                                `json:"description,omitempty"`
	TeamID              string                                `json:"teamId,omitempty"`
	Enabled             bool                                  `json:"enabled"`
	Order               float64                               `json:"order,omitempty"`
	Filter              *NotificationFilterDto                `json:"filter,omitempty"`
	TimeRestriction     *NotificationPolicyTimeRestrictionDto `json:"timeRestriction,omitempty"`
	AutoRestartAction   *AutoRestartActionDto                 `json:"autoRestartAction,omitempty"`
	AutoCloseAction     *AutoCloseActionDto                   `json:"autoCloseAction,omitempty"`
	DeduplicationAction *DeduplicationActionDto               `json:"deduplicationAction,omitempty"`
	DelayAction         *DelayActionDto                       `json:"delayAction,omitempty"`
	Suppress            bool                                  `json:"suppress"`
}

type NotificationPolicyTimeRestrictionDto struct {
	Enabled          bool                                           `json:"enabled"`
	AppliedTimeZone  string                                         `json:"appliedTimeZone"`
	TimeRestrictions []NotificationPolicyTimeRestrictionSettingsDto `json:"timeRestrictions"`
}

type NotificationPolicyTimeRestrictionSettingsDto struct {
	StartHour   int `json:"startHour"`
	EndHour     int `json:"endHour"`
	StartMinute int `json:"startMinute"`
	EndMinute   int `json:"endMinute"`
}

type AutoRestartActionDto struct {
	WaitDuration   int    `json:"waitDuration,omitempty"`
	MaxRepeatCount int    `json:"maxRepeatCount"`
	DurationFormat string `json:"durationFormat,omitempty"`
}

type AutoCloseActionDto struct {
	WaitDuration   int    `json:"waitDuration,omitempty"`
	DurationFormat string `json:"durationFormat,omitempty"`
}

type DeduplicationActionDto struct {
	DeduplicationActionType string `json:"deduplicationActionType"`
	Frequency               int    `json:"frequency,omitempty"`
	CountValueLimit         int    `json:"countValueLimit,omitempty"`
	WaitDuration            int    `json:"waitDuration,omitempty"`
	DurationFormat          string `json:"durationFormat,omitempty"`
}

type DelayActionDto struct {
	DelayTime      *DelayTimeDto `json:"delayTime"`
	DelayOption    string        `json:"delayOption"`
	WaitDuration   int           `json:"waitDuration,omitempty"`
	DurationFormat string        `json:"durationFormat,omitempty"`
}

type DelayTimeDto struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
}

type NotificationFilterDto struct {
	Type       string                     `json:"type"`
	Conditions []NotificationConditionDto `json:"conditions"`
}

type NotificationConditionDto struct {
	Field         string `json:"field"`
	Key           string `json:"key,omitempty"`
	Not           bool   `json:"not,omitempty"`
	Operation     string `json:"operation"`
	ExpectedValue string `json:"expectedValue"`
	Order         int    `json:"order,omitempty"`
}
