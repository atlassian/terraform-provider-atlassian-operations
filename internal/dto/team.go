package dto

const (
	OPEN          teamType       = "OPEN"
	MEMBER_INVITE teamType       = "MEMBER_INVITE"
	EXTERNAL      teamType       = "EXTERNAL"
	ADMIN         teamMemberRole = "ADMIN"
	USER          teamMemberRole = "USER"
)

type (
	teamType                 string
	teamMemberRole           string
	PublicApiUserPermissions struct {
		AddMembers    bool `json:"ADD_MEMBERS"`
		DeleteTeam    bool `json:"DELETE_TEAM"`
		RemoveMembers bool `json:"REMOVE_MEMBERS"`
		UpdateTeam    bool `json:"UPDATE_TEAM"`
	}
	TeamDto struct {
		Description     string                   `json:"description"`
		DisplayName     string                   `json:"display_name"`
		OrganizationId  string                   `json:"organizationId"`
		TeamId          string                   `json:"teamId"`
		TeamType        teamType                 `json:"teamType"`
		UserPermissions PublicApiUserPermissions `json:"userPermissions"`
		Member          []TeamMemberDto          `json:"member"`
	}
	TeamMemberDto struct {
		Id       string         `json:"id"`
		Username string         `json:"username"`
		Role     teamMemberRole `json:"role"`
	}
)
