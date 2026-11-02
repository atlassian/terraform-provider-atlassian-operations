// Package agg holds the request and response types of the Teams GraphQL API (Atlassian GraphQL Gateway).
// Every team operation lives under the `team` namespace: query { team { teamV2 } } and
// mutation { team { createTeam | updateTeam | deleteTeam | addMembers | removeMembers } }.
package agg

const (
	// TeamStateV2
	TeamStateActive    TeamState = "ACTIVE"
	TeamStateDisbanded TeamState = "DISBANDED"
	TeamStatePurged    TeamState = "PURGED"

	// TeamPermissions: what the caller may do with the team. There is no separate add/remove members permission;
	// managing members falls under CAN_MANAGE_TEAM.
	PermissionViewTeam   Permission = "CAN_VIEW_TEAM"
	PermissionManageTeam Permission = "CAN_MANAGE_TEAM"
	PermissionDeleteTeam Permission = "CAN_DELETE_TEAM"
)

type (
	// MembershipSettings is the API's TeamMembershipSettings: who can join the team (OPEN or MEMBER_INVITE for teams
	// the provider creates).
	MembershipSettings string
	TeamState          string
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
	}

	// UpdateTeamInput is TeamUpdateTeamInput. TeamId is the team ARI (ari:cloud:identity::team/<teamId>). The input
	// has no membershipSettings: the team type cannot be changed after creation through the Teams GraphQL API, and
	// the team resource refuses an update that changes team_type.
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

	// Member is TeamMemberV2. The members query asks for the account ID only and filters on FULL_MEMBER itself.
	Member struct {
		Member struct {
			AccountId string `json:"accountId"`
		} `json:"member"`
	}

	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	}

	// MutationPayload is the payload of every team mutation (TeamCreatePayload, TeamUpdatePayload, TeamDeletePayload,
	// TeamAddMembersPayload, TeamRemoveMembersPayload). createTeam and updateTeam return the team, with its permissions
	// when they are requested, so no follow-up read is needed; deleteTeam returns it in state PURGED. FailedMembers
	// exists only on the member mutations: a member that could not be added or removed is reported there, with
	// success=false and an empty errors list.
	MutationPayload struct {
		Team          *Team          `json:"team,omitempty"`
		Success       bool           `json:"success"`
		Errors        []Error        `json:"errors"`
		FailedMembers []FailedMember `json:"failedMembers,omitempty"`
	}

	// FailedMember is TeamMembershipMutationError: one member that addMembers or removeMembers rejected.
	FailedMember struct {
		AccountId string `json:"accountId"`
		Code      string `json:"code"`
		Message   string `json:"message"`
	}

	// ErrorExtensions carries a GraphQL error's extensions. Its two homes set different fields: a top-level error,
	// such as a deleted team read through teamV2 (teamV2: null, path [team teamV2]), has Classification Gone, or
	// NotFound for an unknown team. An error inside a mutation payload (MutationError) has StatusCode and ErrorType
	// instead: 410 Gone or 404 NotFound when deleteTeam or updateTeam name a deleted or unknown team.
	ErrorExtensions struct {
		StatusCode     int    `json:"statusCode,omitempty"`
		Classification string `json:"classification,omitempty"`
		ErrorType      string `json:"errorType,omitempty"`
	}

	// Error is a GraphQL error.
	Error struct {
		Message    string          `json:"message"`
		Path       []any           `json:"path,omitempty"`
		Extensions ErrorExtensions `json:"extensions"`
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
