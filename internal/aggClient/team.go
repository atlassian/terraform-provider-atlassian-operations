package aggClient

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto/agg"
)

// Every team operation lives under the `team` namespace field, and the @optIn directives go on that field.
const (
	teamV2Query = `query TeamV2($id: ID!, $siteId: String!) {
  team @optIn(to: "Team-crud") {
    teamV2(id: $id, siteId: $siteId) {
      id
      displayName
      description
      membershipSettings
      state
      permissions
    }
  }
}`

	teamMembersQuery = `query TeamMembers($id: ID!, $siteId: String!, $first: Int!, $after: String) {
  team @optIn(to: "Team-crud") {
    teamV2(id: $id, siteId: $siteId) {
      state
      members(first: $first, after: $after, state: FULL_MEMBER) {
        nodes { member { accountId } }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`

	// The mutations ask for the caller's permissions on the returned team (createTeam and updateTeam return them, so
	// no follow-up read is needed), for the status and type of every payload error (410 Gone / 404 NotFound name a
	// team that no longer exists) and, on the member mutations, for the members that were rejected.
	createTeamMutation = `mutation CreateTeam($input: TeamCreateTeamInput!) {
  team @optIn(to: "Team-crud") {
    createTeam(input: $input) {
      team { id displayName description membershipSettings state permissions }
      success
      errors { message extensions { statusCode errorType } }
    }
  }
}`

	updateTeamMutation = `mutation UpdateTeam($input: TeamUpdateTeamInput!) {
  team @optIn(to: "Team-crud") {
    updateTeam(input: $input) {
      team { id displayName description membershipSettings state permissions }
      success
      errors { message extensions { statusCode errorType } }
    }
  }
}`

	deleteTeamMutation = `mutation DeleteTeam($input: TeamDeleteTeamInput!) {
  team @optIn(to: "Team-crud") {
    deleteTeam(input: $input) {
      success
      errors { message extensions { statusCode errorType } }
    }
  }
}`

	addMembersMutation = `mutation AddMembers($input: TeamAddMembersInput!) {
  team @optIn(to: "Team") @optIn(to: "Team-members") {
    addMembers(input: $input) {
      success
      errors { message extensions { statusCode errorType } }
      failedMembers { accountId code message }
    }
  }
}`

	removeMembersMutation = `mutation RemoveMembers($input: TeamRemoveMembersInput!) {
  team @optIn(to: "Team") @optIn(to: "Team-members") {
    removeMembers(input: $input) {
      success
      errors { message extensions { statusCode errorType } }
      failedMembers { accountId code message }
    }
  }
}`

	teamARIPrefix = "ari:cloud:identity::team/"
)

// Configuration and state hold raw IDs (team ID, account ID, site/cloud ID), exactly as with the Teams REST API, so
// switching to a scoped API token needs no configuration change. The conversions to and from ARIs happen only here,
// and accept either form so an ARI is never prefixed twice.
const (
	userARIPrefix = "ari:cloud:identity::user/"
	siteARIPrefix = "ari:cloud:platform::site/"
)

// teamIdFromARI converts a team ARI (as returned by AGG) to the raw team ID; a raw ID is returned unchanged.
func teamIdFromARI(teamId string) string {
	return strings.TrimPrefix(teamId, teamARIPrefix)
}

// teamARI builds the team ARI that AGG mutations expect.
func teamARI(teamId string) string {
	return teamARIPrefix + teamIdFromARI(teamId)
}

func accountIdFromARI(accountId string) string {
	return strings.TrimPrefix(accountId, userARIPrefix)
}

// userARI builds the user ARI that createTeam expects; addMembers/removeMembers take raw account IDs.
func userARI(accountId string) string {
	return userARIPrefix + accountIdFromARI(accountId)
}

func rawAccountIds(accountIds []string) []string {
	raw := make([]string, len(accountIds))
	for i, id := range accountIds {
		raw[i] = accountIdFromARI(id)
	}
	return raw
}

// siteIdOrCloudId returns the raw site ID teamV2 expects, falling back to the provider's cloud_id.
func (c Client) siteIdOrCloudId(siteId string) string {
	if siteId == "" {
		siteId = c.config.GetCloudId()
	}
	return strings.TrimPrefix(siteId, siteARIPrefix)
}

