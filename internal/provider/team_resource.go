// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/aggClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/url"
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TeamResource{}
var _ resource.ResourceWithImportState = &TeamResource{}

func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

// TeamResource defines the resource implementation.
type TeamResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
	agg                 aggClient.Client
}

func (r *TeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *TeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Atlassian team and enables JSM Operations for it.\n\n" +
			"With `auth_type = \"scoped\"` (a scoped API token or a service account), the team is managed through the Teams GraphQL API, " +
			"which Atlassian currently labels beta. See the [Scoped API Tokens and Service Accounts](https://registry.terraform.io/providers/atlassian/atlassian-operations/latest/docs/guides/scoped-api-tokens-and-service-accounts) guide for the " +
			"required scopes and the service account requirements. A service account can update or delete a team, and manage its " +
			"members, only while it is a member of that team, and the provider removes the creator from a new team unless it is " +
			"listed in `member`; list the service account in `member`, together with at least one human member, when Terraform must " +
			"update members or delete the team later. To create the team's schedules, policies, integrations and heartbeats in the " +
			"same apply, the account running Terraform must be a JSM Operations product admin or the member whose account ID sorts " +
			"first as text, which the provider nominates as the team's Operations admin. " +
			"Otherwise those resources fail with 403 \"You are not authorized …\" until a team admin makes the account an " +
			"Operations admin of the team in the Operations UI. Service accounts are not affected.",
		Attributes: schemaAttributes.TeamResourceAttributes,
	}
}

