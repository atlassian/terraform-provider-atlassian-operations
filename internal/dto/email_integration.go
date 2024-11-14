package dto

type (
	TypeSpecificPropertiesDto struct {
		EmailUsername         string `json:"emailUsername"`
		SuppressNotifications bool   `json:"suppressNotifications"`
	}

	EmailIntegration struct {
		Id                     string                    `json:"id"`
		Name                   string                    `json:"name"`
		Type                   string                    `json:"type"`
		TeamId                 string                    `json:"teamId"`
		Enabled                bool                      `json:"enabled"`
		TypeSpecificProperties TypeSpecificPropertiesDto `json:"typeSpecificProperties"`
	}
)
