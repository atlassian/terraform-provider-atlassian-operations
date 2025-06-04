package dto

type IntegrationActionDto struct {
	ID                     string                 `json:"id,omitempty"`
	Type                   string                 `json:"type"`
	Name                   string                 `json:"name"`
	Domain                 string                 `json:"domain"`
	Direction              string                 `json:"direction"`
	GroupType              string                 `json:"groupType"`
	Filter                 *FilterDto             `json:"filter,omitempty"`
	TypeSpecificProperties map[string]interface{} `json:"typeSpecificProperties,omitempty"`
	FieldMappings          map[string]interface{} `json:"fieldMappings,omitempty"`
	ActionMapping          *ActionMappingDto      `json:"actionMapping,omitempty"`
	Enabled                *bool                  `json:"enabled"`
}

type FilterDto struct {
	ConditionsEmpty    bool                 `json:"conditionsEmpty"`
	ConditionMatchType string               `json:"conditionMatchType"`
	Conditions         []FilterConditionDto `json:"conditions"`
}

type FilterConditionDto struct {
	Field           string `json:"field"`
	Operation       string `json:"operation"`
	ExpectedValue   string `json:"expectedValue"`
	Key             string `json:"key"`
	Not             bool   `json:"not"`
	Order           int64  `json:"order"`
	SystemCondition bool   `json:"systemCondition"`
}

type ActionMappingDto struct {
	Type      string                 `json:"type"`
	Parameter map[string]interface{} `json:"parameter,omitempty"`
}
