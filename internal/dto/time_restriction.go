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
	StartHour int32
	EndHour   int32
	StartMin  int32
	EndMin    int32
}

type WeekdayTimeRestrictionSettings struct {
	StartDay  Weekday
	EndDay    Weekday
	StartHour int32
	EndHour   int32
	StartMin  int32
	EndMin    int32
}