func (r *TeamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring TeamResource")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)

	if !ok {
		tflog.Error(ctx, "Unexpected Resource Configure Type")
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client
	r.agg = aggClient.New(client)

	tflog.Trace(ctx, "Configured TeamResource")
}

func (r *TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating the TeamResource")

	var data dataModels.TeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	plannedSiteId := data.SiteId

	teamDto, membersDto := TeamModelToDto(ctx, data)
	siteId := data.SiteId.ValueString()
	if siteId == "" {
		siteId = r.clientConfiguration.GetCloudId()
	}
	teamDto.SiteId = &siteId

	if r.clientConfiguration.IsScopedToken() {
		r.createViaAGG(ctx, data, teamDto, siteId, membersDto, resp)
		return
	}

	tflog.Trace(ctx, "Creating the Team")

	httpResp, err := httpClientHelpers.
		GenerateTeamsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("%s/teams/", teamDto.OrganizationId)).
		Method(httpClient.POST).
		SetBody(teamDto).
		SetBodyParseObject(&teamDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create team, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create team, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Team created")
	tflog.Trace(ctx, "Fetch auto created members")

	autoAddedMembers, err := r.fetchTeamMembers(teamDto.OrganizationId, teamDto.TeamId, data.SiteId.ValueString())
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to fetch members for the created team, %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch members for the created team, %s", err.Error()))
	}
	if resp.Diagnostics.HasError() {
		r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId.ValueString(), resp)
		return
	}

	addedUsers, removedUsers := diffUsers(membersDto, autoAddedMembers)

	if len(addedUsers) > 0 {
		tflog.Trace(ctx, "Adding users to the team")
		memberAddResponse := dto.PublicApiMembershipAddResponse{}
		httpResp, err = httpClientHelpers.
			GenerateTeamsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("%s/teams/%s/members/add", teamDto.OrganizationId, teamDto.TeamId)).
			Method(httpClient.POST).
			SetBody(dto.TeamMemberList{Members: addedUsers}).
			SetBodyParseObject(&memberAddResponse).
			Send()

		if httpResp == nil {
			tflog.Error(ctx, "Client Error. Unable to add users to the team, got nil response")
			resp.Diagnostics.AddError("Client Error", "Unable to add users to the team, got nil response")
		} else if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add users to the team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add users to the team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			} else {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add users to the team, got http response: %d", statusCode))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add users to the team, got http response: %d", statusCode))
			}
		}
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add users to the team, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add users to the team, got error: %s", err))
		} else if len(memberAddResponse.Errors) > 0 {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add users to the team, got errors: %v", memberAddResponse.Errors))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add users to the team, got errors: %v", memberAddResponse.Errors))
		}

		if resp.Diagnostics.HasError() {
			// If there is an error while adding users, the creation fails on Terraform's side, even though there is still a team on JSM side.
			// So, we need to delete the team on JSM side if the adding users fails.
			r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId.ValueString(), resp)
			return
		}
		tflog.Trace(ctx, "Users added to the team")
	}

	r.enableOps(ctx, teamDto, membersDto, resp)
	if resp.Diagnostics.HasError() {
		// If there is an error while enabling ops, the creation fails on Terraform's side, even though there is still a team on JSM side.
		// So, we need to delete the team on JSM side if the enabling ops fails.
		r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId.ValueString(), resp)
		return
	}
	tflog.Trace(ctx, "Enabled Operations for the Team")

	if data.DeleteDefaultResources.ValueBool() {
		r.deleteDefaultResources(ctx, teamDto.TeamId)
	}

	// Extra members (usually the creator) are removed last, so that a failure before this step can still be rolled back.
	if len(removedUsers) > 0 {
		tflog.Trace(ctx, "Removing extra users from the team")
		removeMemberResponse := dto.PublicApiMembershipRemoveResponse{}
		httpResp, err = httpClientHelpers.
			GenerateTeamsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("%s/teams/%s/members/remove", teamDto.OrganizationId, teamDto.TeamId)).
			Method(httpClient.POST).
			SetBody(dto.TeamMemberList{Members: removedUsers}).
			SetBodyParseObject(&removeMemberResponse).
			Send()

		if httpResp == nil {
			tflog.Error(ctx, "Client Error. Unable to remove extra team members, got nil response")
			resp.Diagnostics.AddError("Client Error", "Unable to remove extra team members, got nil response")
		} else if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove extra team members, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove extra team members, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			} else {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove extra team members, got http response: %d", statusCode))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove extra team members, got http response: %d", statusCode))
			}
		}
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove extra team members, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove extra team members, got error: %s", err))
		} else if len(removeMemberResponse.Errors) > 0 {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove extra team members, got errors: %v", removeMemberResponse.Errors))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove extra team members, got errors: %v", removeMemberResponse.Errors))
		}
	}

	if resp.Diagnostics.HasError() {
		r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId.ValueString(), resp)
		return
	}
	tflog.Trace(ctx, "Extra users removed from the team")

	data = TeamDtoToModel(teamDto, membersDto, data.DeleteDefaultResources, plannedSiteId.ValueString())

	tflog.Trace(ctx, "Created the TeamResource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

// enableOps enables JSM Operations for a newly created team. A new team takes a while to reach JSM Ops (404, or 422
// "Another request is already being processed now."), so this retries on its own dedicated budget for about two
// minutes, independently of api_retry_count. A failure is reported in resp.
func (r *TeamResource) enableOps(ctx context.Context, teamDto dto.TeamDto, membersDto []dto.TeamMember, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Enabling Operations for the Team")
	enableOpsBody := dto.TeamEnableOps{
		TeamId:          teamDto.TeamId,
		AdminAccountIds: []string{adminAccountId(membersDto)},
		InviteUsernames: make([]string, 0),
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		SetRetryCount(enableOpsRetryCount).SetRetryWaitTime(enableOpsRetryWait).SetRetryMaxWaitTime(enableOpsRetryWaitMax).
		AddRetryCondition(func(response *httpClient.Response, err error) bool {
			if response.GetStatusCode() == 404 || response.GetStatusCode() == 422 {
				return true
			}
			return false
		}).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/enable-ops", teamDto.TeamId)).
		Method(httpClient.POST).
		SetBody(enableOpsBody).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to enable Operations for the created team")
		resp.Diagnostics.AddError("Client Error", "Unable to enable Operations for the created team")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to enable Operations for the created team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable Operations for the created team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to enable Operations for the created team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable Operations for the created team, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to enable Operations for the created team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable Operations for the created team, got error: %s", err))
	}
}

// deleteDefaultResources deletes the routing rule, escalation and schedule that JSM creates for a new team.
func (r *TeamResource) deleteDefaultResources(ctx context.Context, teamId string) {
	tflog.Trace(ctx, "Deleting default resources for the team")

	err := findAndUpdateDefaultRoutingRule(teamId, r.clientConfiguration)
	if err != nil {
		tflog.Trace(ctx, "Could not find and update default routing rule for team", map[string]interface{}{"teamId": teamId, "error": err.Error()})
	}

	err = findAndDeleteDefaultEscalation(teamId, r.clientConfiguration)
	if err != nil {
		tflog.Trace(ctx, "Could not find and delete default escalation for team", map[string]interface{}{"teamId": teamId, "error": err.Error()})
	}

	err = findAndDeleteDefaultSchedule(teamId, r.clientConfiguration)
	if err != nil {
		tflog.Trace(ctx, "Could not find and delete default schedule for team", map[string]interface{}{"teamId": teamId, "error": err.Error()})
	}
}

// createViaAGG is Create for scoped API tokens: the same steps in the same order as the Teams REST flow, through the
// Teams GraphQL API.
func (r *TeamResource) createViaAGG(ctx context.Context, data dataModels.TeamModel, teamDto dto.TeamDto, siteId string, membersDto []dto.TeamMember, resp *resource.CreateResponse) {
	plannedSiteId := data.SiteId.ValueString()

	tflog.Trace(ctx, "Creating the Team")

	created, err := r.agg.CreateTeam(ctx, teamDto, siteId, memberAccountIds(membersDto))
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got error: %s", err))
	} else {
		teamDto = created
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Team created")
	tflog.Trace(ctx, "Fetch auto created members")

	autoAddedMembers, err := fetchTeamMembersViaAGG(ctx, r.agg, teamDto.TeamId, siteId)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to fetch members for the created team, %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch members for the created team, %s", err.Error()))
	}
	if resp.Diagnostics.HasError() {
		r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId, resp)
		return
	}

	addedUsers, removedUsers := diffUsers(membersDto, autoAddedMembers)

	if len(addedUsers) > 0 {
		tflog.Trace(ctx, "Adding users to the team")
		if err = r.agg.AddMembers(ctx, teamDto.TeamId, siteId, memberAccountIds(addedUsers)); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add users to the team, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add users to the team, got error: %s", err))
		}

		if resp.Diagnostics.HasError() {
			// If there is an error while adding users, the creation fails on Terraform's side, even though there is still a team on JSM side.
			// So, we need to delete the team on JSM side if the adding users fails.
			r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId, resp)
			return
		}
		tflog.Trace(ctx, "Users added to the team")
	}

	r.enableOps(ctx, teamDto, membersDto, resp)
	if resp.Diagnostics.HasError() {
		// If there is an error while enabling ops, the creation fails on Terraform's side, even though there is still a team on JSM side.
		// So, we need to delete the team on JSM side if the enabling ops fails.
		r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId, resp)
		return
	}
	tflog.Trace(ctx, "Enabled Operations for the Team")

	if data.DeleteDefaultResources.ValueBool() {
		r.deleteDefaultResources(ctx, teamDto.TeamId)
	}

	// Extra members (usually the creator) are removed last, after Operations is enabled and the default resources are
	// handled. A service account that is not listed in `member` keeps its team permissions until here, so a creation
	// that fails in one of the steps above can still be rolled back: DeleteTeam needs CAN_DELETE_TEAM, which a service
	// account only has while it is a member.
	if len(removedUsers) > 0 {
		tflog.Trace(ctx, "Removing extra users from the team")
		if err = r.agg.RemoveMembers(ctx, teamDto.TeamId, siteId, memberAccountIds(removedUsers)); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove extra team members, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove extra team members, got error: %s", err))
		}
	}

	if resp.Diagnostics.HasError() {
		r.rollbackCreatedTeam(ctx, teamDto, siteId, plannedSiteId, resp)
		return
	}
	tflog.Trace(ctx, "Extra users removed from the team")

	data = TeamDtoToModel(teamDto, membersDto, data.DeleteDefaultResources, plannedSiteId)

	tflog.Trace(ctx, "Created the TeamResource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

// list schedules using teamId then delete its default schedule
func findAndDeleteDefaultSchedule(teamId string, configuration dto.AtlassianOpsProviderModel) error {
	tflog.Trace(context.Background(), "Finding and deleting default schedule for team", map[string]interface{}{"teamId": teamId})

	baseURL := "/v1/schedules"
	queryParams := map[string]string{}

	doneLooping := false
	deleted := false

	for !doneLooping && !deleted {
		var listScheduleDto = dto.ListSchedule{}
		req := httpClientHelpers.
			GenerateJsmOpsClientRequest(configuration).
			JoinBaseUrl(baseURL).
			Method(httpClient.GET).
			SetQueryParams(queryParams).
			SetBodyParseObject(&listScheduleDto)

		httpResp, err := req.Send()

		if err != nil {
			return fmt.Errorf("error fetching schedules: %w", err)
		}

		if httpResp == nil || httpResp.IsError() {
			return fmt.Errorf("error fetching schedules: empty response")
		}

		for _, schedule := range listScheduleDto.Values {
			if strings.EqualFold(schedule.TeamId, teamId) {
				deleteResp, err := httpClientHelpers.
					GenerateJsmOpsClientRequest(configuration).
					JoinBaseUrl(fmt.Sprintf("/v1/schedules/%s", schedule.Id)).
					Method(httpClient.DELETE).
					Send()

				if err != nil || deleteResp.IsError() {
					return fmt.Errorf("error deleting schedule: %w", err)
				}
				tflog.Trace(context.Background(), "Deleted default schedule for team", map[string]interface{}{"teamId": teamId, "scheduleId": schedule.Id})
				deleted = true
				break
			}
		}

		if deleted || listScheduleDto.Links.Next == "" {
			doneLooping = true
		} else {
			nextURL := listScheduleDto.Links.Next
			parsedURL, err := url.Parse(nextURL)
			if err != nil {
				return fmt.Errorf("error parsing next URL: %w", err)
			}
			urlValues := parsedURL.Query()
			queryParams = make(map[string]string)
			for key, values := range urlValues {
				if len(values) > 0 {
					queryParams[key] = values[0]
				}
			}
			baseURL = parsedURL.Path
		}
	}
	return nil
}

// list escalations using teamId then delete its default escalation
func findAndDeleteDefaultEscalation(teamId string, configuration dto.AtlassianOpsProviderModel) error {
	tflog.Trace(context.Background(), "Finding and deleting default escalation for team", map[string]interface{}{"teamId": teamId})

	var listEscalationDto = dto.ListEscalationDto{}
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(configuration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/escalations", teamId)).
		Method(httpClient.GET).
		SetBodyParseObject(&listEscalationDto).
		Send()

	if err != nil {
		return fmt.Errorf("error fetching escalations: %w", err)
	}

	if httpResp == nil || httpResp.IsError() {
		return fmt.Errorf("error fetching escalations: empty response")
	}

	for _, escalation := range listEscalationDto.Values {
		deleteResp, err := httpClientHelpers.
			GenerateJsmOpsClientRequest(configuration).
			JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/escalations/%s", teamId, escalation.Id)).
			Method(httpClient.DELETE).
			Send()

		if err != nil || deleteResp.IsError() {
			return fmt.Errorf("error deleting escalation: %w", err)
		}
		tflog.Trace(context.Background(), "Deleted default escalation for team", map[string]interface{}{"teamId": teamId, "escalationId": escalation.Id})
		break
	}

	return nil
}

// list routing rules using teamId then update its Notify to None
func findAndUpdateDefaultRoutingRule(teamId string, configuration dto.AtlassianOpsProviderModel) error {
	tflog.Trace(context.Background(), "Finding and updating default routing rule for team", map[string]interface{}{"teamId": teamId})

	var listRoutingRuleDto = dto.ListRoutingRuleDto{}
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(configuration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/routing-rules", teamId)).
		Method(httpClient.GET).
		SetBodyParseObject(&listRoutingRuleDto).
		Send()

	if err != nil {
		return fmt.Errorf("error fetching routing rules: %w", err)
	}

	if httpResp == nil || httpResp.IsError() {
		return fmt.Errorf("error fetching routing rules: empty response")
	}

	for _, rule := range listRoutingRuleDto.Values {
		if rule.IsDefault {
			rule.Notify = &dto.RoutingRuleNotifyDto{
				Type: "none",
				ID:   "",
			}
			updateResp, err := httpClientHelpers.
				GenerateJsmOpsClientRequest(configuration).
				JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/routing-rules/%s", teamId, rule.ID)).
				Method(httpClient.PATCH).
				SetBody(rule).
				Send()

			if err != nil || updateResp.IsError() {
				return fmt.Errorf("error updating routing rule: %w", err)
			}
			tflog.Trace(context.Background(), "Updated default routing rule for team", map[string]interface{}{"teamId": teamId, "ruleId": rule.ID})
			break
		}
	}

	return nil
}

func (r *TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.TeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Reading the TeamResource")

	teamDto := dto.TeamDto{}
	siteId := data.SiteId.ValueString()
	if siteId == "" {
		siteId = r.clientConfiguration.GetCloudId()
	}

	if r.clientConfiguration.IsScopedToken() {
		r.readViaAGG(ctx, data, siteId, resp)
		return
	}

	httpResp, err := httpClientHelpers.
		GenerateTeamsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("%s/teams/%s", data.OrganizationId.ValueString(), data.Id.ValueString())).
		Method(httpClient.GET).
		SetQueryParam("siteId", siteId).
		SetBodyParseObject(&teamDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read team, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read team, got nil response")
	} else if httpResp.GetStatusCode() == 404 {
		resp.State.RemoveResource(ctx)

		return
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team or to parse received data, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Fetching team members")

	memberData, err := r.fetchTeamMembers(data.OrganizationId.ValueString(), data.Id.ValueString(), data.SiteId.ValueString())
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to fetch members for the created team, %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch members for the created team, %s", err.Error()))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "Done fetching team members")

	tflog.Trace(ctx, "Converting Team Data into Terraform Model")

	data = TeamDtoToModel(teamDto, memberData, data.DeleteDefaultResources, data.SiteId.ValueString())

	tflog.Trace(ctx, "Read the TeamResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

// readViaAGG is Read for scoped API tokens: the team and its members come from the Teams GraphQL API.
func (r *TeamResource) readViaAGG(ctx context.Context, data dataModels.TeamModel, siteId string, resp *resource.ReadResponse) {
	teamDto, err := r.agg.GetTeam(ctx, data.Id.ValueString(), siteId)
	if errors.Is(err, aggClient.ErrTeamNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, got error: %s", err))
		return
	}
	// AGG does not return the organization ID used by the REST API; keep the one in state.
	teamDto.OrganizationId = data.OrganizationId.ValueString()
	if teamDto.TeamType == "" {
		teamDto.TeamType = dto.TeamType(data.TeamType.ValueString())
	}

	tflog.Trace(ctx, "Fetching team members")

	memberData, err := fetchTeamMembersViaAGG(ctx, r.agg, data.Id.ValueString(), siteId)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team members, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team members, got error: %s", err))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "Done fetching team members")

	tflog.Trace(ctx, "Converting Team Data into Terraform Model")

	data = TeamDtoToModel(teamDto, memberData, data.DeleteDefaultResources, data.SiteId.ValueString())

	tflog.Trace(ctx, "Read the TeamResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

func (r *TeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var currentData dataModels.TeamModel
	var newData dataModels.TeamModel

	req.State.Get(ctx, &currentData)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newData)...)

	if !currentData.OrganizationId.Equal(newData.OrganizationId) && currentData.Id.Equal(newData.Id) {
		tflog.Error(ctx, "Invalid Update. Organization ID cannot be changed, once a resource is created")
		resp.Diagnostics.AddError("Invalid Update", "Organization ID cannot be changed, once a resource is created")
	}

	if !currentData.TeamType.Equal(newData.TeamType) && currentData.Id.Equal(newData.Id) {
		tflog.Error(ctx, "Invalid Update. Team Type cannot be changed, once a resource is created")
		resp.Diagnostics.AddError("Invalid Update", "Team Type cannot be changed, once a resource is created")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Updating the TeamResource")

	if r.clientConfiguration.IsScopedToken() {
		r.updateViaAGG(ctx, currentData, newData, resp)
		return
	}

	newTeamDto, newUsersDto := TeamModelToDto(ctx, newData)
	_, currentUsersDto := TeamModelToDto(ctx, currentData)

	httpResp, err := httpClientHelpers.
		GenerateTeamsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("%s/teams/%s", newData.OrganizationId.ValueString(), newData.Id.ValueString())).
		Method(httpClient.PATCH).
		SetBody(newTeamDto).
		SetBodyParseObject(&newTeamDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update team, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update team, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Updating the team members")
	addedUsers, removedUsers := diffUsers(newUsersDto, currentUsersDto)

	if len(addedUsers) > 0 {
		tflog.Trace(ctx, "Adding new team members")
		httpResp, err = httpClientHelpers.
			GenerateTeamsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("%s/teams/%s/members/add", newData.OrganizationId.ValueString(), newData.Id.ValueString())).
			Method(httpClient.POST).
			SetBody(dto.TeamMemberList{Members: addedUsers}).
			Send()

		if httpResp == nil {
			tflog.Error(ctx, "Client Error. Unable to add new team members, got nil response")
			resp.Diagnostics.AddError("Client Error", "Unable to add new team members, got nil response")
		} else if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add new team members, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add new team members, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			} else {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add new team members, got http response: %d", statusCode))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add new team members, got http response: %d", statusCode))
			}
		}
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add new team members, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add new team members, got error: %s", err))
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(removedUsers) > 0 {
		tflog.Trace(ctx, "Removing old team members")
		removeMembersResponse := dto.PublicApiMembershipRemoveResponse{}
		httpResp, err = httpClientHelpers.
			GenerateTeamsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("%s/teams/%s/members/remove", currentData.OrganizationId.ValueString(), currentData.Id.ValueString())).
			Method(httpClient.POST).
			SetBody(dto.TeamMemberList{Members: removedUsers}).
			SetBodyParseObject(&removeMembersResponse).
			Send()

		if httpResp == nil {
			tflog.Error(ctx, "Client Error. Unable to remove old team members, got nil response")
			resp.Diagnostics.AddError("Client Error", "Unable to remove old team members, got nil response")
		} else if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove old team members, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove old team members, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			} else {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove old team members, got http response: %d", statusCode))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove old team members, got http response: %d", statusCode))
			}
		}
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove old team members, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove old team members, got error: %s", err))
		} else if len(removeMembersResponse.Errors) > 0 {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove old team members, got errors: %v", removeMembersResponse.Errors))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove old team members, got errors: %v", removeMembersResponse.Errors))
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	newData = TeamDtoToModel(newTeamDto, newUsersDto, newData.DeleteDefaultResources, newData.SiteId.ValueString())

	tflog.Trace(ctx, "Updated the TeamResource")

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

