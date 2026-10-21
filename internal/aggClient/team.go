package aggClient

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto/agg"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

	createTeamMutation = `mutation CreateTeam($input: TeamCreateTeamInput!) {
  team @optIn(to: "Team-crud") {
    createTeam(input: $input) {
      team { id displayName description membershipSettings }
      success
      errors { message }
    }
  }
}`

	updateTeamMutation = `mutation UpdateTeam($input: TeamUpdateTeamInput!) {
  team @optIn(to: "Team-crud") {
    updateTeam(input: $input) {
      team { id displayName description membershipSettings }
      success
      errors { message }
    }
  }
}`

	deleteTeamMutation = `mutation DeleteTeam($input: TeamDeleteTeamInput!) {
  team @optIn(to: "Team-crud") {
    deleteTeam(input: $input) {
      success
      errors { message }
    }
  }
}`

	addMembersMutation = `mutation AddMembers($input: TeamAddMembersInput!) {
  team @optIn(to: "Team") @optIn(to: "Team-members") {
    addMembers(input: $input) {
      success
      errors { message }
    }
  }
}`

	removeMembersMutation = `mutation RemoveMembers($input: TeamRemoveMembersInput!) {
  team @optIn(to: "Team") @optIn(to: "Team-members") {
    removeMembers(input: $input) {
      success
      errors { message }
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

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
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

// withPermissions fills in the caller's permissions, which only teamV2 is known to return, after a create or update.
// Best effort: when the read fails the team is returned without them, and the next refresh fills them in.
func (c Client) withPermissions(ctx context.Context, teamDto dto.TeamDto) dto.TeamDto {
	if team, err := c.GetTeam(ctx, teamDto.TeamId, deref(teamDto.SiteId)); err == nil {
		teamDto.UserPermissions = team.UserPermissions
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Could not read the permissions of team %s: %s", teamDto.TeamId, err))
	}
	return teamDto
}

// CreateTeam creates a team with the given members (raw account IDs) and returns it with its new raw team ID.
func (c Client) CreateTeam(ctx context.Context, teamDto dto.TeamDto, memberAccountIds []string) (dto.TeamDto, error) {
	siteId := deref(teamDto.SiteId)
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

	resp, err := c.send(ctx, "CreateTeam", createTeamMutation, map[string]any{"input": input})
	if err != nil {
		return dto.TeamDto{}, err
	}
	payload := resp.Data.Team.CreateTeam
	if err = checkMutation("CreateTeam", payload, resp.Errors); err != nil {
		// The team can exist even when the mutation reports a failure (for example a member was rejected);
		// delete it so it is not left behind outside Terraform state.
		if payload != nil && payload.Team != nil && payload.Team.Id != "" {
			_ = c.DeleteTeam(ctx, teamIdFromARI(payload.Team.Id), siteId)
			return dto.TeamDto{}, fmt.Errorf("%w (the partially created team was deleted)", err)
		}
		return dto.TeamDto{}, err
	}
	if payload.Team == nil || payload.Team.Id == "" {
		return dto.TeamDto{}, fmt.Errorf("AGG CreateTeam returned no team ID")
	}
	return c.withPermissions(ctx, toTeamDto(*payload.Team, teamDto.OrganizationId, teamDto.SiteId, teamDto.TeamType)), nil
}

// teamFromResponse returns the team of a teamV2 response. It returns ErrTeamNotFound only when teamV2 explicitly
// reports the team as missing or deleted, so an ambiguous response never makes Terraform drop a live team.
func teamFromResponse(resp agg.Response, teamId string) (*agg.Team, error) {
	team := resp.Data.Team.TeamV2
	if team == nil {
		for _, e := range resp.Errors {
			if isTeamGone(e) {
				return nil, ErrTeamNotFound
			}
		}
		if len(resp.Errors) > 0 {
			return nil, fmt.Errorf("AGG TeamV2 errors: %s", errorMessages(resp.Errors))
		}
		return nil, fmt.Errorf("AGG TeamV2 returned no team %s", teamId)
	}
	if team.State == agg.TeamStateDisbanded || team.State == agg.TeamStatePurged {
		return nil, ErrTeamNotFound
	}
	return team, nil
}

// GetTeam reads a team without its members (see ListTeamMembers). The returned team has no OrganizationId:
// AGG does not return the organization ID used by the REST API, so the caller fills it in from config or state.
func (c Client) GetTeam(ctx context.Context, teamId string, siteId string) (dto.TeamDto, error) {
	siteId = c.siteIdOrCloudId(siteId)
	resp, err := c.send(ctx, "TeamV2", teamV2Query, map[string]any{"id": teamIdFromARI(teamId), "siteId": siteId})
	if err != nil {
		return dto.TeamDto{}, err
	}
	team, err := teamFromResponse(resp, teamId)
	if err != nil {
		return dto.TeamDto{}, err
	}
	return toTeamDto(*team, "", &siteId, ""), nil
}

// ListTeamMembers returns one page of the team's current members, in the same shape as the REST members list:
// pass request.After = PageInfo.EndCursor to get the next page.
func (c Client) ListTeamMembers(ctx context.Context, teamId string, siteId string, request dto.TeamMemberListRequest) (dto.TeamMemberListResponse, error) {
	variables := map[string]any{"id": teamIdFromARI(teamId), "siteId": c.siteIdOrCloudId(siteId), "first": request.First}
	if request.After != "" {
		variables["after"] = request.After
	}
	resp, err := c.send(ctx, "TeamMembers", teamMembersQuery, variables)
	if err != nil {
		return dto.TeamMemberListResponse{}, err
	}
	team, err := teamFromResponse(resp, teamId)
	if err != nil {
		return dto.TeamMemberListResponse{}, err
	}

	var page dto.TeamMemberListResponse
	page.Results = make([]dto.TeamMember, 0)
	if team.Members != nil {
		for _, node := range team.Members.Nodes {
			page.Results = append(page.Results, dto.TeamMember{AccountId: accountIdFromARI(node.Member.AccountId)})
		}
		page.PageInfo.HasNextPage = team.Members.PageInfo.HasNextPage
		page.PageInfo.EndCursor = team.Members.PageInfo.EndCursor
	}
	return page, nil
}

// UpdateTeam updates the team's display name and description.
func (c Client) UpdateTeam(ctx context.Context, teamDto dto.TeamDto) (dto.TeamDto, error) {
	input := agg.UpdateTeamInput{
		TeamId:      teamARI(teamDto.TeamId),
		ScopeId:     c.siteScope(deref(teamDto.SiteId)),
		DisplayName: teamDto.DisplayName,
		Description: teamDto.Description,
	}
	resp, err := c.send(ctx, "UpdateTeam", updateTeamMutation, map[string]any{"input": input})
	if err != nil {
		return dto.TeamDto{}, err
	}
	payload := resp.Data.Team.UpdateTeam
	if err = checkMutation("UpdateTeam", payload, resp.Errors); err != nil {
		return dto.TeamDto{}, err
	}
	if payload.Team == nil {
		return c.withPermissions(ctx, teamDto), nil
	}
	updated := toTeamDto(*payload.Team, teamDto.OrganizationId, teamDto.SiteId, teamDto.TeamType)
	if updated.TeamId == "" {
		updated.TeamId = teamDto.TeamId
	}
	return c.withPermissions(ctx, updated), nil
}

// DeleteTeam deletes a team. A team that is already gone counts as deleted.
func (c Client) DeleteTeam(ctx context.Context, teamId string, siteId string) error {
	input := agg.DeleteTeamInput{TeamId: teamARI(teamId), ScopeId: c.siteScope(siteId)}
	resp, err := c.send(ctx, "DeleteTeam", deleteTeamMutation, map[string]any{"input": input})
	if err == nil {
		err = checkMutation("DeleteTeam", resp.Data.Team.DeleteTeam, resp.Errors)
	}
	if err != nil {
		if _, readErr := c.GetTeam(ctx, teamId, siteId); errors.Is(readErr, ErrTeamNotFound) {
			tflog.Trace(ctx, "Team is already deleted")
			return nil
		}
		return err
	}
	return nil
}

// AddMembers adds members (raw account IDs) to the team.
func (c Client) AddMembers(ctx context.Context, teamId string, siteId string, accountIds []string) error {
	input := agg.MembersInput{TeamId: teamARI(teamId), ScopeId: c.siteScope(siteId), MemberIds: rawAccountIds(accountIds)}
	resp, err := c.send(ctx, "AddMembers", addMembersMutation, map[string]any{"input": input})
	if err != nil {
		return err
	}
	return checkMutation("AddMembers", resp.Data.Team.AddMembers, resp.Errors)
}

// RemoveMembers removes members (raw account IDs) from the team.
func (c Client) RemoveMembers(ctx context.Context, teamId string, siteId string, accountIds []string) error {
	input := agg.MembersInput{TeamId: teamARI(teamId), ScopeId: c.siteScope(siteId), MemberIds: rawAccountIds(accountIds)}
	resp, err := c.send(ctx, "RemoveMembers", removeMembersMutation, map[string]any{"input": input})
	if err != nil {
		return err
	}
	return checkMutation("RemoveMembers", resp.Data.Team.RemoveMembers, resp.Errors)
}
