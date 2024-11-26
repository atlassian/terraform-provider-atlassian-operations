package provider

import (
	"context"
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
	"time"
)

// jsmopsProviderModel maps provider schema data to a Go type.
type jsmopsProviderModel struct {
	CloudId         types.String `tfsdk:"cloud_id"`
	DomainName      types.String `tfsdk:"domain_name"`
	EmailAddress    types.String `tfsdk:"email_address"`
	Token           types.String `tfsdk:"token"`
	ApiRetryCount   types.Int32  `tfsdk:"api_retry_count"`
	ApiRetryWait    types.Int32  `tfsdk:"api_retry_wait"`
	ApiRetryWaitMax types.Int32  `tfsdk:"api_retry_wait_max"`
}

type JsmOpsClient struct {
	OpsClient  *httpClient.HttpClient
	TeamClient *httpClient.HttpClient
	UserClient *httpClient.HttpClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &jsmopsProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &jsmopsProvider{
			version: version,
		}
	}
}

// jsmopsProvider is the provider implementation.
type jsmopsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *jsmopsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "atlassian-operations"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *jsmopsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud_id": schema.StringAttribute{
				Required: true,
			},
			"domain_name": schema.StringAttribute{
				Required: true,
			},
			"email_address": schema.StringAttribute{
				Required: true,
			},
			"token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"api_retry_count": schema.Int32Attribute{
				Optional: true,
			},
			"api_retry_wait": schema.Int32Attribute{
				Optional: true,
			},
			"api_retry_wait_max": schema.Int32Attribute{
				Optional: true,
			},
		},
	}
}

// Configure prepares a atlassian-operations API client for data sources and resources.
func (p *jsmopsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring atlassian-operations provider")

	// Retrieve provider data from configuration
	var config jsmopsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.CloudId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("cloud_id"),
			"Unknown cloud instance ID",
			"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the cloudId. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.DomainName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("domain_name"),
			"Unknown domain name",
			"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the domain_name. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.EmailAddress.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("email_address"),
			"Unknown atlassian-operations API EmailAddress",
			"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the atlassian-operations API email_address. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown atlassian-operations API Token",
			"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the atlassian-operations API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.ApiRetryCount.IsUnknown() {
		config.ApiRetryCount = types.Int32Value(3)
	}

	if config.ApiRetryWait.IsUnknown() {
		config.ApiRetryWait = types.Int32Value(5)
	}

	if config.ApiRetryWaitMax.IsUnknown() {
		config.ApiRetryWaitMax = types.Int32Value(20)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	isStaging := os.Getenv("ATLASSIAN_OPS_STAGING") == "1"

	cloudId := os.Getenv("ATLASSIAN_OPS_CLOUD_ID")
	domainName := os.Getenv("ATLASSIAN_OPS_DOMAIN_NAME")
	email_address := os.Getenv("ATLASSIAN_OPS_API_USERNAME")
	token := os.Getenv("ATLASSIAN_OPS_API_TOKEN")

	if cloudId == "" {
		if config.CloudId.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloud_id"),
				"Invalid cloud instance ID",
				"The provider cannot create the atlassian-operations API client as there is a null / an empty configuration value for the cloudId.",
			)
		} else {
			cloudId = config.CloudId.ValueString()
		}
	}

	if domainName == "" {
		if config.DomainName.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("domain_name"),
				"Unknown domain name",
				"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the domain_name. ",
			)
		} else {
			domainName = config.DomainName.ValueString()
		}
	}

	if email_address == "" {
		if config.EmailAddress.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("email_address"),
				"Unknown atlassian-operations API EmailAddress",
				"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the atlassian-operations API email_address. ",
			)
		} else {
			email_address = config.EmailAddress.ValueString()
		}
	}

	if token == "" {
		if config.Token.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("token"),
				"Unknown atlassian-operations API Token",
				"The provider cannot create the atlassian-operations API client as there is an unknown configuration value for the atlassian-operations API token. ",
			)
		} else {
			token = config.Token.ValueString()
		}
	}

	ctx = tflog.SetField(ctx, "atlassian-operations_cloud_id", cloudId)
	ctx = tflog.SetField(ctx, "atlassian-operations_domain_name", domainName)
	ctx = tflog.SetField(ctx, "atlassian-operations_email_address", email_address)
	ctx = tflog.SetField(ctx, "atlassian-operations_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "atlassian-operations_token")

	tflog.Debug(ctx, "Creating atlassian-operations client")

	// Create a new atlassian-operations client using the configuration values
	client := &JsmOpsClient{
		OpsClient: httpClient.
			NewHttpClient().
			SetRetryCount(int(config.ApiRetryCount.ValueInt32())).
			SetRetryWaitTime(time.Duration(config.ApiRetryWait.ValueInt32())*time.Second).
			SetRetryMaxWaitTime(time.Duration(config.ApiRetryWaitMax.ValueInt32())*time.Second).
			SetDefaultBasicAuth(email_address, token).
			SetBaseUrl(fmt.Sprintf("https://api.atlassian.com/jsm/ops/api/%s", cloudId)),
		TeamClient: httpClient.
			NewHttpClient().
			SetRetryCount(int(config.ApiRetryCount.ValueInt32())).
			SetRetryWaitTime(time.Duration(config.ApiRetryWait.ValueInt32())*time.Second).
			SetRetryMaxWaitTime(time.Duration(config.ApiRetryWaitMax.ValueInt32())*time.Second).
			SetDefaultBasicAuth(email_address, token).
			SetBaseUrl(fmt.Sprintf("https://%s/gateway/api/public/teams/v1/org/", domainName)),
		UserClient: httpClient.
			NewHttpClient().
			SetRetryCount(int(config.ApiRetryCount.ValueInt32())).
			SetRetryWaitTime(time.Duration(config.ApiRetryWait.ValueInt32())*time.Second).
			SetRetryMaxWaitTime(time.Duration(config.ApiRetryWaitMax.ValueInt32())*time.Second).
			SetDefaultBasicAuth(email_address, token).
			SetBaseUrl(fmt.Sprintf("https://%s/rest/api/3/user/", domainName)),
	}

	if isStaging {
		client.OpsClient = client.OpsClient.SetBaseUrl(fmt.Sprintf("https://api.stg.atlassian.com/jsm/ops/api/%s", cloudId))
	}

	// Make the atlassian-operations client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured atlassian-operations client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *jsmopsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewTeamDataSource,
		NewScheduleDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *jsmopsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewScheduleRotationResource,
		NewScheduleResource,
		NewTeamResource,
		NewEscalationResource,
		NewEmailIntegrationResource,
		NewApiIntegrationResource,
	}
}
