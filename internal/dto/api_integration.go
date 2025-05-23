package dto

type (
	ApiIntegration struct {
		Id                     string                 `json:"id"`
		Name                   string                 `json:"name"`
		ApiKey                 string                 `json:"apiKey"`
		Type                   string                 `json:"type"`
		Enabled                bool                   `json:"enabled"`
		TeamId                 string                 `json:"teamId"`
		Advanced               bool                   `json:"advanced,omitempty"`
		MaintenanceSources     []MaintenanceSource    `json:"maintenanceSources,omitempty"`
		Directions             []string               `json:"directions,omitempty"`
		Domains                []string               `json:"domains,omitempty"`
		TypeSpecificProperties map[string]interface{} `json:"typeSpecificProperties"`
	}
)
