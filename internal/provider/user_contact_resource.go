package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"

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
	_ resource.Resource                = &UserContactResource{}
	_ resource.ResourceWithConfigure   = &UserContactResource{}
	_ resource.ResourceWithImportState = &UserContactResource{}
)

type UserContactResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func NewUserContactResource() resource.Resource {
	return &UserContactResource{}
}

func (r *UserContactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_contact"
}

func (r *UserContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.UserContactResourceAttributes,
	}
}

func (r *UserContactResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring UserContactResource")

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
	tflog.Trace(ctx, "Configured UserContactResource")
}

func (r *UserContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating UserContactResource")

	var data dataModels.UserContactModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	contactDto := UserContactModelToDto(&data)

	// Create user contact
	var responseDto dto.UserContactCUDResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl("/v1/users/contacts").
		Method(httpClient.POST).
		SetBody(contactDto).
		SetBodyParseObject(&responseDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create user contact, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create user contact, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create user contact, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user contact, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create user contact, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user contact, got error: %s", err))
		return
	}

	// Update state with response
	result := UserContactCUDDtoToModel(&responseDto, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *UserContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.UserContactModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Reading UserContactResource")

	var responseDto dto.UserContactDataReadResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/users/contacts/%s", data.ID.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&responseDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read user contact, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read user contact, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user contact, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user contact, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user contact, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user contact or to parse received data, got error: %s", err))
		return
	}

	result := UserContactReadDtoToModel(&responseDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *UserContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.UserContactModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	contactDto := UserContactModelToDto(&data)

	methods, err := r.updateMethodFinder(ctx, &data, resp)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, got error: %s", err))
		return
	}

	if methods == nil {
		tflog.Info(ctx, "No changes detected. Skipping update.")
		return
	}

	var responseDto dto.UserContactCUDResponseDto
	// Check if the contact to field has changed
	if slices.Contains(methods, "patch") {
		httpResp, err := httpClientHelpers.
			GenerateJsmOpsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("/v1/users/contacts/%s", data.ID.ValueString())).
			Method(httpClient.PATCH).
			SetBody(contactDto).
			SetBodyParseObject(&responseDto).
			Send()
		err = updateClientErrorHandler(ctx, httpResp, err, resp)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, got error: %s", err))
			return
		}
	}
	// Check if the contact is activated
	if slices.Contains(methods, "activate") {
		httpResp, err := httpClientHelpers.
			GenerateJsmOpsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("/v1/users/contacts/%s/activate", data.ID.ValueString())).
			Method(httpClient.PATCH).
			SetBodyParseObject(&responseDto).
			Send()
		err = updateClientErrorHandler(ctx, httpResp, err, resp)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, got error: %s", err))
			return
		}
	}
	// Check if the contact is deactivated
	if slices.Contains(methods, "deactivate") {
		httpResp, err := httpClientHelpers.
			GenerateJsmOpsClientRequest(r.clientConfiguration).
			JoinBaseUrl(fmt.Sprintf("/v1/users/contacts/%s/deactivate", data.ID.ValueString())).
			Method(httpClient.PATCH).
			SetBodyParseObject(&responseDto).
			Send()
		err = updateClientErrorHandler(ctx, httpResp, err, resp)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, got error: %s", err))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, got error: %s", err))
			return
		}
	}

	result := UserContactCUDDtoToModel(&responseDto, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *UserContactResource) updateMethodFinder(ctx context.Context, data *dataModels.UserContactModel, resp *resource.UpdateResponse) ([]string, error) {
	var responseDto dto.UserContactDataReadResponseDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/users/contacts/%s", data.ID.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&responseDto).
		Send()

	updateError := updateClientErrorHandler(ctx, httpResp, err, resp)
	if updateError != nil {
		return []string{}, updateError
	}

	var changedFields []string
	// Check if the contact is activated
	if data.Enabled.ValueBool() && !responseDto.Status.Enabled {
		changedFields = append(changedFields, "activate")
	} else if !data.Enabled.ValueBool() && responseDto.Status.Enabled {
		changedFields = append(changedFields, "deactivate")
	}

	// Check if the contact to field has changed
	if data.To.ValueString() != responseDto.To {
		changedFields = append(changedFields, "patch")
	}

	return changedFields, nil
}

func updateClientErrorHandler(ctx context.Context, httpResp *httpClient.Response, err error, resp *resource.UpdateResponse) error {
	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update user contact, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update user contact, got nil response")
		return errors.New("Unable to update user contact, got nil response")
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, got http response: %d", statusCode))
		}
		return errors.New(fmt.Sprintf("Unable to update user contact, got http response: %d", statusCode))
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update user contact, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user contact, got error: %s", err))
		return errors.New(fmt.Sprintf("Unable to update user contact, got error: %s", err))
	}
	return nil
}

func (r *UserContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.UserContactModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/users/contacts/%s", data.ID.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete user contact, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete user contact, got nil response")
		return
	}

	if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user contact, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete user contact, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user contact, got http response: %d", statusCode))
		}
		return
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete user contact, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user contact, got error: %s", err))
		return
	}
}

func (r *UserContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
