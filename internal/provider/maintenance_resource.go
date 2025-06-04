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
	_ resource.Resource                = &MaintenanceResource{}
	_ resource.ResourceWithConfigure   = &MaintenanceResource{}
	_ resource.ResourceWithImportState = &MaintenanceResource{}
)

// MaintenanceResource defines the resource implementation for maintenances
type MaintenanceResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func NewMaintenanceResource() resource.Resource {
	return &MaintenanceResource{}
}

// Metadata returns metadata for the resource
func (r *MaintenanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

// Schema defines the schema for the resource
func (r *MaintenanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage maintenance windows in Atlassian Operations.",
		Attributes:  schemaAttributes.MaintenanceResourceAttributes,
	}
}

// Configure sets up the resource with provider configuration
func (r *MaintenanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring MaintenanceResource")

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected dto.JsmopsProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client
	tflog.Trace(ctx, "Configured MaintenanceResource")
}

// Create handles the create operation for the resource
func (r *MaintenanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating MaintenanceResource")

	var plan dataModels.MaintenanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	maintenanceDto, diags := MaintenanceModelToDto(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create maintenance window
	var endpoint string
	if plan.TeamID.ValueString() != "" {
		endpoint = fmt.Sprintf("/v1/teams/%s/maintenances", plan.TeamID.ValueString())
	} else {
		endpoint = "/v1/maintenances"
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(endpoint).
		Method(httpClient.POST).
		SetBody(maintenanceDto).
		SetBodyParseObject(&maintenanceDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create maintenance window, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create maintenance window, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create maintenance window, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create maintenance window, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create maintenance window, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create maintenance window, got error: %s", err))
		return
	}

	// Update state with response
	result, diags := MaintenanceDtoToModel(ctx, maintenanceDto)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read handles the read operation for the resource
func (r *MaintenanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataModels.MaintenanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading MaintenanceResource")

	// Determine endpoint based on whether we have a team ID
	var endpoint string
	if state.TeamID.ValueString() != "" {
		endpoint = fmt.Sprintf("/v1/teams/%s/maintenances/%s", state.TeamID.ValueString(), state.ID.ValueString())
	} else {
		endpoint = fmt.Sprintf("/v1/maintenances/%s", state.ID.ValueString())
	}

	var maintenanceDto dto.MaintenanceDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(endpoint).
		Method(httpClient.GET).
		SetBodyParseObject(&maintenanceDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read maintenance window, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read maintenance window, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		if statusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read maintenance window, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read maintenance window, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read maintenance window, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read maintenance window or to parse received data, got error: %s", err))
		return
	}

	result, diags := MaintenanceDtoToModel(ctx, &maintenanceDto)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Update handles the update operation for the resource
func (r *MaintenanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dataModels.MaintenanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	maintenanceDto, diags := MaintenanceModelToDto(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine endpoint based on whether we have a team ID
	var endpoint string
	if plan.TeamID.ValueString() != "" {
		endpoint = fmt.Sprintf("/v1/teams/%s/maintenances/%s", plan.TeamID.ValueString(), plan.ID.ValueString())
	} else {
		endpoint = fmt.Sprintf("/v1/maintenances/%s", plan.ID.ValueString())
	}

	// Update maintenance window
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(endpoint).
		Method(httpClient.PATCH).
		SetBody(maintenanceDto).
		SetBodyParseObject(&maintenanceDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update maintenance window, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update maintenance window, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update maintenance window, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update maintenance window, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update maintenance window, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update maintenance window, got error: %s", err))
		return
	}

	result, diags := MaintenanceDtoToModel(ctx, maintenanceDto)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Delete handles the delete operation for the resource
func (r *MaintenanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataModels.MaintenanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine endpoint based on whether we have a team ID
	var endpoint string
	if state.TeamID.ValueString() != "" {
		endpoint = fmt.Sprintf("/v1/teams/%s/maintenances/%s", state.TeamID.ValueString(), state.ID.ValueString())
	} else {
		endpoint = fmt.Sprintf("/v1/maintenances/%s", state.ID.ValueString())
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(endpoint).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete maintenance window, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete maintenance window, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete maintenance window, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete maintenance window, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete maintenance window, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete maintenance window, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete maintenance window, got error: %s", err))
		return
	}
}

// ImportState handles importing the state of an existing resource
func (r *MaintenanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 1 && len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: maintenance_id or maintenance_id,team_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)

	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), idParts[1])...)
	}
}
