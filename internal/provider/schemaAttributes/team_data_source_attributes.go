package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var TeamDataSourceAttributes = map[string]schema.Attribute{
	"description": schema.StringAttribute{
		Description: "The description of the team",
		Computed:    true,
	},
	"display_name": schema.StringAttribute{
		Description: "The display name of the team",
		Computed:    true,
	},
	"organization_id": schema.StringAttribute{
		Description: "The organization ID of the team",
		Required:    true,
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team",
		Required:    true,
	},
	"team_type": schema.StringAttribute{
		Description: "The type of the team",
		Computed:    true,
	},
	"user_permissions": schema.SingleNestedAttribute{
		Description: "The user permissions of the team",
		Computed:    true,
		Attributes:  PublicApiUserPermissionsDataSourceAttributes,
	},
	"member": schema.ListNestedAttribute{
		Description: "The members of the team",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: UserDataSourceAttributes,
		},
	},
}

var PublicApiUserPermissionsDataSourceAttributes = map[string]schema.Attribute{
	"add_members": schema.BoolAttribute{
		Description: "The permission to add members to the team",
		Computed:    true,
	},
	"delete_team": schema.BoolAttribute{
		Description: "The permission to delete the team",
		Computed:    true,
	},
	"remove_members": schema.BoolAttribute{
		Description: "The permission to remove members from the team",
		Computed:    true,
	},
	"update_team": schema.BoolAttribute{
		Description: "The permission to update the team",
		Computed:    true,
	},
}
