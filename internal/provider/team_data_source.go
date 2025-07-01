// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &teamDataSource{}
	_ datasource.DataSourceWithConfigure = &teamDataSource{}
)

func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

// teamDataSource defines the data source implementation.
type teamDataSource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Team data source",
		Attributes:          schemaAttributes.TeamDataSourceAttributes,
	}
}

func (d *teamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring team_data_source")
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)

	if !ok {
		tflog.Error(ctx, "Cannot configure team_data_source."+
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *httpClient.HttpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.clientConfiguration = client
	tflog.Trace(ctx, "Configured team_data_source")
}

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model dataModels.TeamModel
	var data dto.TeamDto
	var memberData dto.TeamMemberListResponse

	tflog.Trace(ctx, "Reading team data source")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Unable to read team data source configuration. Configuration data provided is invalid.")
		return
	}

	tflog.Trace(ctx, "Preparing HTTP Request to fetch team data from JSM Teams API")

	teamFetchUrl := fmt.Sprintf("/%s/teams/%s",
		model.OrganizationId.ValueString(),
		model.Id.ValueString())

	tflog.Trace(ctx, "Preparing HTTP Request to fetch team member data from JSM Team Members API")

	teamMembersFetchUrl := fmt.Sprintf("/%s/teams/%s/members",
		model.OrganizationId.ValueString(),
		model.Id.ValueString())

	tflog.Trace(ctx, "Sending HTTP request to JSM Teams API")

	clientResp, err := httpClientHelpers.
		GenerateTeamsClientRequest(d.clientConfiguration).
		Method("GET").
		JoinBaseUrl(teamFetchUrl).
		SetQueryParam("siteId", model.SiteId.ValueString()).
		SetBodyParseObject(&data).
		Send()

	if err != nil {
		tflog.Error(ctx, "Sending HTTP request to JSM Teams API Failed")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read team, got error: %s", err))
	} else if clientResp.IsError() {
		statusCode := clientResp.GetStatusCode()
		errorResponse := clientResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read team, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, got http response: %d", statusCode))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Sending HTTP request to JSM Team Members API")

	clientResp, err = httpClientHelpers.
		GenerateTeamsClientRequest(d.clientConfiguration).
		Method("POST").
		JoinBaseUrl(teamMembersFetchUrl).
		SetBodyParseObject(&memberData).
		Send()

	if err != nil {
		tflog.Error(ctx, "Sending HTTP request to JSM Team Members API Failed")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read team members, got error: %s", err))
		return
	} else if clientResp.IsError() {
		tflog.Error(ctx, "HTTP request to JSM Team Members API Returned an Error Code")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read team members, got status code: %d", clientResp.GetStatusCode()))
		return
	}

	tflog.Trace(ctx, "Converting Team Data into Terraform Model")
	// Convert the fetched data into the model
	model = TeamDtoToModel(data, memberData.Results, basetypes.NewBoolValue(false))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Successfully read team data")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
