// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TeamResource{}
var _ resource.ResourceWithImportState = &TeamResource{}

func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

// TeamResource defines the resource implementation.
type TeamResource struct {
	teamClient *httpClient.HttpClient
	opsClient  *httpClient.HttpClient
}

func (r *TeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *TeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.TeamResourceAttributes,
	}
}

func (r *TeamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring TeamResource")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*JsmOpsClient)

	if !ok {
		tflog.Error(ctx, "Unexpected Resource Configure Type")
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.teamClient = client.TeamClient
	r.opsClient = client.OpsClient

	tflog.Trace(ctx, "Configured TeamResource")
}

func (r *TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating the TeamResource")

	var data dataModels.TeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	teamDto, membersDto := TeamModelToDto(ctx, data)

	tflog.Trace(ctx, "Creating the Team")

	httpResp, err := r.teamClient.NewRequest().
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
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create team, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, status code: %d. Got response: %s", statusCode, *errorResponse))
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

	autoAddedMembers, err := r.fetchTeamMembers(teamDto.OrganizationId, teamDto.TeamId)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to fetch members for the created team, %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch members for the created team, %s", err.Error()))
	}
	if resp.Diagnostics.HasError() {
		tflog.Trace(ctx, "Deleting dangling team resource")
		r.cleanupTeamSilent(teamDto)
		return
	}

	addedUsers, removedUsers := diffUsers(membersDto, autoAddedMembers)

	if len(addedUsers) > 0 {
		tflog.Trace(ctx, "Adding users to the team")
		memberAddResponse := dto.PublicApiMembershipAddResponse{}
		httpResp, err = r.teamClient.NewRequest().
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
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add users to the team, status code: %d. Got response: %s", statusCode, *errorResponse))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add users to the team, status code: %d. Got response: %s", statusCode, *errorResponse))
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
			tflog.Trace(ctx, "Deleting dangling team resource")
			r.cleanupTeamSilent(teamDto)
			return
		}
		tflog.Trace(ctx, "Users added to the team")
	}

	if len(removedUsers) > 0 {
		tflog.Trace(ctx, "Removing extra users from the team")
		removeMemberResponse := dto.PublicApiMembershipRemoveResponse{}
		httpResp, err = r.teamClient.NewRequest().
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
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove extra team members, status code: %d. Got response: %s", statusCode, *errorResponse))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove extra team members, status code: %d. Got response: %s", statusCode, *errorResponse))
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
		tflog.Trace(ctx, "Deleting dangling team resource")
		r.cleanupTeamSilent(teamDto)
		return
	}
	tflog.Trace(ctx, "Extra users removed from the team")

	tflog.Trace(ctx, "Enabling Operations for the Team")
	enableOpsBody := dto.TeamEnableOps{
		TeamId:          teamDto.TeamId,
		AdminAccountIds: []string{membersDto[0].AccountId},
		InviteUsernames: make([]string, 0),
	}

	// Enable OPS for the Team
	httpResp, err = r.opsClient.
		AddRetryCondition(func(response *httpClient.Response, err error) bool {
			if response.GetStatusCode() == 404 {
				return true
			}
			return false
		}).
		NewRequest().
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
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to enable Operations for the created team, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable Operations for the created team, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to enable Operations for the created team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable Operations for the created team, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to enable Operations for the created team, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable Operations for the created team, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		// If there is an error while enabling ops, the creation fails on Terraform's side, even though there is still a team on JSM side.
		// So, we need to delete the team on JSM side if the enabling ops fails.
		tflog.Trace(ctx, "Deleting dangling team resource")
		r.cleanupTeamSilent(teamDto)
		return
	}
	tflog.Trace(ctx, "Enabled Operations for the Team")

	data = TeamDtoToModel(teamDto, membersDto)

	tflog.Trace(ctx, "Created the TeamResource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

func (r *TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.TeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Reading the TeamResource")

	teamDto := dto.TeamDto{}

	httpResp, err := r.teamClient.NewRequest().
		JoinBaseUrl(fmt.Sprintf("%s/teams/%s", data.OrganizationId.ValueString(), data.Id.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&teamDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read team, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read team, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, status code: %d. Got response: %s", statusCode, *errorResponse))
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

	memberData, err := r.fetchTeamMembers(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to fetch members for the created team, %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch members for the created team, %s", err.Error()))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "Done fetching team members")

	tflog.Trace(ctx, "Converting Team Data into Terraform Model")

	data = TeamDtoToModel(teamDto, memberData)

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

	newTeamDto, newUsersDto := TeamModelToDto(ctx, newData)
	_, currentUsersDto := TeamModelToDto(ctx, currentData)

	httpResp, err := r.teamClient.NewRequest().
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
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update team, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, status code: %d. Got response: %s", statusCode, *errorResponse))
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
		httpResp, err = r.teamClient.NewRequest().
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
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to add new team members, status code: %d. Got response: %s", statusCode, *errorResponse))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add new team members, status code: %d. Got response: %s", statusCode, *errorResponse))
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
		httpResp, err = r.teamClient.NewRequest().
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
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to remove old team members, status code: %d. Got response: %s", statusCode, *errorResponse))
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove old team members, status code: %d. Got response: %s", statusCode, *errorResponse))
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

	newData = TeamDtoToModel(newTeamDto, newUsersDto)

	tflog.Trace(ctx, "Updated the TeamResource")

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
	tflog.Trace(ctx, "Saved the TeamResource into Terraform state")
}

func (r *TeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.TeamModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Deleting the TeamResource")

	httpResp, err := r.teamClient.NewRequest().
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
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete team, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, status code: %d. Got response: %s", statusCode, *errorResponse))
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
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,organization_id. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[1])...)
}

func (r *TeamResource) fetchTeamMembers(organizationId string, teamId string) ([]dto.TeamMember, error) {
	var members []dto.TeamMember

	doneLooping := false
	request := dto.DefaultTeamMemberListRequest()
	for !doneLooping {
		response := dto.TeamMemberListResponse{}

		httpResp, err := r.teamClient.NewRequest().
			JoinBaseUrl(fmt.Sprintf("/%s/teams/%s/members", organizationId, teamId)).
			Method("POST").
			SetBody(request).
			SetBodyParseObject(&response).
			Send()

		if err != nil {
			return nil, err
		} else if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				return nil, fmt.Errorf("error while fetching team members. Status Code: %d. Got response: %s", statusCode, *errorResponse)
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

func (r *TeamResource) cleanupTeamSilent(teamDto dto.TeamDto) {
	_, _ = r.teamClient.NewRequest().
		JoinBaseUrl(fmt.Sprintf("%s/teams/%s", teamDto.OrganizationId, teamDto.TeamId)).
		Method(httpClient.DELETE).
		Send()
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
