package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var UserDataSourceAttributes = map[string]schema.Attribute{
	"account_id": schema.StringAttribute{
		Description: "The account ID of the user",
		Computed:    true,
	},
	"account_type": schema.StringAttribute{
		Description: "The account type of the user",
		Computed:    true,
	},
	"active": schema.BoolAttribute{
		Description: "The active status of the user",
		Computed:    true,
	},
	"application_roles": schema.ListNestedAttribute{
		Description: "The application roles of the user",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: ApplicationRoleDataSourceAttributes,
		},
	},
	"avatar_urls": schema.SingleNestedAttribute{
		Description: "The avatar URLs of the user",
		Computed:    true,
		Attributes:  AvatarUrlsBeanDataSourceAttributes,
	},
	"display_name": schema.StringAttribute{
		Description: "The display name of the user",
		Computed:    true,
	},
	"email_address": schema.StringAttribute{
		Description: "The email address of the user",
		Required:    true,
	},
	"expand": schema.StringAttribute{
		Description: "The expand of the user",
		Computed:    true,
	},
	"groups": schema.ListNestedAttribute{
		Description: "The groups of the user",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: GroupNameDataSourceAttributes,
		},
	},
	"locale": schema.StringAttribute{
		Description: "The locale of the user",
		Computed:    true,
	},
	"timezone": schema.StringAttribute{
		Description: "The time zone of the user",
		Computed:    true,
	},
}

var ApplicationRoleDataSourceAttributes = map[string]schema.Attribute{
	"default_groups": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "The default groups of the application role",
	},
	"default_groups_details": schema.ListNestedAttribute{
		Description: "The default groups details of the application role",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: GroupNameDataSourceAttributes,
		},
	},
	"defined": schema.BoolAttribute{
		Description: "The defined status of the application role",
		Computed:    true,
	},
	"group_details": schema.ListNestedAttribute{
		Description: "The group details of the application role",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: GroupNameDataSourceAttributes,
		},
	},
	"groups": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "The groups of the application role",
	},
	"has_unlimited_seats": schema.BoolAttribute{
		Description: "The has unlimited seats status of the application role",
		Computed:    true,
	},
	"key": schema.StringAttribute{
		Description: "The key of the application role",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the application role",
		Computed:    true,
	},
	"number_of_seats": schema.Int32Attribute{
		Description: "The number of seats of the application role",
		Computed:    true,
	},
	"platform": schema.BoolAttribute{
		Description: "The platform status of the application role",
		Computed:    true,
	},
}

var GroupNameDataSourceAttributes = map[string]schema.Attribute{
	"group_id": schema.StringAttribute{
		Description: "The group ID",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the group",
		Computed:    true,
	},
	"self": schema.StringAttribute{
		Description: "The self of the group",
		Computed:    true,
	},
}

var AvatarUrlsBeanDataSourceAttributes = map[string]schema.Attribute{
	"a_16x16": schema.StringAttribute{
		Description: "The 16x16 avatar URL",
		Computed:    true,
	},
	"a_24x24": schema.StringAttribute{
		Description: "The 24x24 avatar URL",
		Computed:    true,
	},
	"a_32x32": schema.StringAttribute{
		Description: "The 32x32 avatar URL",
		Computed:    true,
	},
	"a_48x48": schema.StringAttribute{
		Description: "The 48x48 avatar URL",
		Computed:    true,
	},
}
