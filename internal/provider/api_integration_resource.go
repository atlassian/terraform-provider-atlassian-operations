// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApiIntegrationResource{}
var _ resource.ResourceWithImportState = &ApiIntegrationResource{}

func NewApiIntegrationResource() resource.Resource {
	return &ApiIntegrationResource{}
}

// ApiIntegrationResource defines the resource implementation.
type ApiIntegrationResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func (r *ApiIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_integration"
}

func (r *ApiIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.ApiIntegrationResourceAttributes,
	}
}

func (r *ApiIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring ApiIntegrationResource")

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

	tflog.Trace(ctx, "Configured ApiIntegrationResource")
}

func (r *ApiIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating the ApiIntegrationResource")

	var data dataModels.ApiIntegrationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	dtoObj := ApiIntegrationModelToDto(ctx, data)

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl("v1/integrations").
		Method(httpClient.POST).
		SetBody(dtoObj).
		SetBodyParseObject(&dtoObj).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create api integration, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create api integration, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create api integration, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create api integration, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create api integration, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create api integration, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if data.DeleteDefaultActions.ValueBool() {
		// List default actions using the API Integration ID then using delete action endpoint delete each action
		tflog.Debug(ctx, fmt.Sprintf("Attempting to delete default actions for API integration: %s", dtoObj.Id))
		err = listDefaultActionsAndDelete(ctx, r.clientConfiguration, dtoObj.Id)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Error deleting default actions for API integration: %s", err))
			resp.Diagnostics.AddWarning("Error Deleting Default Actions", fmt.Sprintf("Unable to delete default actions for API integration: %s", err))
		}
	}

	data = ApiIntegrationDtoToModel(dtoObj, data)

	tflog.Trace(ctx, "Created the ApiIntegrationResource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ApiIntegrationResource into Terraform state")
}

func listDefaultActionsAndDelete(ctx context.Context, configuration dto.AtlassianOpsProviderModel, integrationId string) error {
	// Retry configuration for handling race condition where default actions
	// may not be immediately available after integration creation.
	// We use a manual retry loop because the HTTP client's AddRetryCondition
	// cannot read the response body without closing it (which breaks subsequent reads).
	const maxRetries = 5
	const retryDelaySeconds = 2

	var defaultActions dto.IntegrationActionListDto
	var lastErr error
	var lastHttpResp *httpClient.Response

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Default actions list is empty for integration %s, retrying in %d seconds (attempt %d/%d)", integrationId, retryDelaySeconds, attempt+1, maxRetries))
			time.Sleep(time.Duration(retryDelaySeconds) * time.Second)
		}

		defaultActions = dto.IntegrationActionListDto{}
		httpResp, err := httpClientHelpers.
			GenerateJsmOpsClientRequest(configuration).
			JoinBaseUrl(fmt.Sprintf("v1/integrations/%s/actions", integrationId)).
			Method(httpClient.GET).
			SetBodyParseObject(&defaultActions).
			Send()

		lastHttpResp = httpResp
		lastErr = err

		if err != nil {
			tflog.Debug(ctx, fmt.Sprintf("Error listing default actions for integration %s: %s", integrationId, err))
			continue
		}
		if httpResp.IsError() {
			tflog.Debug(ctx, fmt.Sprintf("HTTP error listing default actions for integration %s: %d", integrationId, httpResp.GetStatusCode()))
			continue
		}

		// If we found actions, break out of retry loop
		if len(defaultActions.Values) > 0 {
			break
		}
	}

	if lastErr != nil {
		return fmt.Errorf("unable to list default actions, got error: %s. Couldn't delete default actions automatically. Please delete it through UI", lastErr)
	}
	if lastHttpResp != nil && lastHttpResp.IsError() {
		return fmt.Errorf("unable to list default actions, got http response: %d. Couldn't delete default actions automatically. Please delete it through UI", lastHttpResp.GetStatusCode())
	}

	// If no actions found after retries, it's normal for some integration types
	if len(defaultActions.Values) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("No default actions found for integration %s after retries, nothing to delete", integrationId))
		return nil
	}

	tflog.Debug(ctx, fmt.Sprintf("Found %d default actions to delete for integration %s", len(defaultActions.Values), integrationId))

	// Delete all found default actions
	for _, action := range defaultActions.Values {
		tflog.Debug(ctx, fmt.Sprintf("Deleting default action %s (%s) for integration %s", action.Name, action.ID, integrationId))
		httpResp, err := httpClientHelpers.
			GenerateJsmOpsClientRequest(configuration).
			JoinBaseUrl(fmt.Sprintf("v1/integrations/%s/actions/%s", integrationId, action.ID)).
			Method(httpClient.DELETE).
			Send()
		if err != nil {
			return fmt.Errorf("unable to delete default action %s, got error: %s. Couldn't delete default actions automatically. Please delete it through UI", action.ID, err)
		}
		if httpResp.IsError() {
			return fmt.Errorf("unable to delete default action %s, got http response: %d. Couldn't delete default actions automatically. Please delete it through UI", action.Name, httpResp.GetStatusCode())
		}
		tflog.Debug(ctx, fmt.Sprintf("Successfully deleted default action %s (%s) for integration %s", action.Name, action.ID, integrationId))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted all %d default actions for integration %s", len(defaultActions.Values), integrationId))
	return nil
}

func (r *ApiIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.ApiIntegrationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Reading the ApiIntegrationResource")

	ApiIntegration := dto.ApiIntegration{}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/integrations/%s", data.Id.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&ApiIntegration).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read api integration, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read api integration, got nil response")
	} else if httpResp.GetStatusCode() == 404 {
		resp.State.RemoveResource(ctx)

		return
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read api integration, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api integration, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read api integration, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read api integration or to parse received data, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data = ApiIntegrationDtoToModel(ApiIntegration, data)

	tflog.Trace(ctx, "Read the ApiIntegrationResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ApiIntegrationResource into Terraform state")
}

func (r *ApiIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.ApiIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	tflog.Trace(ctx, "Updating the ApiIntegrationResource")

	dtoObj := ApiIntegrationModelToDto(ctx, data)

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/integrations/%s", data.Id.ValueString())).
		Method(httpClient.PATCH).
		SetBody(dtoObj).
		SetBodyParseObject(&dtoObj).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update api integration, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update api integration, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update api integration, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update api integration, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update api integration, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update api integration, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data = ApiIntegrationDtoToModel(dtoObj, data)

	tflog.Trace(ctx, "Updated the ApiIntegrationResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ApiIntegrationResource into Terraform state")
}

func (r *ApiIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.ApiIntegrationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Deleting the ApiIntegrationResource")

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/integrations/%s", data.Id.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete api integration, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete api integration, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete api integration, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete api integration, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete api integration, got http response: %d", statusCode))
		}
	} else if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete api integration, got http response: %d", httpResp.GetStatusCode()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete api integration, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleted the ApiIntegrationResource")
}

func (r *ApiIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
