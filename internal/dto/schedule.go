package dto

type (
	Schedule struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Timezone    string `json:"timezone"`
		Enabled     bool   `json:"enabled"`
		TeamId      string `json:"teamId"`
	}

	ListSchedule struct {
		Values []Schedule `json:"values"`
		Links  LinksDto   `json:"links"`
	}
)