// siteScope is the scopeId of every team mutation: the site ARI of the team's site, or of the provider's cloud_id.
func (c Client) siteScope(siteId string) string {
	return siteARIPrefix + c.siteIdOrCloudId(siteId)
}

// toTeamDto converts an AGG team. OrganizationId and SiteId come from the caller: AGG does not return the
// organization ID used by the REST API, and the site is part of the request.
func toTeamDto(t agg.Team, organizationId string, siteId *string, fallbackType dto.TeamType) dto.TeamDto {
	teamType := fallbackType
	if t.MembershipSettings != "" {
		teamType = dto.TeamType(t.MembershipSettings)
	}
	return dto.TeamDto{
		TeamId:          teamIdFromARI(t.Id),
		DisplayName:     t.DisplayName,
		Description:     t.Description,
		OrganizationId:  organizationId,
		SiteId:          siteId,
		TeamType:        teamType,
		UserPermissions: toUserPermissions(t.Permissions),
	}
}

// toUserPermissions maps the caller's AGG team permissions to the REST user_permissions shape. AGG has no separate
// add/remove members permission: managing members falls under CAN_MANAGE_TEAM. Returns nil when AGG sent none.
func toUserPermissions(permissions []agg.Permission) *dto.PublicApiUserPermissions {
	if permissions == nil {
		return nil
	}
	canManage := slices.Contains(permissions, agg.PermissionManageTeam)
	return &dto.PublicApiUserPermissions{
		UpdateTeam:    canManage,
		AddMembers:    canManage,
		RemoveMembers: canManage,
		DeleteTeam:    slices.Contains(permissions, agg.PermissionDeleteTeam),
	}
}

// CreateTeam creates a team with the given members (raw account IDs) and returns it with its new raw team ID and the
// caller's permissions, both taken from the mutation payload.
func (c Client) CreateTeam(ctx context.Context, teamDto dto.TeamDto, siteId string, memberAccountIds []string) (dto.TeamDto, error) {
	members := make([]string, len(memberAccountIds))
	for i, id := range memberAccountIds {
		members[i] = userARI(id)
	}
	input := agg.CreateTeamInput{
		DisplayName:        teamDto.DisplayName,
		Description:        teamDto.Description,
		MembershipSettings: agg.MembershipSettings(teamDto.TeamType),
		ScopeId:            c.siteScope(siteId),
		Members:            members,
	}

	resp, err := c.mutate(ctx, "CreateTeam", createTeamMutation, map[string]any{"input": input})
	if err != nil {
		if errors.Is(err, errRejected) {
			// The gateway answered, nothing was created.
			return dto.TeamDto{}, err
		}
		// The gateway did not answer, so it is unknown whether the team was created; CreateTeam is sent once on purpose.
		return dto.TeamDto{}, fmt.Errorf("%w; a team named %q may have been created and left outside Terraform state; check the Teams UI before retrying", err, teamDto.DisplayName)
	}
	payload := resp.Data.Team.CreateTeam
	if err = checkMutation("CreateTeam", payload, resp.Errors); err != nil {
		// The team can exist even when the mutation reports a failure (for example a member was rejected);
		// delete it so it is not left behind outside Terraform state. When that delete fails too, the error says
		// so and names the team, so the operator knows a team leaked and can import or delete it.
		if payload != nil && payload.Team != nil && payload.Team.Id != "" {
			createdId := teamIdFromARI(payload.Team.Id)
			// Ctrl-C or a deadline cancels ctx; the cleanup gets its own short budget so the team is not left behind.
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			defer cancel()
			if delErr := c.DeleteTeam(cleanupCtx, createdId, siteId); delErr != nil {
				return dto.TeamDto{}, fmt.Errorf("%w; the partially created team %s could not be deleted and is left behind outside Terraform state: %v", err, createdId, delErr)
			}
			return dto.TeamDto{}, fmt.Errorf("%w (the partially created team was deleted)", err)
		}
		// Otherwise the API's own error is returned, as on the classic REST path.
		return dto.TeamDto{}, err
	}
	if payload.Team == nil || payload.Team.Id == "" {
		return dto.TeamDto{}, fmt.Errorf("Teams GraphQL API CreateTeam returned no team ID")
	}
	return toTeamDto(*payload.Team, teamDto.OrganizationId, teamDto.SiteId, teamDto.TeamType), nil
}

