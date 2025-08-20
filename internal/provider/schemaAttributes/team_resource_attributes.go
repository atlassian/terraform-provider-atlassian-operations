package schemaAttributes

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var TeamResourceAttributes = map[string]schema.Attribute{
	"description": schema.StringAttribute{
		Description: "A detailed description of the team's purpose, responsibilities, and scope of operations.",
		Required:    true,
	},
	"display_name": schema.StringAttribute{
		Description: "The human-readable name of the team as it appears in the Atlassian interface. This should be clear and identifiable.",
		Required:    true,
	},
	"organization_id": schema.StringAttribute{
		Description: "The unique identifier of the organization this team belongs to. This determines the team's organizational context.",
		Required:    true,
	},
	"id": schema.StringAttribute{
		Description: "The unique identifier of the team. This is automatically generated when the team is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"site_id": schema.StringAttribute{
		Description: "The identifier of the Atlassian site where this team is configured. Must be between 1 and 255 characters.",
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 255),
		},
	},
	"team_type": schema.StringAttribute{
		Description: "The type of team that determines access and invitation policies. Valid values are 'OPEN' (anyone can join), 'MEMBER_INVITE' (members can invite others), or 'EXTERNAL' (managed externally).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf(string(dto.OPEN), string(dto.MEMBER_INVITE), string(dto.EXTERNAL)),
		},
	},
	"user_permissions": schema.SingleNestedAttribute{
		Description: "The set of permissions that define what operations users can perform on this team. These are computed based on team type and user roles.",
		Computed:    true,
		Optional:    false,
		Required:    false,
		Attributes:  PublicApiUserPermissionsResourceAttributes,
	},
	"member": schema.SetNestedAttribute{
		Description: "The set of users who are members of this team. Must contain at least one member. Each member is identified by their Atlassian account ID.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: TeamMemberResourceAttributes,
		},
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
	},
	"delete_default_resources": schema.BoolAttribute{
		Description: "Set to true to remove default escalation and schedule for newly created team. Be careful its also changes that team routing rule to None. That means you have to define routing rule as well. Defaults to false.",
		Optional:    true,
		Required:    false,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
	},
}

var PublicApiUserPermissionsResourceAttributes = map[string]schema.Attribute{
	"add_members": schema.BoolAttribute{
		Description: "Whether the user has permission to add new members to the team. This is computed based on the user's role and team type.",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
	"delete_team": schema.BoolAttribute{
		Description: "Whether the user has permission to delete the entire team. This is typically restricted to team administrators.",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
	"remove_members": schema.BoolAttribute{
		Description: "Whether the user has permission to remove existing members from the team. This is computed based on the user's role and team type.",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
	"update_team": schema.BoolAttribute{
		Description: "Whether the user has permission to modify team settings and properties. This includes changing the team name, description, and other configurations.",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
}

var TeamMemberResourceAttributes = map[string]schema.Attribute{
	"account_id": schema.StringAttribute{
		Description: "The unique Atlassian account identifier for the team member. This is used to uniquely identify users across Atlassian products.",
		Required:    true,
	},
}
