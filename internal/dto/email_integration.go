package dto

type (
	TypeSpecificPropertiesDto struct {
		EmailUsername         string `json:"emailUsername"`
		SuppressNotifications bool   `json:"suppressNotifications"`
	}

	MaintenanceInterval struct {
		StartTimeMillis int64 `json:"startTimeMillis"`
		EndTimeMillis   int64 `json:"endTimeMillis"`
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
		MaintenanceSources     []MaintenanceSource       `json:"maintenanceSources,omitempty"`
		Directions             []string                  `json:"directions,omitempty"`
		Domains                []string                  `json:"domains,omitempty"`
		TypeSpecificProperties TypeSpecificPropertiesDto `json:"typeSpecificProperties"`
	}
)
