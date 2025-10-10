package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var ProviderAttributes = map[string]schema.Attribute{
	"product_type": schema.StringAttribute{
		Optional:    true,
		Description: "The type of Atlassian Operations product you are using. This can be 'jira-service-desk' or 'compass'. Defaults to 'jira-service-desk'.",
		Validators: []validator.String{
			stringvalidator.OneOf("jira-service-desk", "compass"),
		},
	},
	"cloud_id": schema.StringAttribute{
		Description: "The unique identifier of your Atlassian Cloud instance. This can be found in your Atlassian Cloud URL.",
		Optional:    true,
	},
	"domain_name": schema.StringAttribute{
		Description: "The domain name of your Atlassian Cloud instance (e.g., 'your-domain.atlassian.net').",
		Optional:    true,
	},
	"email_address": schema.StringAttribute{
		Description: "The email address associated with your Atlassian Cloud account. This must be an admin account.",
		Optional:    true,
	},
	"token": schema.StringAttribute{
		Description: "Your Atlassian API token. You can generate this from your Atlassian account settings.",
		Optional:    true,
		Sensitive:   true,
	},
	"org_admin_token": schema.StringAttribute{
		Description: "The API token of the organization admin, to be able to use User APIs. This field is only required & used for Compass.",
		Optional:    true,
		Sensitive:   true,
	},
	"api_retry_count": schema.Int32Attribute{
		Description: "The number of times to retry failed API requests. Defaults to 3.",
		Optional:    true,
	},
	"api_retry_wait": schema.Int32Attribute{
		Description: "The initial wait time in seconds between API retries. This value is doubled for each subsequent retry. Defaults to 1.",
		Optional:    true,
	},
	"api_retry_wait_max": schema.Int32Attribute{
		Description: "The maximum wait time in seconds between API retries. Defaults to 30.",
		Optional:    true,
	},
}
