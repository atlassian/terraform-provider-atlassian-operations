// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	clientConfiguration dto.AtlassianOpsProviderModel
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "User data source.\n\n" +
			"With `product_type = \"jira-service-desk\"` and `auth_type = \"scoped\"` (a scoped API token or a service account), users are looked up through " +
			"`api.atlassian.com/ex/jira/{cloud_id}`; the token needs the `read:jira-user` scope. The account the token belongs to " +
			"also needs the Jira **Browse users and groups** global permission; a service account must be added to a group that has it, " +
			"otherwise the search returns no users (see [Find users](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-user-search/#api-rest-api-3-user-search-get) and " +
			"[Manage global permissions](https://support.atlassian.com/jira-cloud-administration/docs/manage-global-permissions/)).\n\n" +
			"With `product_type = \"compass\"`, the lookup instead always goes through the organization admin API and authenticates with `org_admin_token`, " +
			"regardless of `auth_type`. Service accounts cannot be looked up by email on Compass; find the service account's `account_id` " +
			"at admin.atlassian.com and use it directly wherever it is needed, instead of this data source.\n\n" +
			"Each instance of this data source makes one API call on every `plan`, `apply` and `refresh`; the organization admin API " +
			"rate-limits (`429`) a busy site. For large configurations, prefer raw account IDs in a team's `member` set over this " +
			"data source where you can.",
		Attributes: schemaAttributes.UserDataSourceAttributes,
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Trace(ctx, "Configuring user_data_source")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)

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
	productType := d.clientConfiguration.GetProductType()

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	tflog.Trace(ctx, "Reading user data source")

	switch productType {
	case "jira-service-desk":
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "Unable to read user data source. Configuration data provided is invalid.")
			return
		}

		jiraHint := func(statusCode int, errorResponse string) string {
			return userLookupAuthHint(d.clientConfiguration.IsScopedToken(), statusCode, errorResponse)
		}

		tflog.Trace(ctx, "Sending HTTP request to JSM User Search API")

		var data []dto.UserDto
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

		updateDiagnostics(ctx, err, clientResp, resp, jiraHint)
		// Only when the request itself succeeded: a 401/403 is already reported by updateDiagnostics.
		if len(data) == 0 && !resp.Diagnostics.HasError() {
			if authenticationFailed(clientResp) {
				// Jira answers an unauthenticated search as anonymous: 200 with an empty list and this header.
				tflog.Error(ctx, "HTTP request to User Search API was not authenticated ("+seraphLoginReasonHeader+": "+seraphAuthenticationFailed+")")
				resp.Diagnostics.AddError("Authentication Error",
					"Unable to read user, authentication failed: Jira rejected the credentials "+
						"("+seraphLoginReasonHeader+": "+seraphAuthenticationFailed+")."+jiraHint(http.StatusUnauthorized, ""))
			} else {
				tflog.Error(ctx, "HTTP request to User Search API Returned an Empty Response: "+emptySearchReasons)
				resp.Diagnostics.AddError("Client Error", "Unable to read user, got an empty response: "+emptySearchReasons)
			}
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

		updateDiagnostics(ctx, err, clientResp, resp, jiraHint)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Trace(ctx, "HTTP request to User API Succeeded. Parsing the fetched data to Terraform model")
		model = UserDtoToModel(data[0])

	default:
		if model.OrganizationId.IsNull() || model.OrganizationId.IsUnknown() {
			tflog.Error(
				ctx,
				fmt.Sprintf("Organization ID is required for %s. Please provide a valid organization ID.", productType),
			)
			resp.Diagnostics.AddAttributeError(
				path.Root("organization_id"),
				"Missing Required Attribute",
				fmt.Sprintf("Organization ID is required for %s. Please provide a valid organization ID.", productType),
			)
			return
		} else if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "Unable to read user data source. Configuration data provided is invalid.")
			return
		}

		tflog.Trace(ctx, "Sending HTTP request to Org Admin User Search API")

		var searchResponseDto dto.OrgUserSearchResponseDto

		clientResp, err := httpClientHelpers.
			GenerateUserClientRequest(d.clientConfiguration).
			Method("GET").
			JoinBaseUrl(fmt.Sprintf("%s/directories/-/users", model.OrganizationId.ValueString())).
			SetQueryParams(map[string]string{
				"limit":      "1",
				"searchTerm": model.EmailAddress.ValueString(),
			}).
			SetBodyParseObject(&searchResponseDto).
			Send()

		updateDiagnostics(ctx, err, clientResp, resp, orgAdminAuthHint)
		if len(searchResponseDto.Data) == 0 && !resp.Diagnostics.HasError() {
			msg := emptyOrgAdminSearchError(model.EmailAddress.ValueString())
			tflog.Error(ctx, "HTTP request to User Search API Returned an Empty Response: "+msg)
			resp.Diagnostics.AddError("Client Error", msg)
		}
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Trace(ctx, "HTTP request to User API Succeeded. Parsing the fetched data to Terraform model")
		model = OrgUserDtoToModel(searchResponseDto.Data[0], model)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Successfully read user data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// emptyOrgAdminSearchError returns the error to report when the org-admin directory search (product_type =
