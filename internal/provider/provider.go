package provider

import (
	"context"
	"fmt"
	"github.com/atlassian/terraform-provider-jsm-ops/internal/httpClient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
)

// jsmopsProviderModel maps provider schema data to a Go type.
type jsmopsProviderModel struct {
	CloudId    types.String `tfsdk:"cloud_id"`
	DomainName types.String `tfsdk:"domain_name"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
}

type JsmOpsClient struct {
	OpsClient           *httpClient.HttpClient
	TeamClient          *httpClient.HttpClient
	TeamEnableOpsClient *httpClient.HttpClient
	UserClient          *httpClient.HttpClient
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
	resp.TypeName = "jsm-ops"
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
			"username": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a jsm-ops API client for data sources and resources.
func (p *jsmopsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring jsm-ops provider")

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
			"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the cloudId. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.DomainName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("domain_name"),
			"Unknown domain name",
			"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the domain_name. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown jsm-ops API Username",
			"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the jsm-ops API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown jsm-ops API Password",
			"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the jsm-ops API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	isStaging := os.Getenv("JSM_OPS_STAGING") == "1"

	cloudId := os.Getenv("JSM_OPS_CLOUD_ID")
	domainName := os.Getenv("JSM_OPS_DOMAIN_NAME")
	username := os.Getenv("JSM_OPS_API_USERNAME")
	password := os.Getenv("JSM_OPS_API_TOKEN")

	if cloudId == "" {
		if config.CloudId.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloud_id"),
				"Invalid cloud instance ID",
				"The provider cannot create the jsm-ops API client as there is a null / an empty configuration value for the cloudId.",
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
				"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the domain_name. ",
			)
		} else {
			domainName = config.DomainName.ValueString()
		}
	}

	if username == "" {
		if config.Username.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Unknown jsm-ops API Username",
				"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the jsm-ops API username. ",
			)
		} else {
			username = config.Username.ValueString()
		}
	}

	if password == "" {
		if config.Password.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Unknown jsm-ops API Password",
				"The provider cannot create the jsm-ops API client as there is an unknown configuration value for the jsm-ops API password. ",
			)
		} else {
			password = config.Password.ValueString()
		}
	}

	ctx = tflog.SetField(ctx, "jsm-ops_cloud_id", cloudId)
	ctx = tflog.SetField(ctx, "jsm-ops_domain_name", domainName)
	ctx = tflog.SetField(ctx, "jsm-ops_username", username)
	ctx = tflog.SetField(ctx, "jsm-ops_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "jsm-ops_password")

	tflog.Debug(ctx, "Creating jsm-ops client")

	// Create a new jsm-ops client using the configuration values
	client := &JsmOpsClient{
		OpsClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(username, password).
			SetBaseUrl(fmt.Sprintf("https://api.atlassian.com/jsm/ops/api/%s", cloudId)),
		TeamClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(username, password).
			SetBaseUrl(fmt.Sprintf("https://%s/gateway/api/public/teams/v1/org/", domainName)),
		// Undocumented API for enabling Operations for teams
		TeamEnableOpsClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(username, password).
			SetBaseUrl(fmt.Sprintf("https://%s/gateway/api/jsm/ops/web/%s/v1/teams/enable-ops", domainName, cloudId)),
		UserClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(username, password).
			SetBaseUrl(fmt.Sprintf("https://%s/rest/api/3/user/", domainName)),
	}

	if isStaging {
		client.OpsClient = client.OpsClient.SetBaseUrl(fmt.Sprintf("https://api.stg.atlassian.com/jsm/ops/api/%s", cloudId))
	}

	// Make the jsm-ops client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured jsm-ops client", map[string]any{"success": true})
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
	}
}
