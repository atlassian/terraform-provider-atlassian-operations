// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
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
	"slices"
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ScheduleRotationResource{}
var _ resource.ResourceWithImportState = &ScheduleRotationResource{}

func NewScheduleRotationResource() resource.Resource {
	return &ScheduleRotationResource{}
}

// ScheduleRotationResource defines the resource implementation.
type ScheduleRotationResource struct {
	clientConfiguration dto.JsmopsProviderModel
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

	client, ok := req.ProviderData.(dto.JsmopsProviderModel)

	if !ok {
		tflog.Error(ctx, "Unexpected Resource Configure Type")
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client

	tflog.Trace(ctx, "Configured ScheduleRotationResource")
}

func (r *ScheduleRotationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "Creating the ScheduleRotationResource")

	var data dataModels.RotationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// We need to compare the participant lists of the initial config vs. the server response
	plannedDto := RotationModelToDto(ctx, data)
	rotationDto := RotationModelToDto(ctx, data)

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations", data.ScheduleId.ValueString())).
		Method(httpClient.POST).
		SetBody(rotationDto).
		SetBodyParseObject(&rotationDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to create rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to create rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotation schedule, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotation schedule, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation, got error: %s", err))
	}

	if !(resp.Diagnostics.HasError() || areUserListsEqual(plannedDto.Participants, rotationDto.Participants)) {
		plannedParticipants, _ := json.Marshal(plannedDto.Participants)
		newParticipants, _ := json.Marshal(rotationDto.Participants)
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation. The received participants list from the server does not match the one defined in Terraform configuration. Expected: %s, Got: %s", plannedParticipants, newParticipants))
		resp.Diagnostics.AddAttributeError(
			path.Root("participants"),
			"Unable to create rotation. The received participants list from the server does not match the one defined in Terraform configuration.",
			fmt.Sprintf(
				"Expected: %s, Got: %s.\n"+
					"This can be caused by the use of old Opsgenie UserIDs, instead of Atlassian Account IDs. The server automatically converts old IDs into new ones, which causes a state mismatch in Terraform.\n"+
					"Please consider checking the ID values you specified.", plannedParticipants, newParticipants,
			),
		)
	}

	if resp.Diagnostics.HasError() {
		cleanupRotationSilent(r, data.ScheduleId.ValueString(), rotationDto.Id)
		return
	}

	// HEIMDALL-12257 Workaround for time format from the server not matching the one in the config despite denoting the same time
	startDateRequest, _ := time.Parse(time.RFC3339, data.StartDate.ValueString())
	startDateResponse, _ := time.Parse(time.RFC3339, rotationDto.StartDate)

	endDateRequest, _ := time.Parse(time.RFC3339, data.EndDate.ValueString())
	endDateResponse, _ := time.Parse(time.RFC3339, rotationDto.EndDate)

	if startDateRequest.Equal(startDateResponse) {
		rotationDto.StartDate = data.StartDate.ValueString()
	}

	if endDateRequest.Equal(endDateResponse) {
		rotationDto.EndDate = data.EndDate.ValueString()
	}
	//

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

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", data.ScheduleId.ValueString(), data.Id.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&rotationDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to read rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to read rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
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

	// HEIMDALL-12257 Workaround for time format from the server not matching the one in the config despite denoting the same time
	startDateRequest, _ := time.Parse(time.RFC3339, data.StartDate.ValueString())
	startDateResponse, _ := time.Parse(time.RFC3339, rotationDto.StartDate)

	endDateRequest, _ := time.Parse(time.RFC3339, data.EndDate.ValueString())
	endDateResponse, _ := time.Parse(time.RFC3339, rotationDto.EndDate)

	if startDateRequest.Equal(startDateResponse) {
		rotationDto.StartDate = data.StartDate.ValueString()
	}

	if endDateRequest.Equal(endDateResponse) {
		rotationDto.EndDate = data.EndDate.ValueString()
	}
	//

	data = RotationDtoToModel(data.ScheduleId.ValueString(), rotationDto)

	tflog.Trace(ctx, "Read the ScheduleRotationResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ScheduleRotationResource into Terraform state")
}

