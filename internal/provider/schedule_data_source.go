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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ScheduleDataSource{}
	_ datasource.DataSourceWithConfigure = &ScheduleDataSource{}
)

func NewScheduleDataSource() datasource.DataSource {
	return &ScheduleDataSource{}
}

// ScheduleDataSource defines the data source implementation.
type ScheduleDataSource struct {
	client *httpClient.HttpClient
}

func (d *ScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (d *ScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Schedule data source",
		Attributes:          schemaAttributes.ScheduleDataSourceAttributes,
	}
}

func (d *ScheduleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring schedule_data_source")
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*JsmOpsClient)

	if !ok {
		tflog.Error(ctx, "Cannot configure schedule_data_source."+
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *httpClient.HttpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client.OpsClient
	tflog.Trace(ctx, "Configured schedule_data_source")
}

func (d *ScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model dataModels.ScheduleModel
	var data dto.ListResponse[dto.Schedule]

	tflog.Trace(ctx, "Reading schedule data source from JSM OPS API")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Unable to read schedule configuration. Configuration data provided is invalid.")
		return
	}

	tflog.Trace(ctx, "Sending HTTP request to JSM OPS API")

	errorMap := httpClient.NewOpsClientErrorMap()
	clientResp, err := d.client.NewRequest().
		Method("GET").
		JoinBaseUrl("/v1/schedules/").
		SetQueryParams(map[string]string{
			"query":  model.Name.ValueString(),
			"size":   "1",
			"expand": "rotation",
		}).
		SetBodyParseObject(&data).
		SetErrorParseMap(&errorMap).
		Send()

	if err != nil {
		tflog.Error(ctx, "Sending HTTP request to JSM OPS API Failed")
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read schedule, got error: %s", err))
	} else if clientResp.IsError() {
		statusCode := clientResp.GetStatusCode()
		errorResponse := errorMap[statusCode]
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read schedule, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read schedule, status code: %d. Got response: %s", statusCode, errorResponse.Error()))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read schedule, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read schedule, got http response: %d", statusCode))
		}
	} else if len(data.Values) == 0 {
		tflog.Error(ctx, "No schedules found")
		resp.Diagnostics.AddError("Client Error", "No schedules found")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "HTTP request to JSM OPS API Succeeded. Parsing the fetched data to Terraform model")
	model = ScheduleDtoToModel(data.Values[0])

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Successfully read schedule data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
