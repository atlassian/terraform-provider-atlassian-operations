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
	_ resource.Resource                = &AlertPolicyResource{}
	_ resource.ResourceWithConfigure   = &AlertPolicyResource{}
	_ resource.ResourceWithImportState = &AlertPolicyResource{}
)

type AlertPolicyResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func NewAlertPolicyResource() resource.Resource {
	return &AlertPolicyResource{}
}

func (r *AlertPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_policy"
}

func (r *AlertPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.AlertPolicyResourceAttributes,
	}
}

func (r *AlertPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring AlertPolicyResource")

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
	tflog.Trace(ctx, "Configured AlertPolicyResource")
}

func (r *AlertPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating AlertPolicyResource")

	var data dataModels.AlertPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	alertPolicyDto, _ := AlertPolicyModelToDto(ctx, &data)

	var createBaseUrl string
	if data.TeamID.IsUnknown() || data.TeamID.IsNull() {
		createBaseUrl = "/v1/alerts/policies"
	} else {
		createBaseUrl = fmt.Sprintf("/v1/teams/%s/policies", data.TeamID.ValueString())
	}

	// Create alert policy
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(createBaseUrl).
		Method(httpClient.POST).
		SetBody(alertPolicyDto).
		SetBodyParseObject(&alertPolicyDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create alert policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create alert policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create alert policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create alert policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert policy, got error: %s", err))
		return
	}

	if !data.Order.IsUnknown() && !data.Order.IsNull() {
		requestedOrder := int32(data.Order.ValueInt64())
		var teamId string
		if !data.TeamID.IsUnknown() && !data.TeamID.IsNull() {
			teamId = data.TeamID.ValueString()
		}
		
		r.updatePolicyOrder(ctx, teamId, alertPolicyDto.ID, requestedOrder)
	}

	order := getAlertPolicyOrder(ctx, r.clientConfiguration, data.TeamID.ValueString(), alertPolicyDto.ID)
	// Update state with response
	result, _ := AlertPolicyDtoToModel(ctx, order, alertPolicyDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *AlertPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.AlertPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading AlertPolicyResource")

	var alertPolicyDto dto.AlertPolicyDto
	var readBaseUrl string

	if data.TeamID.IsUnknown() || data.TeamID.IsNull() {
		readBaseUrl = fmt.Sprintf("/v1/alerts/policies/%s", data.ID.ValueString())
	} else {
		readBaseUrl = fmt.Sprintf("/v1/teams/%s/policies/%s", data.TeamID.ValueString(), data.ID.ValueString())
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(readBaseUrl).
		Method(httpClient.GET).
		SetBodyParseObject(&alertPolicyDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read alert policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read alert policy, got nil response")
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
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read alert policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read alert policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert policy or to parse received data, got error: %s", err))
		return
	}

	order := getAlertPolicyOrder(ctx, r.clientConfiguration, data.TeamID.ValueString(), alertPolicyDto.ID)

	result, _ := AlertPolicyDtoToModel(ctx, order, &alertPolicyDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *AlertPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.AlertPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	alertPolicyDto, _ := AlertPolicyModelToDto(ctx, &data)

	var updateBaseUrl string
	if data.TeamID.IsUnknown() || data.TeamID.IsNull() {
		updateBaseUrl = fmt.Sprintf("/v1/alerts/policies/%s", data.ID.ValueString())
	} else {
		updateBaseUrl = fmt.Sprintf("/v1/teams/%s/policies/%s", data.TeamID.ValueString(), data.ID.ValueString())
	}

	// Update alert policy
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(updateBaseUrl).
		Method(httpClient.PUT).
		SetBody(alertPolicyDto).
		SetBodyParseObject(&alertPolicyDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update alert policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update alert policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update alert policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update alert policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert policy, got error: %s", err))
		return
	}
	
	if !data.Order.IsUnknown() && !data.Order.IsNull() {
		requestedOrder := int32(data.Order.ValueInt64())
		var teamId string
		if !data.TeamID.IsUnknown() && !data.TeamID.IsNull() {
			teamId = data.TeamID.ValueString()
		}
		
		r.updatePolicyOrder(ctx, teamId, alertPolicyDto.ID, requestedOrder)
	}
	
	order := getAlertPolicyOrder(ctx, r.clientConfiguration, data.TeamID.ValueString(), alertPolicyDto.ID)
	result, _ := AlertPolicyDtoToModel(ctx, order, alertPolicyDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *AlertPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.AlertPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deleteBaseUrl string
	if data.TeamID.IsUnknown() || data.TeamID.IsNull() {
		deleteBaseUrl = fmt.Sprintf("/v1/alerts/policies/%s", data.ID.ValueString())
	} else {
		deleteBaseUrl = fmt.Sprintf("/v1/teams/%s/policies/%s", data.TeamID.ValueString(), data.ID.ValueString())
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(deleteBaseUrl).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete alert policy, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete alert policy, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete alert policy, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert policy, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete alert policy, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert policy, got error: %s", err))
		return
	}
}

func (r *AlertPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) > 2 || idParts[0] == "" || (len(idParts) == 2 && idParts[1] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,team_id (for team policies); or: id (for global policies). Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), idParts[1])...)
	}
}