func (r *ScheduleRotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.RotationModel
	var existingData dataModels.RotationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &existingData)...)

	tflog.Trace(ctx, "Updating the ScheduleRotationResource")

	plannedDto := RotationModelToDto(ctx, data)
	newDto := dto.Rotation{}
	existingRotationDto := RotationModelToDto(ctx, existingData)

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", data.ScheduleId.ValueString(), data.Id.ValueString())).
		Method(httpClient.PATCH).
		SetBody(plannedDto).
		SetBodyParseObject(&newDto).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to update rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to update rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update rotation, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation, got http response: %d", statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to update rotation, got error: %s", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation, got error: %s", err))
	}

	if !(resp.Diagnostics.HasError() || areUserListsEqual(plannedDto.Participants, newDto.Participants)) {
		plannedParticipants, _ := json.Marshal(plannedDto.Participants)
		newParticipants, _ := json.Marshal(newDto.Participants)
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to create rotation. The received participants list from the server does not match the one defined in Terraform configuration. Expected: %s, Got: %s", plannedParticipants, newParticipants))
		resp.Diagnostics.AddAttributeError(
			path.Root("participants"),
			"Unable to create rotation. The received participants list from the server does not match the one defined in Terraform configuration.",
			fmt.Sprintf(
				"Expected: %s, Got: %s.\n"+
					"This can be caused by the use of old Opsgenie UserIDs, instead of Atlassian Account IDs. The server automatically converts old IDs into new ones, which causes a state mismatch in Terraform.\n"+
					"Please consider checking the ID values you specified.", plannedParticipants, newParticipants,
			),
		)
		restoreRotationSlient(r, data.ScheduleId.ValueString(), existingRotationDto)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// HEIMDALL-12257 Workaround for time format from the server not matching the one in the config despite denoting the same time
	startDateRequest, _ := time.Parse(time.RFC3339, data.StartDate.ValueString())
	startDateResponse, _ := time.Parse(time.RFC3339, newDto.StartDate)

	endDateRequest, _ := time.Parse(time.RFC3339, data.EndDate.ValueString())
	endDateResponse, _ := time.Parse(time.RFC3339, newDto.EndDate)

	if startDateRequest.Equal(startDateResponse) {
		newDto.StartDate = data.StartDate.ValueString()
	}

	if endDateRequest.Equal(endDateResponse) {
		newDto.EndDate = data.EndDate.ValueString()
	}
	//

	data = RotationDtoToModel(data.ScheduleId.ValueString(), newDto)

	tflog.Trace(ctx, "Updated the ScheduleRotationResource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "Saved the ScheduleRotationResource into Terraform state")
}

func (r *ScheduleRotationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.RotationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "Deleting the ScheduleRotationResource")

	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", data.ScheduleId.ValueString(), data.Id.ValueString())).
		Method(httpClient.DELETE).
		Send()

	if httpResp == nil {
		tflog.Error(ctx, "Client Error. Unable to delete rotation, got nil response")
		resp.Diagnostics.AddError("Client Error", "Unable to delete rotation, got nil response")
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to delete rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation, status code: %d. Got response: %s", statusCode, *errorResponse))
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

func areUserListsEqual(givenUserList []dto.ResponderInfo, receivedUserList []dto.ResponderInfo) bool {
	if len(givenUserList) != len(receivedUserList) {
		return false
	}
	for _, givenUser := range givenUserList {
		contains := slices.ContainsFunc(receivedUserList, func(receivedUser dto.ResponderInfo) bool {
			return givenUser.Equal(&receivedUser)
		})
		if !contains {
			return false
		}
	}
	return true
}

func cleanupRotationSilent(r *ScheduleRotationResource, scheduleID string, rotationID string) {
	_, _ = httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", scheduleID, rotationID)).
		Method(httpClient.DELETE).
		Send()
}

func restoreRotationSlient(r *ScheduleRotationResource, scheduleID string, rotationDto dto.Rotation) {
	_, _ = httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("v1/schedules/%s/rotations/%s", scheduleID, rotationDto.Id)).
		Method(httpClient.PATCH).
		SetBody(rotationDto).
		Send()
}
