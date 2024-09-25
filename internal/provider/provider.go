package provider

import (
	"context"
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
	OpsClient  any
	TeamClient any
	UserClient any
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
	resp.TypeName = "jsmops"
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

// Configure prepares a jsmops API client for data sources and resources.
func (p *jsmopsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring jsmops provider")

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
			"The provider cannot create the jsmops API client as there is an unknown configuration value for the cloudId. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.DomainName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("domain_name"),
			"Unknown domain name",
			"The provider cannot create the jsmops API client as there is an unknown configuration value for the domain_name. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown jsmops API Username",
			"The provider cannot create the jsmops API client as there is an unknown configuration value for the jsmops API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown jsmops API Password",
			"The provider cannot create the jsmops API client as there is an unknown configuration value for the jsmops API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "jsmOps_CloudId", config.CloudId)
	ctx = tflog.SetField(ctx, "jsmOps_domainName", config.DomainName)
	ctx = tflog.SetField(ctx, "jsmOps_username", config.Username)
	ctx = tflog.SetField(ctx, "jsmOps_password", config.Password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "jsmOps_password")

	tflog.Debug(ctx, "Creating JsmOps client")

	// Create a new jsmops client using the configuration values
	client := &JsmOpsClient{}

	// Make the jsmops client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured JsmOps client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *jsmopsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *jsmopsProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
