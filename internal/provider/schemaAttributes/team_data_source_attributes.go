package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var TeamDataSourceAttributes = map[string]schema.Attribute{
	"description": schema.StringAttribute{
		Description: "A detailed description of the team's purpose, responsibilities, and scope of operations.",
		Computed:    true,
	},
	"display_name": schema.StringAttribute{
		Description: "The human-readable name of the team as it appears in the Atlassian interface.",
		Computed:    true,
	},
	"organization_id": schema.StringAttribute{
		Description: "The unique identifier of the organization this team belongs to. Required for team lookup.",
		Required:    true,
	},
	"id": schema.StringAttribute{
		Description: "The unique identifier of the team. Used to look up specific team information.",
		Required:    true,
	},
	"site_id": schema.StringAttribute{
		Description: "The identifier of the Atlassian site where this team is configured. Must be between 1 and 255 characters.",
		Computed:    true,
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 255),
		},
	},
	"team_type": schema.StringAttribute{
		Description: "The type of team (e.g., 'OPEN', 'MEMBER_INVITE', 'EXTERNAL'). Determines team access and invitation policies.",
		Computed:    true,
	},
	"user_permissions": schema.SingleNestedAttribute{
		Description: "The set of permissions that define what operations users can perform on this team.",
		Computed:    true,
		Attributes:  PublicApiUserPermissionsDataSourceAttributes,
	},
	"member": schema.SetNestedAttribute{
		Description: "The set of users who are members of this team. Each member has their own role and permissions.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: TeamMemberDataSourceAttributes,
		},
	},
	"delete_default_resources": schema.BoolAttribute{
		Description: "Set to true to remove default escalation and schedule for newly created team. Be careful its also changes that team routing rule to None. That means you have to define routing rule as well. Defaults to false.",
		Optional:    false,
		Required:    false,
		Computed:    true,
	},
}

var PublicApiUserPermissionsDataSourceAttributes = map[string]schema.Attribute{
	"add_members": schema.BoolAttribute{
		Description: "Indicates whether the user has permission to add new members to the team.",
		Computed:    true,
	},
	"delete_team": schema.BoolAttribute{
		Description: "Indicates whether the user has permission to delete the entire team.",
		Computed:    true,
	},
	"remove_members": schema.BoolAttribute{
		Description: "Indicates whether the user has permission to remove existing members from the team.",
		Computed:    true,
	},
	"update_team": schema.BoolAttribute{
		Description: "Indicates whether the user has permission to modify team settings and properties.",
		Computed:    true,
	},
}

var TeamMemberDataSourceAttributes = map[string]schema.Attribute{
	"account_id": schema.StringAttribute{
		Description: "The unique Atlassian account identifier for the team member.",
		Computed:    true,
	},
}
