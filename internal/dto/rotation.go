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
	Id   *string       `json:"id"`
	Type ResponderType `json:"type"`
}

type Rotation struct {
	Id              string           `json:"id"`
	Name            string           `json:"name,omitempty"`
	StartDate       string           `json:"startDate"`
	EndDate         string           `json:"endDate"`
	Type            RotationType     `json:"type"`
	Length          int32            `json:"length"`
	Participants    []ResponderInfo  `json:"participants"`
	TimeRestriction *TimeRestriction `json:"timeRestriction"`
}

func (r *ResponderInfo) Equal(other *ResponderInfo) bool {
	if r == nil || other == nil {
		return false
	}
	if (r.Id == nil && other.Id == nil) && r.Type == other.Type {
		return true
	} else if r.Id == nil || other.Id == nil {
		return false
	} else {
		return *r.Id == *other.Id && r.Type == other.Type
	}
}
