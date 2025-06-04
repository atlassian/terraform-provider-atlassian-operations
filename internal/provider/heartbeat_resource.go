package provider

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &HeartbeatResource{}
	_ resource.ResourceWithConfigure   = &HeartbeatResource{}
	_ resource.ResourceWithImportState = &HeartbeatResource{}
)

type HeartbeatResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func NewHeartbeatResource() resource.Resource {
	return &HeartbeatResource{}
}

func (r *HeartbeatResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_heartbeat"
}

func (r *HeartbeatResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage heartbeats in Atlassian Operations.",
		Attributes:  schemaAttributes.HeartbeatResourceAttributes,
	}
}

func (r *HeartbeatResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring HeartbeatResource")

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client
	tflog.Trace(ctx, "Configured HeartbeatResource")
}

func (r *HeartbeatResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating HeartbeatResource")

	var data dataModels.HeartbeatModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	heartbeatDto, diags := HeartbeatModelToDto(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create heartbeat
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/heartbeats", data.TeamID.ValueString())).
		Method(httpClient.POST).
		SetBody(heartbeatDto).
		SetBodyParseObject(&heartbeatDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create heartbeat, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create heartbeat, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create heartbeat, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create heartbeat, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create heartbeat, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create heartbeat, got error: %s", err))
		return
	}

	// Update state with response
	result, diags := HeartbeatDtoToModel(ctx, heartbeatDto, data.TeamID.ValueString())
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *HeartbeatResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.HeartbeatModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading HeartbeatResource")

	// Get heartbeats and find the one with the specified name
	var heartbeatPaginatedResponseDto dto.HeartbeatPaginatedResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/heartbeats", data.TeamID.ValueString())).
		Method(httpClient.GET).
		SetQueryParam("name", data.Name.ValueString()).
		SetBodyParseObject(&heartbeatPaginatedResponseDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read heartbeat, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read heartbeat, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read heartbeat, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read heartbeat, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read heartbeat, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read heartbeat or to parse received data, got error: %s", err))
		return
	}

	// Find the heartbeat with the matching name
	var heartbeatDto *dto.HeartbeatDto
	for _, hb := range heartbeatPaginatedResponseDto.Values {
		if hb.Name == data.Name.ValueString() {
			heartbeatDto = &hb
			break
		}
	}

	if heartbeatDto == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	result, diags := HeartbeatDtoToModel(ctx, heartbeatDto, data.TeamID.ValueString())
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *HeartbeatResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.HeartbeatModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	heartbeatDto, diags := HeartbeatModelToDto(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update heartbeat
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/heartbeats", data.TeamID.ValueString())).
		Method(httpClient.PATCH).
		SetQueryParam("name", data.Name.ValueString()).
		SetBody(heartbeatDto).
		SetBodyParseObject(&heartbeatDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update heartbeat, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update heartbeat, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update heartbeat, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update heartbeat, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update heartbeat, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update heartbeat, got error: %s", err))
		return
	}

	result, diags := HeartbeatDtoToModel(ctx, heartbeatDto, data.TeamID.ValueString())
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *HeartbeatResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.HeartbeatModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/heartbeats", data.TeamID.ValueString())).
		Method(httpClient.DELETE).
		SetQueryParam("name", data.Name.ValueString()).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete heartbeat, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete heartbeat, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete heartbeat, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete heartbeat, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete heartbeat, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete heartbeat, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete heartbeat, got error: %s", err))
		return
	}
}

func (r *HeartbeatResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,team_id. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), idParts[1])...)
}
