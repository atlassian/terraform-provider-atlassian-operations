package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

var ProviderAttributes = map[string]schema.Attribute{
	"cloud_id": schema.StringAttribute{
		Description: "The unique identifier of your Atlassian Cloud instance. This can be found in your Atlassian Cloud URL.",
		Required:    true,
	},
	"domain_name": schema.StringAttribute{
		Description: "The domain name of your Atlassian Cloud instance (e.g., 'your-domain.atlassian.net').",
		Required:    true,
	},
	"email_address": schema.StringAttribute{
		Description: "The email address associated with your Atlassian Cloud account. This must be an admin account.",
		Required:    true,
	},
	"token": schema.StringAttribute{
		Description: "Your Atlassian API token. You can generate this from your Atlassian account settings.",
		Required:    true,
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
