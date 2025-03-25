package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

var ProviderAttributes = map[string]schema.Attribute{
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
}
