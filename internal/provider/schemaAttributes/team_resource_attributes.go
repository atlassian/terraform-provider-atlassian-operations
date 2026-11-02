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
		Description: "The identifier of the Atlassian site where this team is configured. Must be between 1 and 255 characters. When importing a team that is scoped to a site, include the site_id as the third part of the import identifier: `<team_id>,<organization_id>,<site_id>`; without site scoping the identifier is `<team_id>,<organization_id>`.",
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 255),
		},
	},
	"team_type": schema.StringAttribute{
		Description: "The type of team that determines access and invitation policies. Valid values are 'OPEN' (anyone can join), 'MEMBER_INVITE' (members can invite others), or 'EXTERNAL' (managed externally).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf(string(dto.OPEN), string(dto.MEMBER_INVITE)),
		},
	},
	"user_permissions": schema.SingleNestedAttribute{
		Description: "The set of permissions that define what operations users can perform on this team. These are computed based on team type and user roles. With `auth_type = \"scoped\"`, `update_team`, `add_members` and `remove_members` are reported together, and `delete_team` separately, based on what the token's account can do on this team.",
		Computed:    true,
		Optional:    false,
		Required:    false,
		Attributes:  PublicApiUserPermissionsResourceAttributes,
	},
	"member": schema.SetNestedAttribute{
		Description: "The set of users who are members of this team. Must contain at least one member. Each member is identified by their Atlassian account ID. The provider nominates the member whose account ID sorts first as text as the team's Operations admin; include at least one human member, who can administer the team in the Operations UI. " +
			"If the token belongs to a service account, that service account's own account ID must also be listed here, together with at least one human member, or it is removed from the team once creation finishes; without membership it can no longer update or delete the team, so every later `update` or `destroy` of this resource then fails. " +
			"Find a service account's account ID at admin.atlassian.com, under Directory > Service accounts.",
		Required: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: TeamMemberResourceAttributes,
		},
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
	},
	"delete_default_resources": schema.BoolAttribute{
		Description: "Set to true to remove the default escalation and schedule of a newly created team. Be careful: it also sets the team's routing rule to None, so you must define a routing rule yourself. Defaults to false.",
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
		Description: "The unique Atlassian account identifier for the team member, a raw account ID such as `70121:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`. This is used to uniquely identify users across Atlassian products.",
		Required:    true,
	},
}