func getAlertPolicyOrder(ctx context.Context, configuration dto.AtlassianOpsProviderModel, teamId string, alertPolicyId string) int64 {
	// list alert policies find the one we just created, and get its order value
	listAlertPoliciesResponse := &dto.AlertPolicyListDto{}
	baseURL := fmt.Sprintf("/v1/teams/%s/policies", teamId)
	queryParams := map[string]string{
		"type": "alert",
	}
	var order int64
	doneLooping := false

	for !doneLooping {
		req := httpClientHelpers.
			GenerateJsmOpsClientRequest(configuration).
			JoinBaseUrl(baseURL).
			Method(httpClient.GET).
			SetBodyParseObject(&listAlertPoliciesResponse).
			SetQueryParams(queryParams)

		httpResp, err := req.Send()

		if httpResp == nil {
			tflog.Error(ctx, "Client Error. Unable to list alert policies, got nil response")
			return 0
		}
		if httpResp.IsError() {
			statusCode := httpResp.GetStatusCode()
			errorResponse := httpResp.GetErrorBody()
			if errorResponse != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to list alert policies, status code: %d. Got response: %s", statusCode, *errorResponse))
			} else {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to list alert policies, got http response: %d", statusCode))
			}
			return 0
		}
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to list alert policies, got error: %s", err))
			return 0
		}

		for _, policy := range listAlertPoliciesResponse.Values {
			if policy.ID == alertPolicyId {
				order = int64(policy.Order)
				return order
			}
		}

		if listAlertPoliciesResponse.Links.Next == "" {
			doneLooping = true
		} else {
			nextURL := listAlertPoliciesResponse.Links.Next
			parsedURL, err := url.Parse(nextURL)
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to parse next URL, got error: %s", err))
				return 0
			}
			urlValues := parsedURL.Query()
			queryParams = make(map[string]string)
			for key, values := range urlValues {
				if len(values) > 0 {
					queryParams[key] = values[0]
				}
			}
			baseURL = parsedURL.Path
			listAlertPoliciesResponse = &dto.AlertPolicyListDto{}
		}
	}
	return order
}

func (r *AlertPolicyResource) updatePolicyOrder(ctx context.Context, teamId string, policyId string, requestedOrder int32) {
	var orderBaseUrl string
	if teamId == "" {
		orderBaseUrl = fmt.Sprintf("/v1/alerts/policies/%s/change-order", policyId)
	} else {
		orderBaseUrl = fmt.Sprintf("/v1/teams/%s/policies/%s/change-order", teamId, policyId)
	}
	
	orderDto := map[string]int32{"order": requestedOrder}

	orderResp, orderErr := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(orderBaseUrl).
		Method(httpClient.POST).
		SetBody(orderDto).
		Send()

	if orderResp.IsError() {
		statusCode := orderResp.GetStatusCode()
		errorResponse := orderResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to set order for alert policy, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to set order for alert policy, got http response: %d", statusCode))
		}
	}
	if orderErr != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to set order for alert policy, got error: %s", orderErr))
	}
}
