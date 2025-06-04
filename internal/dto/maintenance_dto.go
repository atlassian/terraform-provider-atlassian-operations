package dto

// MaintenanceDto represents a maintenance window in the Atlassian API
type MaintenanceDto struct {
	ID          string               `json:"id,omitempty"`
	Status      string               `json:"status,omitempty"`
	Description string               `json:"description,omitempty"`
	StartDate   string               `json:"startDate,omitempty"`
	EndDate     string               `json:"endDate,omitempty"`
	TeamID      string               `json:"teamId,omitempty"`
	Rules       []MaintenanceRuleDto `json:"rules,omitempty"`
}

// MaintenanceRuleDto represents a rule within a maintenance window
type MaintenanceRuleDto struct {
	State  string                   `json:"state,omitempty"`
	Entity MaintenanceRuleEntityDto `json:"entity,omitempty"`
}

// MaintenanceRuleEntityDto represents an entity affected by a maintenance rule
type MaintenanceRuleEntityDto struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}
