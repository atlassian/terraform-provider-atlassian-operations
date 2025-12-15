package provider

import (
	"context"
	"fmt"
	"net/url"
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

	// Set the order if specified by user
	if !data.Order.IsUnknown() && !data.Order.IsNull() {
		// Convert from 1-indexed (user input) to 0-indexed (API expects)
		// User provides: 1, 2, 3... → API expects: 0, 1, 2...
		requestedOrder := int32(data.Order.ValueFloat64() - 1)
		if requestedOrder < 0 {
			requestedOrder = 0
		}
		r.updatePolicyOrder(ctx, data.TeamID.ValueString(), notificationPolicyDto.ID, requestedOrder)
	}

	// list notification policies find the one we just created, and get its order value
	order := getNotificationPolicyOrder(ctx, r.clientConfiguration, data.TeamID.ValueString(), notificationPolicyDto.ID)

	// Update state with response
	result, _ := NotificationPolicyDtoToModel(ctx, order, notificationPolicyDto)
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

	if httpResp.GetStatusCode() == 404 {
		resp.State.RemoveResource(ctx)

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

	order := getNotificationPolicyOrder(ctx, r.clientConfiguration, data.TeamID.ValueString(), notificationPolicyDto.ID)

	result, _ := NotificationPolicyDtoToModel(ctx, order, &notificationPolicyDto)
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

	// Set the order if specified by user
	if !data.Order.IsUnknown() && !data.Order.IsNull() {
		// Convert from 1-indexed (user input) to 0-indexed (API expects)
		// User provides: 1, 2, 3... → API expects: 0, 1, 2...
		requestedOrder := int32(data.Order.ValueFloat64() - 1)
		if requestedOrder < 0 {
			requestedOrder = 0
		}
		r.updatePolicyOrder(ctx, data.TeamID.ValueString(), notificationPolicyDto.ID, requestedOrder)
	}

	order := getNotificationPolicyOrder(ctx, r.clientConfiguration, data.TeamID.ValueString(), notificationPolicyDto.ID)
	result, _ := NotificationPolicyDtoToModel(ctx, order, notificationPolicyDto)
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

func getNotificationPolicyOrder(ctx context.Context, configuration dto.AtlassianOpsProviderModel, teamId string, notificationPolicyId string) float64 {
	// list notification policies find the one we just created, and get its order value
	listNotificationPoliciesDto := &dto.NotificationPolicyListDto{}
	baseURL := fmt.Sprintf("/v1/teams/%s/policies", teamId)
	queryParams := map[string]string{
		"type": "notification",
	}
	var order float64
	doneLooping := false

	for !doneLooping {
		req := httpClientHelpers.
			GenerateJsmOpsClientRequest(configuration).
			JoinBaseUrl(baseURL).
			Method(httpClient.GET).
			SetBodyParseObject(&listNotificationPoliciesDto).
			SetQueryParams(queryParams)

		httpResp, err := req.Send()

		if httpResp == nil {
			tflog.Error(ctx, "Client Error. Unable to list notification policies, got nil response")
			return 0.0
		}
		if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to list notification policies, status code: %d. Got response: %s", statusCode, *errorResponse))
			} else {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to list notification policies, got http response: %d", statusCode))
			}
			return 0.0
		}
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to list notification policies, got error: %s", err))
			return 0.0
		}

		for _, policy := range listNotificationPoliciesDto.Values {
			if policy.ID == notificationPolicyId {
				order = policy.Order
				return order
			}
		}

		if listNotificationPoliciesDto.Links.Next == "" {
			doneLooping = true
		} else {
			nextURL := listNotificationPoliciesDto.Links.Next
			parsedURL, err := url.Parse(nextURL)
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to parse next URL, got error: %s", err))
				return 0.0
			}
			urlValues := parsedURL.Query()
			queryParams = make(map[string]string)
			for key, values := range urlValues {
				if len(values) > 0 {
					queryParams[key] = values[0]
				}
			}
			baseURL = parsedURL.Path
			listNotificationPoliciesDto = &dto.NotificationPolicyListDto{}
		}
	}
	return order
}

func (r *NotificationPolicyResource) updatePolicyOrder(ctx context.Context, teamId string, policyId string, requestedOrder int32) {
	orderBaseUrl := fmt.Sprintf("/v1/teams/%s/policies/%s/change-order", teamId, policyId)
	orderDto := map[string]int32{"order": requestedOrder}

	orderResp, orderErr := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(orderBaseUrl).
		Method(httpClient.POST).
		SetBody(orderDto).
		Send()

	if orderResp == nil {
		tflog.Error(ctx, "Client Error. Unable to set order for notification policy, got nil response")
		return
	}

	if orderResp.IsError() {
		statusCode := orderResp.GetStatusCode()
		errorResponse := orderResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to set order for notification policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to set order for notification policy, got http response: %d", statusCode))
		}
		return
	}

	if orderErr != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to set order for notification policy, got error: %s", orderErr))
		return
	}
}
