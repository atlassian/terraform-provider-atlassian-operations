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
var _ resource.Resource = &ServiceResource{}
var _ resource.ResourceWithImportState = &ServiceResource{}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

// ServiceResource defines the resource implementation.
type ServiceResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func (r *ServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource only works for Jira Service Management services. " +
			"[Read more about services](https://support.atlassian.com/jira-service-management-cloud/docs/what-is-services/).",
		Attributes: schemaAttributes.ServiceResourceAttributes,
	}
}

func (r *ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring ServiceResource")

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected dto.AtlassianOpsProviderModel, got: %T", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client
	tflog.Trace(ctx, "Configured ServiceResource")
}

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating ServiceResource")

	var data dataModels.ServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	ServiceDto, diags := ServiceModelToDto(ctx, &data, r.clientConfiguration.GetCloudId())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create JSM service
	httpResp, err := httpClientHelpers.
		GenerateServiceClientRequest(r.clientConfiguration).
		JoinBaseUrl("/v1/services").
		Method(httpClient.POST).
		SetBody(ServiceDto).
		SetBodyParseObject(&ServiceDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create JSM service, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create JSM service, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create JSM service, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create JSM service, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create JSM service, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create JSM service, got error: %s", err))
		return
	}

	// Update state with response
	modelPtr, diags := ServiceDtoToModel(ctx, ServiceDto)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data = *modelPtr
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.ServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading ServiceResource")

	var ServiceDto dto.ServiceDto
	httpResp, err := httpClientHelpers.
		GenerateServiceClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/services/%s", data.ID.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&ServiceDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read JSM service, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read JSM service, got nil response")
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
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read JSM service, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read JSM service, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read JSM service, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read JSM service or to parse received data, got error: %s", err))
		return
	}

	modelPtr, diags := ServiceDtoToModel(ctx, &ServiceDto)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data = *modelPtr
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.ServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	ServiceDto, diags := ServiceModelToDto(ctx, &data, r.clientConfiguration.GetCloudId())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update JSM service
	httpResp, err := httpClientHelpers.
		GenerateServiceClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/services/%s", data.ID.ValueString())).
		Method(httpClient.PATCH).
		SetBody(ServiceDto).
		SetBodyParseObject(&ServiceDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update JSM service, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update JSM service, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update JSM service, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update JSM service, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update JSM service, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update JSM service, got error: %s", err))
		return
	}

	modelPtr, diags := ServiceDtoToModel(ctx, ServiceDto)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data = *modelPtr
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.ServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete JSM service
	httpResp, err := httpClientHelpers.
		GenerateServiceClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/services/%s", data.ID.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete JSM service, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete JSM service, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete JSM service, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete JSM service, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete JSM service, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete JSM service, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete JSM service, got error: %s", err))
		return
	}
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For JSM service, we only need the service ID
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: service_id. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
