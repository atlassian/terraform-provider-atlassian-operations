package dto

type TimeRestrictionType string
type Weekday string

const (
	TimeOfDay           TimeRestrictionType = "time-of-day"
	WeekdayAndTimeOfDay TimeRestrictionType = "weekday-and-time-of-day"
	Monday              Weekday             = "monday"
	Tuesday             Weekday             = "tuesday"
	Wednesday           Weekday             = "wednesday"
	Thursday            Weekday             = "thursday"
	Friday              Weekday             = "friday"
	Saturday            Weekday             = "saturday"
	Sunday              Weekday             = "sunday"
)

type TimeRestriction struct {
	Type                        TimeRestrictionType               `json:"type"`
	TimeOfDayRestriction        *TimeOfDayTimeRestrictionSettings `json:"restriction,omitempty"`
	WeekAndTimeOfDayRestriction *[]WeekdayTimeRestrictionSettings `json:"restrictions,omitempty"`
}

type TimeOfDayTimeRestrictionSettings struct {
	StartHour int32 `json:"startHour"`
	EndHour   int32 `json:"endHour"`
	StartMin  int32 `json:"startMin"`
	EndMin    int32 `json:"endMin"`
}

type WeekdayTimeRestrictionSettings struct {
	StartDay  Weekday `json:"startDay"`
	EndDay    Weekday `json:"endDay"`
	StartHour int32   `json:"startHour"`
	EndHour   int32   `json:"endHour"`
	StartMin  int32   `json:"startMin"`
	EndMin    int32   `json:"endMin"`
}
