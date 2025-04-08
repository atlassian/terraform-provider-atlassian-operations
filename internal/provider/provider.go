package provider

import (
	"context"
	"os"
	"time"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

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
		Attributes: schemaAttributes.ProviderAttributes,
	}
}

// Configure prepares a atlassian-operations API clientConfiguration for data sources and resources.
func (p *jsmopsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring atlassian-operations provider")

	// Retrieve provider data from configuration
	var config dataModels.JsmopsProviderTfModel
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
			"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the cloudId. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.DomainName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("domain_name"),
			"Unknown domain name",
			"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the domain_name. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.EmailAddress.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("email_address"),
			"Unknown atlassian-operations API EmailAddress",
			"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the atlassian-operations API email_address. "+
				"Either target apply the source of the value first, set the value statically in the configuration.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown atlassian-operations API Token",
			"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the atlassian-operations API token. "+
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
	emailAddress := os.Getenv("ATLASSIAN_OPS_API_USERNAME")
	token := os.Getenv("ATLASSIAN_OPS_API_TOKEN")

	if cloudId == "" {
		if config.CloudId.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloud_id"),
				"Invalid cloud instance ID",
				"The provider cannot create the atlassian-operations API clientConfiguration as there is a null / an empty configuration value for the cloudId.",
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
				"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the domain_name. ",
			)
		} else {
			domainName = config.DomainName.ValueString()
		}
	}

	if emailAddress == "" {
		if config.EmailAddress.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("email_address"),
				"Unknown atlassian-operations API EmailAddress",
				"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the atlassian-operations API email_address. ",
			)
		} else {
			emailAddress = config.EmailAddress.ValueString()
		}
	}

	if token == "" {
		if config.Token.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("token"),
				"Unknown atlassian-operations API Token",
				"The provider cannot create the atlassian-operations API clientConfiguration as there is an unknown configuration value for the atlassian-operations API token. ",
			)
		} else {
			token = config.Token.ValueString()
		}
	}

	ctx = tflog.SetField(ctx, "atlassian-operations_cloud_id", cloudId)
	ctx = tflog.SetField(ctx, "atlassian-operations_domain_name", domainName)
	ctx = tflog.SetField(ctx, "atlassian-operations_email_address", emailAddress)
	ctx = tflog.SetField(ctx, "atlassian-operations_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "atlassian-operations_token")

	tflog.Debug(ctx, "Creating atlassian-operations clientConfiguration")

	// Create a new atlassian-operations clientConfiguration using the configuration values
	client := dto.NewJsmopsProviderModel(
		cloudId,
		domainName,
		emailAddress,
		token,
		int(config.ApiRetryCount.ValueInt32()),
		time.Duration(config.ApiRetryWait.ValueInt32())*time.Second,
		time.Duration(config.ApiRetryWaitMax.ValueInt32())*time.Second,
		isStaging,
	)

	// Make the atlassian-operations clientConfiguration available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured atlassian-operations clientConfiguration", map[string]any{"success": true})
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
		NewRoutingRuleResource,
		NewNotificationRuleResource,
		NewUserContactResource,
		NewAlertPolicyResource,
	}
}
