package dto

type CriteriaDto struct {
	Type       CriteriaType            `json:"type"`
	Conditions *[]CriteriaConditionDto `json:"conditions,omitempty"`
}

type CriteriaConditionDto struct {
	Field         string `json:"field"`
	Operation     string `json:"operation"`
	ExpectedValue string `json:"expectedValue,omitempty"`
	Key           string `json:"key,omitempty"`
	Not           bool   `json:"not,omitempty"`
	Order         int    `json:"order,omitempty"`
}
