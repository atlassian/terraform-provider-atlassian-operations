// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ScheduleRotationResource{}
var _ resource.ResourceWithImportState = &ScheduleRotationResource{}

func NewScheduleRotationResource() resource.Resource {
	return &ScheduleRotationResource{}
}

// ScheduleRotationResource defines the resource implementation.
type ScheduleRotationResource struct {
	client *httpClient.HttpClient
}

func (r *ScheduleRotationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule_rotation"
}

func (r *ScheduleRotationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.RotationResourceAttributes,
	}
}

func (r *ScheduleRotationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring ScheduleRotationResource")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*JsmOpsClient)

	if !ok {
		tflog.Error(ctx, "Unexpected Resource Configure Type")
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client.OpsClient

	tflog.Trace(ctx, "Configured ScheduleRotationResource")
}

func (r *ScheduleRotationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating the ScheduleRotationResource")

	var data dataModels.RotationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	rotationDto := RotationModelToDto(ctx, data)
	errorMap := httpClient.NewOpsClientErrorMap()

	httpResp, err := r.client.NewRequest().
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations", data.ScheduleId.ValueString())).
		Method(httpClient.POST).
		SetBody(rotationDto).
		SetBodyParseObject(&rotationDto).
		SetErrorParseMap(&errorMap).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := errorMap[statusCode]
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotation schedule, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotation schedule, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data = RotationDtoToModel(data.ScheduleId.ValueString(), rotationDto)

	tflog.Trace(ctx, "Created the ScheduleRotationResource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ScheduleRotationResource into Terraform state")
}

func (r *ScheduleRotationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.RotationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Reading the ScheduleRotationResource")

	rotationDto := dto.Rotation{}
	errorMap := httpClient.NewOpsClientErrorMap()

	httpResp, err := r.client.NewRequest().
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", data.ScheduleId.ValueString(), data.Id.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&rotationDto).
		SetErrorParseMap(&errorMap).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := errorMap[statusCode]
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read rotation, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read rotation, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation or to parse received data, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data = RotationDtoToModel(data.ScheduleId.ValueString(), rotationDto)

	tflog.Trace(ctx, "Read the ScheduleRotationResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ScheduleRotationResource into Terraform state")
}

func (r *ScheduleRotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.RotationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	tflog.Trace(ctx, "Updating the ScheduleRotationResource")

	rotationDto := RotationModelToDto(ctx, data)
	errorMap := httpClient.NewOpsClientErrorMap()

	httpResp, err := r.client.NewRequest().
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", data.ScheduleId.ValueString(), data.Id.ValueString())).
		Method(httpClient.PATCH).
		SetBody(rotationDto).
		SetBodyParseObject(&rotationDto).
		SetErrorParseMap(&errorMap).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := errorMap[statusCode]
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update rotation, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update rotation, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data = RotationDtoToModel(data.ScheduleId.ValueString(), rotationDto)

	tflog.Trace(ctx, "Updated the ScheduleRotationResource")

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ScheduleRotationResource into Terraform state")
}

func (r *ScheduleRotationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.RotationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Deleting the ScheduleRotationResource")
	errorMap := httpClient.NewOpsClientErrorMap()

	httpResp, err := r.client.NewRequest().
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", data.ScheduleId.ValueString(), data.Id.ValueString())).
		Method(httpClient.DELETE).
		SetErrorParseMap(&errorMap).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := errorMap[statusCode]
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete rotation, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation, got http response: %d", statusCode))
		}
	}
	if httpResp != nil && err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete rotation, got http response: %d", httpResp.GetStatusCode()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation, got error: %s", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleted the ScheduleRotationResource")
}

func (r *ScheduleRotationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,schedule_id. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schedule_id"), idParts[1])...)
}
