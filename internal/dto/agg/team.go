// Package agg holds the request and response types of the Teams GraphQL API (Atlassian GraphQL Gateway).
// Every team operation lives under the `team` namespace: query { team { teamV2 } } and
// mutation { team { createTeam | updateTeam | deleteTeam | addMembers | removeMembers } }.
package agg

const (
	// TeamMembershipSettings: who can join the team. Maps to the REST teamType.
	MembershipOpen            MembershipSettings = "OPEN"
	MembershipMemberInvite    MembershipSettings = "MEMBER_INVITE"
	MembershipExternal        MembershipSettings = "EXTERNAL"
	MembershipOrgAdminManaged MembershipSettings = "ORG_ADMIN_MANAGED"

	// TeamStateV2
	TeamStateActive    TeamState = "ACTIVE"
	TeamStateDisbanded TeamState = "DISBANDED"
	TeamStatePurged    TeamState = "PURGED"

	// TeamMembershipState. The default members list only contains FULL_MEMBER; a removed member becomes ALUMNI.
	MemberStateFullMember       MembershipState = "FULL_MEMBER"
	MemberStateAlumni           MembershipState = "ALUMNI"
	MemberStateRequestingToJoin MembershipState = "REQUESTING_TO_JOIN"

	// TeamMembershipRole
	MemberRoleRegular MembershipRole = "REGULAR"
	MemberRoleAdmin   MembershipRole = "ADMIN"

	// TeamPermissions: what the caller may do with the team. There is no separate add/remove members permission;
	// managing members falls under CAN_MANAGE_TEAM.
	PermissionViewTeam   Permission = "CAN_VIEW_TEAM"
	PermissionManageTeam Permission = "CAN_MANAGE_TEAM"
	PermissionDeleteTeam Permission = "CAN_DELETE_TEAM"

	// Member __typename values: AppUser is a service account, AtlassianAccountUser is a person.
	UserTypeApp              = "AppUser"
	UserTypeAtlassianAccount = "AtlassianAccountUser"
)

type (
	MembershipSettings string
	TeamState          string
	MembershipState    string
	MembershipRole     string
	Permission         string

	// Request is the GraphQL request body.
	Request struct {
		Query         string         `json:"query"`
		OperationName string         `json:"operationName,omitempty"`
		Variables     map[string]any `json:"variables,omitempty"`
	}

	// CreateTeamInput is TeamCreateTeamInput. Members are user ARIs (ari:cloud:identity::user/<accountId>).
	CreateTeamInput struct {
		DisplayName        string             `json:"displayName"`
		Description        string             `json:"description"`
		MembershipSettings MembershipSettings `json:"membershipSettings"`
		ScopeId            string             `json:"scopeId"`
		Members            []string           `json:"members"`
		TypeId             string             `json:"typeId,omitempty"`
	}

	// UpdateTeamInput is TeamUpdateTeamInput. TeamId is the team ARI (ari:cloud:identity::team/<teamId>).
	UpdateTeamInput struct {
		TeamId      string `json:"teamId"`
		ScopeId     string `json:"scopeId"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}

	// DeleteTeamInput is TeamDeleteTeamInput. TeamId is the team ARI.
	DeleteTeamInput struct {
		TeamId  string `json:"teamId"`
		ScopeId string `json:"scopeId"`
	}

	// MembersInput is TeamAddMembersInput and TeamRemoveMembersInput. TeamId is the team ARI;
	// unlike createTeam, MemberIds are raw account IDs.
	MembersInput struct {
		TeamId    string   `json:"teamId"`
		ScopeId   string   `json:"scopeId"`
		MemberIds []string `json:"memberIds"`
	}

	// Team is TeamV2. Id is the team ARI. OrganizationId is a platform org ARI that does not match the
	// organization ID used by the REST API, so it is not mapped.
	Team struct {
		Id                 string             `json:"id"`
		DisplayName        string             `json:"displayName"`
		Description        string             `json:"description"`
		MembershipSettings MembershipSettings `json:"membershipSettings"`
		State              TeamState          `json:"state"`
		Permissions        []Permission       `json:"permissions,omitempty"`
		Members            *MemberConnection  `json:"members,omitempty"`
	}

	// MemberConnection is TeamMemberConnectionV2, paged with members(first, after, state).
	MemberConnection struct {
		Nodes    []Member `json:"nodes"`
		PageInfo PageInfo `json:"pageInfo"`
	}

	// Member is TeamMemberV2.
	Member struct {
		Member struct {
			Typename  string `json:"__typename,omitempty"`
			AccountId string `json:"accountId"`
		} `json:"member"`
		State MembershipState `json:"state,omitempty"`
		Role  MembershipRole  `json:"role,omitempty"`
	}

	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	}

	// MutationPayload is the payload of every team mutation (TeamCreatePayload, TeamUpdatePayload, ...).
	// Team is only returned by createTeam and updateTeam.
	MutationPayload struct {
		Team    *Team   `json:"team,omitempty"`
		Success bool    `json:"success"`
		Errors  []Error `json:"errors"`
	}

	// Error is a GraphQL error, either top level or inside a mutation payload. A deleted team comes back as
	// teamV2: null with an error on path [team teamV2], classification Gone and errorCode TEAM_DELETED;
	// an unknown team as classification NotFound.
	Error struct {
		Message    string `json:"message"`
		Path       []any  `json:"path,omitempty"`
		Extensions struct {
			StatusCode     int    `json:"statusCode,omitempty"`
			Classification string `json:"classification,omitempty"`
			ErrorCode      string `json:"errorCode,omitempty"`
		} `json:"extensions"`
	}

	// TeamNamespace is the `team` field of the query and mutation roots; only the requested operation is set.
	TeamNamespace struct {
		TeamV2        *Team            `json:"teamV2,omitempty"`
		CreateTeam    *MutationPayload `json:"createTeam,omitempty"`
		UpdateTeam    *MutationPayload `json:"updateTeam,omitempty"`
		DeleteTeam    *MutationPayload `json:"deleteTeam,omitempty"`
		AddMembers    *MutationPayload `json:"addMembers,omitempty"`
		RemoveMembers *MutationPayload `json:"removeMembers,omitempty"`
	}

	// Response is the GraphQL response body. Top-level errors can be non-fatal warnings, so check the payload too.
	Response struct {
		Data struct {
			Team TeamNamespace `json:"team"`
		} `json:"data"`
		Errors []Error `json:"errors,omitempty"`
	}
)
