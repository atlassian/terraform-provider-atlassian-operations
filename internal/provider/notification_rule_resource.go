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

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NotificationRuleResource{}
var _ resource.ResourceWithImportState = &NotificationRuleResource{}

func NewNotificationRuleResource() resource.Resource {
	return &NotificationRuleResource{}
}

// NotificationRuleResource defines the resource implementation.
type NotificationRuleResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func (r *NotificationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_rule"
}

func (r *NotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.NotificationRuleResourceAttributes,
	}
}

func (r *NotificationRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring NotificationRuleResource")

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
	tflog.Trace(ctx, "Configured NotificationRuleResource")
}

func (r *NotificationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating NotificationRuleResource")

	var data dataModels.NotificationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	notificationRuleDto, err := NotificationRuleModelToDto(ctx, data)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Configuration Error. Unable to convert notification rule model to DTO, got error: %s", err))
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to convert notification rule model to DTO, got error: %s", err))
		return
	}

	// Create notification rule
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl("/v1/notification-rules").
		Method(httpClient.POST).
		SetBody(notificationRuleDto).
		SetBodyParseObject(&notificationRuleDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create notification rule, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create notification rule, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create notification rule, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification rule, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create notification rule, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification rule, got error: %s", err))
		return
	}

	// Update state with response
	data = NotificationRuleDtoToModel(ctx, notificationRuleDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.NotificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading NotificationRuleResource")

	var notificationRuleDto dto.NotificationRuleDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/notification-rules/%s", data.ID.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&notificationRuleDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read notification rule, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read notification rule, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read notification rule, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification rule, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read notification rule, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification rule or to parse received data, got error: %s", err))
		return
	}

	data = NotificationRuleDtoToModel(ctx, notificationRuleDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.NotificationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	notificationRuleDto, err := NotificationRuleModelToDto(ctx, data)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Configuration Error. Unable to convert notification rule model to DTO, got error: %s", err))
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to convert notification rule model to DTO, got error: %s", err))
		return
	}

	// Update notification rule
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/notification-rules/%s", data.ID.ValueString())).
		Method(httpClient.PATCH).
		SetBody(notificationRuleDto).
		SetBodyParseObject(&notificationRuleDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update notification rule, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update notification rule, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update notification rule, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification rule, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update notification rule, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification rule, got error: %s", err))
		return
	}

	data = NotificationRuleDtoToModel(ctx, notificationRuleDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.NotificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete notification rule
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/notification-rules/%s", data.ID.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete notification rule, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete notification rule, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification rule, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete notification rule, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification rule, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete notification rule, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification rule, got error: %s", err))
		return
	}
}

func (r *NotificationRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
