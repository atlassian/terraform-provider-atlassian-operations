package dto

type RotationType string
type ResponderType string

const (
	Weekly     RotationType  = "weekly"
	Daily      RotationType  = "daily"
	Hourly     RotationType  = "hourly"
	User       ResponderType = "user"
	Team       ResponderType = "team"
	Escalation ResponderType = "escalation"
	noone      ResponderType = "noone"
)

type ResponderInfo struct {
	Id   *string
	Type ResponderType
}

type Rotation struct {
	Id              string           `json:"id"`
	Name            string           `json:"name"`
	StartDate       string           `json:"startDate"`
	EndDate         string           `json:"endDate"`
	Type            RotationType     `json:"type"`
	Length          int32            `json:"length"`
	Participants    []ResponderInfo  `json:"participants"`
	TimeRestriction *TimeRestriction `json:"timeRestriction"`
}
