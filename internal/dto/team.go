package dto

const (
	OPEN          TeamType = "OPEN"
	MEMBER_INVITE TeamType = "MEMBER_INVITE"
	EXTERNAL      TeamType = "EXTERNAL"
)

type (
	TeamType                 string
	PublicApiUserPermissions struct {
		AddMembers    bool `json:"ADD_MEMBERS"`
		DeleteTeam    bool `json:"DELETE_TEAM"`
		RemoveMembers bool `json:"REMOVE_MEMBERS"`
		UpdateTeam    bool `json:"UPDATE_TEAM"`
	}
	TeamDto struct {
		Description     string                    `json:"description"`
		DisplayName     string                    `json:"displayName"`
		SiteId          *string                   `json:"siteId"`
		OrganizationId  string                    `json:"organizationId"`
		TeamId          string                    `json:"teamId"`
		TeamType        TeamType                  `json:"teamType"`
		UserPermissions *PublicApiUserPermissions `json:"userPermissions"`
	}
	TeamMember struct {
		AccountId string `json:"accountId"`
	}
	TeamMemberList struct {
		Members []TeamMember `json:"members"`
	}
	TeamEnableOps struct {
		TeamId          string   `json:"platformTeamId"`
		AdminAccountIds []string `json:"adminAccountIds"`
		InviteUsernames []string `json:"inviteUsernames"`
	}
)