// teamFromResponse returns the team of a teamV2 response. It returns ErrTeamNotFound only when teamV2 explicitly
// reports the team as missing or deleted, so an ambiguous response never makes Terraform drop a live team.
func teamFromResponse(resp agg.Response, teamId string) (*agg.Team, error) {
	team := resp.Data.Team.TeamV2
	if team == nil {
		for _, e := range resp.Errors {
			// teamV2 is the only field a wrong hit may be about: it removes the resource from state.
			if teamGone([]agg.Error{e}) && len(e.Path) > 0 && fmt.Sprint(e.Path[len(e.Path)-1]) == "teamV2" {
				return nil, ErrTeamNotFound
			}
		}
		if len(resp.Errors) > 0 {
			return nil, fmt.Errorf("Teams GraphQL API TeamV2 errors: %s", errorMessages(resp.Errors))
		}
		return nil, fmt.Errorf("Teams GraphQL API TeamV2 returned no team %s", teamId)
	}
	if team.State == agg.TeamStateDisbanded || team.State == agg.TeamStatePurged {
		return nil, ErrTeamNotFound
	}
	return team, nil
}

// GetTeam reads a team without its members (see ListTeamMembers). The returned team has no OrganizationId:
// AGG does not return the organization ID used by the REST API, so the caller fills it in from config or state.
func (c Client) GetTeam(ctx context.Context, teamId string, siteId string) (dto.TeamDto, error) {
	resp, err := c.query(ctx, "TeamV2", teamV2Query, map[string]any{"id": teamIdFromARI(teamId), "siteId": c.siteIdOrCloudId(siteId)})
	if err != nil {
		return dto.TeamDto{}, err
	}
	team, err := teamFromResponse(resp, teamId)
	if err != nil {
		return dto.TeamDto{}, err
	}
	// The query always needs a resolved site (falling back to cloud_id), but the returned team must echo the
	// caller's own input: nil when it configured no site_id, so the Terraform model keeps it unset instead of
	// filling in the cloud_id fallback.
	var returnedSiteId *string
	if siteId != "" {
		returnedSiteId = &siteId
	}
	return toTeamDto(*team, "", returnedSiteId, ""), nil
}

// ListTeamMembers returns one page of the team's current members, in the same shape as the REST members list:
// pass request.After = PageInfo.EndCursor to get the next page.
func (c Client) ListTeamMembers(ctx context.Context, teamId string, siteId string, request dto.TeamMemberListRequest) (dto.TeamMemberListResponse, error) {
	variables := map[string]any{"id": teamIdFromARI(teamId), "siteId": c.siteIdOrCloudId(siteId), "first": request.First}
	if request.After != "" {
		variables["after"] = request.After
	}
	resp, err := c.query(ctx, "TeamMembers", teamMembersQuery, variables)
	if err != nil {
		return dto.TeamMemberListResponse{}, err
	}
	team, err := teamFromResponse(resp, teamId)
	if err != nil {
		return dto.TeamMemberListResponse{}, err
	}

	// A real empty team answers with members: {nodes: []}, never members: null; a null members field means the
	// field itself failed to resolve, so it must not be read as "no members, last page".
	if team.Members == nil {
		if len(resp.Errors) > 0 {
			return dto.TeamMemberListResponse{}, fmt.Errorf("Teams GraphQL API TeamMembers errors: %s", errorMessages(resp.Errors))
		}
		return dto.TeamMemberListResponse{}, fmt.Errorf("Teams GraphQL API TeamMembers returned no members for team %s", teamId)
	}

	var page dto.TeamMemberListResponse
	page.Results = make([]dto.TeamMember, 0, len(team.Members.Nodes))
	for _, node := range team.Members.Nodes {
		page.Results = append(page.Results, dto.TeamMember{AccountId: accountIdFromARI(node.Member.AccountId)})
	}
	page.PageInfo.HasNextPage = team.Members.PageInfo.HasNextPage
	page.PageInfo.EndCursor = team.Members.PageInfo.EndCursor
	return page, nil
}

