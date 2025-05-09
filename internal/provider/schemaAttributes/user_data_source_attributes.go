package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var UserDataSourceAttributes = map[string]schema.Attribute{
	"account_id": schema.StringAttribute{
		Description: "The unique Atlassian account identifier for the user. This is a permanent, unchangeable ID.",
		Computed:    true,
	},
	"account_type": schema.StringAttribute{
		Description: "The type of Atlassian account (e.g., 'atlassian', 'customer', 'app'). Determines the user's access level and capabilities.",
		Computed:    true,
	},
	"active": schema.BoolAttribute{
		Description: "Indicates whether the user account is currently active and can access Atlassian services.",
		Computed:    true,
	},
	"application_roles": schema.ListNestedAttribute{
		Description: "List of roles and permissions the user has across different Atlassian applications.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: ApplicationRoleDataSourceAttributes,
		},
	},
	"avatar_urls": schema.SingleNestedAttribute{
		Description: "Collection of URLs for the user's avatar image in different sizes.",
		Computed:    true,
		Attributes:  AvatarUrlsBeanDataSourceAttributes,
	},
	"display_name": schema.StringAttribute{
		Description: "The user's full name as it appears in the Atlassian interface.",
		Computed:    true,
	},
	"email_address": schema.StringAttribute{
		Description: "The user's email address. This is used as the primary identifier for looking up user information.",
		Required:    true,
	},
	"expand": schema.StringAttribute{
		Description: "Comma-separated list of additional user details to include in the response.",
		Computed:    true,
	},
	"groups": schema.ListNestedAttribute{
		Description: "List of groups the user belongs to, determining their access rights and permissions.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: GroupNameDataSourceAttributes,
		},
	},
	"locale": schema.StringAttribute{
		Description: "The user's preferred language and region settings (e.g., 'en_US', 'fr_FR').",
		Computed:    true,
	},
	"timezone": schema.StringAttribute{
		Description: "The user's configured timezone in IANA format (e.g., 'America/New_York'). Used for displaying times and dates.",
		Computed:    true,
	},
}

var ApplicationRoleDataSourceAttributes = map[string]schema.Attribute{
	"default_groups": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "List of group names that are automatically assigned to users with this application role.",
	},
	"default_groups_details": schema.ListNestedAttribute{
		Description: "Detailed information about the default groups associated with this application role.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: GroupNameDataSourceAttributes,
		},
	},
	"defined": schema.BoolAttribute{
		Description: "Indicates whether this application role has been explicitly defined or is inherited.",
		Computed:    true,
	},
	"group_details": schema.ListNestedAttribute{
		Description: "Detailed information about all groups associated with this application role.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: GroupNameDataSourceAttributes,
		},
	},
	"groups": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "List of all group names associated with this application role.",
	},
	"has_unlimited_seats": schema.BoolAttribute{
		Description: "Indicates whether this application role has no limit on the number of users who can be assigned to it.",
		Computed:    true,
	},
	"key": schema.StringAttribute{
		Description: "The unique identifier for this application role.",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The human-readable name of this application role.",
		Computed:    true,
	},
	"number_of_seats": schema.Int32Attribute{
		Description: "The maximum number of users who can be assigned this application role.",
		Computed:    true,
	},
	"platform": schema.BoolAttribute{
		Description: "Indicates whether this is a platform-level application role that applies across all Atlassian products.",
		Computed:    true,
	},
}

var GroupNameDataSourceAttributes = map[string]schema.Attribute{
	"group_id": schema.StringAttribute{
		Description: "The unique identifier for the group.",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The display name of the group.",
		Computed:    true,
	},
	"self": schema.StringAttribute{
		Description: "The URL to the REST API endpoint for this group.",
		Computed:    true,
	},
}

var AvatarUrlsBeanDataSourceAttributes = map[string]schema.Attribute{
	"a_16x16": schema.StringAttribute{
		Description: "URL to the user's 16x16 pixel avatar image.",
		Computed:    true,
	},
	"a_24x24": schema.StringAttribute{
		Description: "URL to the user's 24x24 pixel avatar image.",
		Computed:    true,
	},
	"a_32x32": schema.StringAttribute{
		Description: "URL to the user's 32x32 pixel avatar image.",
		Computed:    true,
	},
	"a_48x48": schema.StringAttribute{
		Description: "URL to the user's 48x48 pixel avatar image.",
		Computed:    true,
	},
}
