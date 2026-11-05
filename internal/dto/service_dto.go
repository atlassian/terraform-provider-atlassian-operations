package dto

type ServiceDto struct {
	ID              string              `json:"id,omitempty"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Tier            int32               `json:"tier"`
	Type            string              `json:"type"`
	Owner           string              `json:"owner,omitempty"`
	ChangeApprovers *ChangeApproversDto `json:"changeApprovers,omitempty"`
	Responders      *RespondersDto      `json:"responders,omitempty"`
	Stakeholders    *StakeholdersDto    `json:"stakeholders,omitempty"`
	Projects        *ProjectsDto        `json:"projects,omitempty"`
}

type ChangeApproversDto struct {
	Groups []string `json:"groups,omitempty"`
}

type RespondersDto struct {
	Users []string `json:"users,omitempty"`
	Teams []string `json:"teams,omitempty"`
}

type StakeholdersDto struct {
	Users []string `json:"users,omitempty"`
}

type ProjectsDto struct {
	IDs []string `json:"ids,omitempty"`
}
