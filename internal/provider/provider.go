package provider

import (
	"context"
	"github.com/atlassian/terraform-provider-jsm-ops/internal/httpClient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// jsmopsProviderModel maps provider schema data to a Go type.
type jsmopsProviderModel struct {
	CloudId    types.String `tfsdk:"cloud_id"`
	DomainName types.String `tfsdk:"domain_name"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
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

	ctx = tflog.SetField(ctx, "jsm-ops_cloud_id", config.CloudId)
	ctx = tflog.SetField(ctx, "jsm-ops_domain_name", config.DomainName)
	ctx = tflog.SetField(ctx, "jsm-ops_username", config.Username)
	ctx = tflog.SetField(ctx, "jsm-ops_password", config.Password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "jsm-ops_password")

	tflog.Debug(ctx, "Creating jsm-ops client")

	// Create a new jsm-ops client using the configuration values
	client := &JsmOpsClient{
		OpsClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(config.Username.ValueString(), config.Password.ValueString()).
			SetBaseUrl("https://api.stg.atlassian.com/jsm/ops/api/" + config.CloudId.ValueString()),
		TeamClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(config.Username.ValueString(), config.Password.ValueString()).
			SetBaseUrl("https://" + config.DomainName.ValueString() + "/gateway/api/public/teams/v1/org/"),
		UserClient: httpClient.
			NewHttpClient().
			SetDefaultBasicAuth(config.Username.ValueString(), config.Password.ValueString()).
			SetBaseUrl("https://" + config.DomainName.ValueString() + "/rest/api/3/user/"),
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