// updateViaAGG is Update for scoped API tokens: the team and its members are updated through the Teams GraphQL API.
func (r *TeamResource) updateViaAGG(ctx context.Context, currentData dataModels.TeamModel, newData dataModels.TeamModel, resp *resource.UpdateResponse) {
	newTeamDto, newUsersDto := TeamModelToDto(ctx, newData)
	_, currentUsersDto := TeamModelToDto(ctx, currentData)

	updated, err := r.agg.UpdateTeam(ctx, newTeamDto, newData.SiteId.ValueString())
	if errors.Is(err, aggClient.ErrTeamNotFound) {
		// Unlike Read, Update does not remove the resource from state: the plan already committed to an
		// update, so the clearer outcome is an error telling the user to refresh and plan again, which lets
		// Read notice the deletion and the next plan decide whether to recreate it.
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update team: %s", err))
		resp.Diagnostics.AddError("Client Error",
			"Unable to update team: it no longer exists. Run `terraform apply` again to let Terraform notice the deletion and plan accordingly.")
	} else if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, got error: %s", err))
	} else {
		newTeamDto = updated
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Updating the team members")
	addedUsers, removedUsers := diffUsers(newUsersDto, currentUsersDto)

	if len(addedUsers) > 0 {
		tflog.Trace(ctx, "Adding new team members")
		if err = r.agg.AddMembers(ctx, newData.Id.ValueString(), newData.SiteId.ValueString(), memberAccountIds(addedUsers)); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add new team members, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add new team members, got error: %s", err))
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(removedUsers) > 0 {
		tflog.Trace(ctx, "Removing old team members")
		if err = r.agg.RemoveMembers(ctx, currentData.Id.ValueString(), currentData.SiteId.ValueString(), memberAccountIds(removedUsers)); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove old team members, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove old team members, got error: %s", err))
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	newData = TeamDtoToModel(newTeamDto, newUsersDto, newData.DeleteDefaultResources, newData.SiteId.ValueString())

	tflog.Trace(ctx, "Updated the TeamResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

func (r *TeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.TeamModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Deleting the TeamResource")

	if r.clientConfiguration.IsScopedToken() {
		if err := r.agg.DeleteTeam(ctx, data.Id.ValueString(), data.SiteId.ValueString()); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete team, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, got error: %s", err))
		}
		if !resp.Diagnostics.HasError() {
			tflog.Trace(ctx, "Deleted the TeamResource")
		}
		return
	}

	httpResp, err := httpClientHelpers.
		GenerateTeamsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("%s/teams/%s", data.OrganizationId.ValueString(), data.Id.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete team, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete team, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, status code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse)))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, got http response: %d", statusCode))
		}
	}
	if httpResp != nil && err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete team, got http response: %d", httpResp.GetStatusCode()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleted the TeamResource")
}