// UpdateTeam updates the team's display name and description; the team type cannot be changed (see UpdateTeamInput).
// The returned team, with the caller's permissions, comes from the mutation payload. Should the payload carry no
// team, the input is returned as it is and the permissions stay unknown until the next refresh.
func (c Client) UpdateTeam(ctx context.Context, teamDto dto.TeamDto, siteId string) (dto.TeamDto, error) {
	input := agg.UpdateTeamInput{
		TeamId:      teamARI(teamDto.TeamId),
		ScopeId:     c.siteScope(siteId),
		DisplayName: teamDto.DisplayName,
		Description: teamDto.Description,
	}
	resp, err := c.mutate(ctx, "UpdateTeam", updateTeamMutation, map[string]any{"input": input})
	if err != nil {
		return dto.TeamDto{}, err
	}
	payload := resp.Data.Team.UpdateTeam
	if err = checkMutation("UpdateTeam", payload, resp.Errors); err != nil {
		if payload != nil && teamGone(payload.Errors) {
			return dto.TeamDto{}, fmt.Errorf("%w: %w", ErrTeamNotFound, err)
		}
		return dto.TeamDto{}, err
	}
	if payload.Team == nil {
		return teamDto, nil
	}
	updated := toTeamDto(*payload.Team, teamDto.OrganizationId, teamDto.SiteId, teamDto.TeamType)
	if updated.TeamId == "" {
		updated.TeamId = teamDto.TeamId
	}
	return updated, nil
}

// DeleteTeam deletes a team; a team that is already gone counts as deleted.
func (c Client) DeleteTeam(ctx context.Context, teamId string, siteId string) error {
	input := agg.DeleteTeamInput{TeamId: teamARI(teamId), ScopeId: c.siteScope(siteId)}
	resp, err := c.mutate(ctx, "DeleteTeam", deleteTeamMutation, map[string]any{"input": input})
	if err != nil {
		// The outcome is unknown and is not confirmed by a read; the next refresh removes the team from state if
		// it is gone.
		return err
	}
	payload := resp.Data.Team.DeleteTeam
	if err = checkMutation("DeleteTeam", payload, resp.Errors); err != nil && !(payload != nil && teamGone(payload.Errors)) {
		return err
	}
	return nil
}

// AddMembers adds members (raw account IDs) to the team. A rejected member is reported in the error with its
// account ID and the reason.
func (c Client) AddMembers(ctx context.Context, teamId string, siteId string, accountIds []string) error {
	input := agg.MembersInput{TeamId: teamARI(teamId), ScopeId: c.siteScope(siteId), MemberIds: rawAccountIds(accountIds)}
	resp, err := c.mutate(ctx, "AddMembers", addMembersMutation, map[string]any{"input": input})
	if err != nil {
		return err
	}
	payload := resp.Data.Team.AddMembers
	err = checkMutation("AddMembers", payload, resp.Errors)
	if err != nil && payload != nil && teamGone(payload.Errors) {
		return fmt.Errorf("%w: %w", ErrTeamNotFound, err)
	}
	return err
}

// RemoveMembers removes members (raw account IDs) from the team. A rejected member is reported in the error with
// its account ID and the reason.
func (c Client) RemoveMembers(ctx context.Context, teamId string, siteId string, accountIds []string) error {
	input := agg.MembersInput{TeamId: teamARI(teamId), ScopeId: c.siteScope(siteId), MemberIds: rawAccountIds(accountIds)}
	resp, err := c.mutate(ctx, "RemoveMembers", removeMembersMutation, map[string]any{"input": input})
	if err != nil {
		return err
	}
	payload := resp.Data.Team.RemoveMembers
	err = checkMutation("RemoveMembers", payload, resp.Errors)
	if err != nil && payload != nil && teamGone(payload.Errors) {
		return fmt.Errorf("%w: %w", ErrTeamNotFound, err)
	}
	return err
}
