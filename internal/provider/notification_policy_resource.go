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
	_ resource.Resource                = &NotificationPolicyResource{}
	_ resource.ResourceWithConfigure   = &NotificationPolicyResource{}
	_ resource.ResourceWithImportState = &NotificationPolicyResource{}
)

type NotificationPolicyResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func NewNotificationPolicyResource() resource.Resource {
	return &NotificationPolicyResource{}
}

func (r *NotificationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_policy"
}

func (r *NotificationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.NotificationPolicyResourceAttributes,
	}
}

func (r *NotificationPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring NotificationPolicyResource")

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
	tflog.Trace(ctx, "Configured NotificationPolicyResource")
}

func (r *NotificationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating NotificationPolicyResource")

	var data dataModels.NotificationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	notificationPolicyDto, _ := NotificationPolicyModelToDto(ctx, &data)

	// Create notification policy
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/policies", data.TeamID.ValueString())).
		Method(httpClient.POST).
		SetBody(notificationPolicyDto).
		SetBodyParseObject(&notificationPolicyDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create notification policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create notification policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create notification policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create notification policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification policy, got error: %s", err))
		return
	}

	// Update state with response
	result, _ := NotificationPolicyDtoToModel(ctx, notificationPolicyDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *NotificationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.NotificationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading NotificationPolicyResource")

	var notificationPolicyDto dto.NotificationPolicyDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/policies/%s", data.TeamID.ValueString(), data.ID.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&notificationPolicyDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read notification policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read notification policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read notification policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read notification policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification policy or to parse received data, got error: %s", err))
		return
	}

	result, _ := NotificationPolicyDtoToModel(ctx, &notificationPolicyDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *NotificationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.NotificationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	notificationPolicyDto, _ := NotificationPolicyModelToDto(ctx, &data)

	// Update notification policy
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/policies/%s", data.TeamID.ValueString(), data.ID.ValueString())).
		Method(httpClient.PUT).
		SetBody(notificationPolicyDto).
		SetBodyParseObject(&notificationPolicyDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update notification policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update notification policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update notification policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update notification policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification policy, got error: %s", err))
		return
	}

	result, _ := NotificationPolicyDtoToModel(ctx, notificationPolicyDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *NotificationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.NotificationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/policies/%s", data.TeamID.ValueString(), data.ID.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete notification policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete notification policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete notification policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete notification policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification policy, got error: %s", err))
		return
	}
}

func (r *NotificationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,team_id. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), idParts[1])...)
}
