// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                     = &TeamRoleResource{}
	_ resource.ResourceWithConfigure        = &TeamRoleResource{}
	_ resource.ResourceWithImportState      = &TeamRoleResource{}
	_ resource.ResourceWithConfigValidators = &TeamRoleResource{}
)

type TeamRoleResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func NewTeamRoleResource() resource.Resource {
	return &TeamRoleResource{}
}

func (r *TeamRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_role"
}

func (r *TeamRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a team role in Atlassian JSM Operations. Team roles control granular permissions for team members.",
		Attributes:  schemaAttributes.TeamRoleResourceAttributes,
	}
}

func (r *TeamRoleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&teamRoleRightsValidator{},
	}
}

func (r *TeamRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring TeamRoleResource")

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected dto.AtlassianOpsProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client
}

func (r *TeamRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating TeamRole resource")

	var plan dataModels.TeamRoleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := plan.TeamId.ValueString()
	teamRoleDto := TeamRoleModelToDto(ctx, &plan)

	var responseDto dto.TeamRoleResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/roles", teamId)).
		Method(httpClient.POST).
		SetBody(teamRoleDto).
		SetBodyParseObject(&responseDto).
		Send()

	if httpResp == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team role, got nil response"))
		return
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team role, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team role, got http response: %d", statusCode))
		}
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team role, got error: %s", err))
		return
	}

	model := TeamRoleDtoToModel(&responseDto, teamId, plan.GrantedRights, plan.DisallowedRights)
	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, fmt.Sprintf("Created TeamRole resource with ID: %s", model.ID.ValueString()))
}

func (r *TeamRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "Reading TeamRole resource")

	var state dataModels.TeamRoleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()
	identifier := state.ID.ValueString()

	var teamRoleDto dto.TeamRoleResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/roles/%s", teamId, identifier)).
		Method(httpClient.GET).
		SetBodyParseObject(&teamRoleDto).
		Send()

	if httpResp == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team role, got nil response"))
		return
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		if statusCode == 404 {
			tflog.Warn(ctx, fmt.Sprintf("Team role with ID %s not found, removing from state", identifier))
			resp.State.RemoveResource(ctx)
			return
		}
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team role, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team role, got http response: %d", statusCode))
		}
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team role, got error: %s", err))
		return
	}

	model := TeamRoleDtoToModel(&teamRoleDto, teamId, state.GrantedRights, state.DisallowedRights)
	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, fmt.Sprintf("Read TeamRole resource with ID: %s", model.ID.ValueString()))
}

func (r *TeamRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "Updating TeamRole resource")

	var plan dataModels.TeamRoleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state dataModels.TeamRoleModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := plan.TeamId.ValueString()
	identifier := state.ID.ValueString()
	teamRoleDto := TeamRoleModelToDto(ctx, &plan)

	var responseDto dto.TeamRoleResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/roles/%s", teamId, identifier)).
		Method(httpClient.PATCH).
		SetBody(teamRoleDto).
		SetBodyParseObject(&responseDto).
		Send()

	if httpResp == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team role, got nil response"))
		return
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team role, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team role, got http response: %d", statusCode))
		}
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team role, got error: %s", err))
		return
	}

	model := TeamRoleDtoToModel(&responseDto, teamId, plan.GrantedRights, plan.DisallowedRights)
	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, fmt.Sprintf("Updated TeamRole resource with ID: %s", model.ID.ValueString()))
}

func (r *TeamRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "Deleting TeamRole resource")

	var state dataModels.TeamRoleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()
	identifier := state.ID.ValueString()

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/roles/%s", teamId, identifier)).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team role, got nil response"))
		return
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team role, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team role, got http response: %d", statusCode))
		}
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team role, got error: %s", err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted TeamRole resource with ID: %s", identifier))
}

func (r *TeamRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, "Importing TeamRole resource")

	// Import using composite ID format: {teamId}/{identifier}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// We need to parse the composite ID to extract team_id and id
	// The user should provide team_id in their configuration
	// For now, we'll use simple ID import and require team_id in config
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// teamRoleRightsValidator validates that no right appears in both granted_rights and disallowed_rights
type teamRoleRightsValidator struct{}

func (v teamRoleRightsValidator) Description(ctx context.Context) string {
	return "Validates that no right appears in both granted_rights and disallowed_rights"
}

func (v teamRoleRightsValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that no right appears in both granted_rights and disallowed_rights"
}

func (v teamRoleRightsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config dataModels.TeamRoleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.GrantedRights.IsNull() || config.GrantedRights.IsUnknown() || config.DisallowedRights.IsNull() || config.DisallowedRights.IsUnknown() {
		return
	}

	var granted []string
	var disallowed []string

	config.GrantedRights.ElementsAs(ctx, &granted, false)
	config.DisallowedRights.ElementsAs(ctx, &disallowed, false)

	// Build a map of granted rights for quick lookup
	grantedMap := make(map[string]bool)
	for _, right := range granted {
		grantedMap[right] = true
	}

	// Check for overlaps
	var overlapping []string
	for _, right := range disallowed {
		if grantedMap[right] {
			overlapping = append(overlapping, right)
		}
	}

	if len(overlapping) > 0 {
		resp.Diagnostics.AddError(
			"Invalid Team Role Configuration",
			fmt.Sprintf("The following rights appear in both granted_rights and disallowed_rights: %v. Each right must appear in only one list.", overlapping),
		)
	}
}