func (r *TeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) < 2 || len(idParts) > 3 || idParts[0] == "" || idParts[1] == "" || (len(idParts) == 3 && idParts[2] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,organization_id (without site scoping); or: id,organization_id,site_id (with site scoping). Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[1])...)
	if len(idParts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), idParts[2])...)
	}
}

func (r *TeamResource) fetchTeamMembers(organizationId string, teamId string, siteId string) ([]dto.TeamMember, error) {
	if siteId == "" {
		siteId = r.clientConfiguration.GetCloudId()
	}
	var members []dto.TeamMember

	doneLooping := false
	request := dto.DefaultTeamMemberListRequest()
	for !doneLooping {
		response := dto.TeamMemberListResponse{}

		httpResp, err := httpClientHelpers.
			GenerateTeamsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("/%s/teams/%s/members", organizationId, teamId)).
			Method("POST").
			SetQueryParam("siteId", siteId).
			SetBody(request).
			SetBodyParseObject(&response).
			Send()

		if err != nil {
			return nil, err
		} else if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				return nil, fmt.Errorf("error while fetching team members. Status Code: %d. Got response: %s%s", statusCode, *errorResponse, scopedTokenHint(r.clientConfiguration.IsScopedToken(), statusCode, *errorResponse))
			} else {
				return nil, fmt.Errorf("error while fetching team members. Status Code: %d", httpResp.GetStatusCode())
			}
		}

		members = append(members, response.Results...)
		if !response.PageInfo.HasNextPage {
			doneLooping = true
		} else {
			request.After = response.PageInfo.EndCursor
		}
	}
	return members, nil
}

