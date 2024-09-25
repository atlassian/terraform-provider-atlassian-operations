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
	Id              string
	Name            string
	StartDate       string
	EndDate         string
	Type            RotationType
	Length          int32
	Participants    []ResponderInfo
	TimeRestriction *TimeRestriction
}