// "compass") found no user for emailAddress. That API cannot look up a service account by email at all, so a
// service account address gets its own message instead of the generic "no user found or bad credentials" one.
func emptyOrgAdminSearchError(emailAddress string) string {
	if dto.IsServiceAccountEmail(emailAddress) {
		return "Unable to read user: service accounts cannot be looked up by email on Compass; use the account_id directly"
	}
	return "Unable to read user, got an empty response: no account with that email address was found in the organization's directory. " +
		"Only accounts managed by the organization can be looked up through the organization admin API."
}

// emptySearchReasons are the causes of an empty result of the Jira user search.
const emptySearchReasons = "no user has that email address, the account lacks the Jira \"Browse users and groups\" global permission " +
	"(a service account must be in a group that has it), or the user hides their email address."

// updateDiagnostics reports a failed HTTP request. hint is called with the status code and error body to append
// credential guidance appropriate to the call that failed: the Jira user lookup and the org-admin directory search
// authenticate differently, so each passes its own (userLookupAuthHint, orgAdminAuthHint).
func updateDiagnostics(ctx context.Context, err error, clientResp *httpClient.Response, resp *datasource.ReadResponse, hint func(statusCode int, errorResponse string) string) {
	if err != nil {
		tflog.Error(ctx, "Sending HTTP request to User Search API Failed")
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read user, got error: %s", err))
	} else if clientResp.IsError() {
		statusCode := clientResp.GetStatusCode()
		errorResponse := clientResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user, status code: %d. Got response: %s", statusCode, *errorResponse))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, status code: %d. Got response: %s", statusCode, *errorResponse)+hint(statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to read user, got http response: %d", statusCode))
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got http response: %d", statusCode)+hint(statusCode, ""))
		}
	}
}

// userLookupAuthHint returns guidance to append when the user lookup was rejected because of the credentials.
// api.atlassian.com (used with auth_type = "scoped") answers an invalid or expired token with 401, whereas Jira
// (classic token) answers 200 with an empty list and the X-Seraph-LoginReason header handled in Read.
//
// The two paths get different guidance for a 401: the scoped path already has auth_type = "scoped", so
// scopedTokenGuideHint's "set auth_type = scoped" is not useful there; it points at the token itself and at the
// possibility that it is actually a classic token instead. The classic path's token may be a scoped one used
// without auth_type = "scoped", so that hint still points there.
func userLookupAuthHint(scoped bool, statusCode int, errorResponse string) string {
	if hint := scopedTokenHint(scoped, statusCode, errorResponse); hint != "" {
		return hint
	}
	if statusCode != http.StatusUnauthorized {
		return ""
	}
	if scoped {
		return " Check `email_address` and `token`; the token may be expired, and a classic API token needs `auth_type = \"classic\"`."
	}
	return " Check `email_address` and `token`. " + scopedTokenGuideHint
}