// fetchTeamMembersViaAGG pages through the team's members in AGG with the same loop as the Teams REST API
// members list above: same request (first, after) and response (results, pageInfo) shapes.
func fetchTeamMembersViaAGG(ctx context.Context, client aggClient.Client, teamId string, siteId string) ([]dto.TeamMember, error) {
	var members []dto.TeamMember

	doneLooping := false
	request := dto.DefaultTeamMemberListRequest()
	for !doneLooping {
		response, err := client.ListTeamMembers(ctx, teamId, siteId, request)
		if err != nil {
			return nil, err
		}

		members = append(members, response.Results...)
		if response.PageInfo.HasNextPage && (response.PageInfo.EndCursor == "" || response.PageInfo.EndCursor == request.After) {
			return nil, fmt.Errorf("Teams GraphQL API TeamMembers returned an invalid page cursor for team %s", teamId)
		}
		if !response.PageInfo.HasNextPage {
			doneLooping = true
		} else {
			request.After = response.PageInfo.EndCursor
		}
	}
	return members, nil
}

// adminAccountId returns the member with the lexicographically smallest account ID: the provider's explicit,
// specified choice for the team's Operations admin, so it does not depend on the order Terraform hands `member`
// to the provider (a set has none) or on cty's internal, undefined ordering for object elements. Any member can be
// made admin later in the Operations UI.
func adminAccountId(members []dto.TeamMember) string {
	admin := members[0].AccountId
	for _, m := range members[1:] {
		if m.AccountId < admin {
			admin = m.AccountId
		}
	}
	return admin
}

