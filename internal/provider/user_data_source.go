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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource defines the data source implementation.
type userDataSource struct {
	clientConfiguration dto.JsmopsProviderModel
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "User data source",
		Attributes:          schemaAttributes.UserDataSourceAttributes,
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring user_data_source")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.JsmopsProviderModel)

	if !ok {
		tflog.Error(ctx, "Cannot configure user_data_source."+
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.clientConfiguration = client
	tflog.Trace(ctx, "Configured user_data_source")
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model dataModels.UserModel
	var data []dto.UserDto

	tflog.Trace(ctx, "Reading user data source")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Unable to read user data source. Configuration data provided is invalid.")
		return
	}

	tflog.Trace(ctx, "Sending HTTP request to JSM User API")

	clientResp, err := httpClientHelpers.
		GenerateUserClientRequest(d.clientConfiguration).
		Method("GET").
		JoinBaseUrl("/search").
		SetQueryParams(map[string]string{
			"query":      model.EmailAddress.ValueString(),
			"maxResults": "1",
		}).
		SetBodyParseObject(&data).
		Send()

	if err != nil {
		tflog.Error(ctx, "Sending HTTP request to JSM User API Failed")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read user, got error: %s", err))
	} else if clientResp.IsError() {
		statusCode := clientResp.GetStatusCode()
		errorResponse := clientResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, status code: %d. Got response: %s", statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got http response: %d", statusCode))
		}
	} else if len(data) == 0 {
		tflog.Error(ctx, "HTTP request to JSM User API Returned an Empty Response."+
			"Either no user is found, or the credentials are invalid")
		resp.Diagnostics.AddError("Client Error",
			"Unable to read user, got an empty response. "+
				"This could be due to invalid credentials or no user being found for the given email address")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	clientResp, err = httpClientHelpers.
		GenerateUserClientRequest(d.clientConfiguration).
		Method("GET").
		SetQueryParams(map[string]string{
			"accountId": data[0].AccountId,
			"expand":    "groups,applicationRoles",
		}).
		SetBodyParseObject(&data[0]).
		Send()

	if err != nil {
		tflog.Error(ctx, "Sending HTTP request to JSM User API Failed")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	} else if clientResp.IsError() {
		tflog.Error(ctx, "HTTP request to JSM User API Returned an Error Status Code")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read user, got status code: %d", clientResp.GetStatusCode()))
		return
	}

	tflog.Trace(ctx, "HTTP request to JSM User API Succeeded. Parsing the fetched data to Terraform model")
	model = UserDtoToModel(data[0])

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Successfully read user data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
