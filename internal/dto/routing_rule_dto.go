package dto

type CriteriaType string

const (
	MatchAll           CriteriaType = "match-all"
	MatchAllConditions CriteriaType = "match-all-conditions"
	MatchAnyConditions CriteriaType = "match-any-condition"
)

type RoutingRuleDto struct {
	ID              string                `json:"id,omitempty"`
	Name            string                `json:"name,omitempty"`
	Order           int64                 `json:"order,omitempty"`
	IsDefault       bool                  `json:"isDefault,omitempty"`
	Timezone        string                `json:"timezone,omitempty"`
	Criteria        *CriteriaDto          `json:"criteria,omitempty"`
	TimeRestriction *TimeRestriction      `json:"timeRestriction,omitempty"`
	Notify          *RoutingRuleNotifyDto `json:"notify"`
}

type ListRoutingRuleDto struct {
	Values []RoutingRuleDto `json:"values"`
	Links  LinksDto         `json:"links"`
}

type LinksDto struct {
	Next string `json:"next"`
}

type TimeRestrictionEntry struct {
	StartHour int    `json:"startHour"`
	EndHour   int    `json:"endHour"`
	StartMin  int    `json:"startMin"`
	EndMin    int    `json:"endMin"`
	StartDay  string `json:"startDay,omitempty"`
	EndDay    string `json:"endDay,omitempty"`
}

type RoutingRuleNotifyDto struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