// rollbackCreatedTeam deletes a team whose creation failed halfway, so that it is not left behind outside Terraform
// state. When the deletion fails too, it reports the orphaned team and how to recover it: the import ID follows
// plannedSiteId (the plan's site_id), not siteId, which Create defaults to the provider's cloud_id when
// site_id is not configured so the API has one to act on; importing that default would store site_id = cloud_id in
// state, which then keeps diffing against a configuration that omits site_id.
func (r *TeamResource) rollbackCreatedTeam(ctx context.Context, teamDto dto.TeamDto, siteId string, plannedSiteId string, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Deleting the team whose creation failed")
	// Ctrl-C or a deadline cancels ctx; the cleanup gets its own short budget so the team is not left behind.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if err := r.deleteCreatedTeam(cleanupCtx, teamDto, siteId); err != nil {
		importId := teamDto.TeamId + "," + teamDto.OrganizationId
		if plannedSiteId != "" {
			importId += "," + plannedSiteId
		}
		tflog.Error(ctx, fmt.Sprintf("Unable to delete the team %s after its creation failed: %s", teamDto.TeamId, err))
		resp.Diagnostics.AddError(
			"Team left behind after a failed creation",
			fmt.Sprintf("The team %q (id %s) was created, but it could not be deleted after the creation failed: %s. "+
				"Delete it in Atlassian Administration, or import it with `terraform import atlassian-operations_team.<name> \"%s\"` and apply again.",
				teamDto.DisplayName, teamDto.TeamId, err, importId),
		)
		return
	}
	tflog.Trace(ctx, "Deleted the team whose creation failed")
}

