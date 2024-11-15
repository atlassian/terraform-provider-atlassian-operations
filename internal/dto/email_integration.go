package dto

type (
	TypeSpecificPropertiesDto struct {
		EmailUsername         string `json:"emailUsername"`
		SuppressNotifications bool   `json:"suppressNotifications"`
	}

	MaintenanceInterval struct {
		StartTimeMillis int32 `json:"startTimeMillis"`
		EndTimeMillis   int32 `json:"endTimeMillis"`
	}
	MaintenanceSource struct {
		MaintenanceId string              `json:"maintenanceId"`
		Enabled       bool                `json:"enabled"`
		Interval      MaintenanceInterval `json:"interval"`
	}
	EmailIntegration struct {
		Id                     string                    `json:"id"`
		Name                   string                    `json:"name"`
		Type                   string                    `json:"type"`
		TeamId                 string                    `json:"teamId"`
		Enabled                bool                      `json:"enabled"`
		Advanced               bool                      `json:"advanced,omitempty"`
		MaintenanceSources     []MaintenanceSource       `json:"maintenance_sources,omitempty"`
		Directions             []string                  `json:"directions,omitempty"`
		Domains                []string                  `json:"domains,omitempty"`
		TypeSpecificProperties TypeSpecificPropertiesDto `json:"typeSpecificProperties"`
	}
)