// deleteCreatedTeam deletes the team through the API that created it and reports a failed deletion.
func (r *TeamResource) deleteCreatedTeam(ctx context.Context, teamDto dto.TeamDto, siteId string) error {
	if r.clientConfiguration.IsScopedToken() {
		return r.agg.DeleteTeam(ctx, teamDto.TeamId, siteId)
	}
	httpResp, err := httpClientHelpers.
		GenerateTeamsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("%s/teams/%s", teamDto.OrganizationId, teamDto.TeamId)).
		Method(httpClient.DELETE).
		Send()
	if err != nil {
		return err
	}
	if httpResp == nil {
		return fmt.Errorf("got nil response")
	}
	if httpResp.IsError() {
		if body := httpResp.GetErrorBody(); body != nil {
			return fmt.Errorf("status code %d: %s", httpResp.GetStatusCode(), *body)
		}
		return fmt.Errorf("status code %d", httpResp.GetStatusCode())
	}
	return nil
}

func diffUsers(newDto []dto.TeamMember, oldDto []dto.TeamMember) ([]dto.TeamMember, []dto.TeamMember) {
	addedUsers := make([]dto.TeamMember, 0)
	removedUsers := make([]dto.TeamMember, 0)

	for _, user := range newDto {
		found := false
		for _, user2 := range oldDto {
			if user.AccountId == user2.AccountId {
				found = true
				break
			}
		}

		if !found {
			addedUsers = append(addedUsers, user)
		}
	}

	for _, user := range oldDto {
		found := false
		for _, user2 := range newDto {
			if user.AccountId == user2.AccountId {
				found = true
				break
			}
		}

		if !found {
			removedUsers = append(removedUsers, user)
		}
	}

	return addedUsers, removedUsers
}

// A newly created team takes a while to reach JSM Ops on both the classic and the scoped path, so enable-ops retries
// for about two minutes.
const (
	enableOpsRetryCount   = 10
	enableOpsRetryWait    = 2 * time.Second
	enableOpsRetryWaitMax = 15 * time.Second
)

func memberAccountIds(members []dto.TeamMember) []string {
	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.AccountId
	}
	return ids
}
