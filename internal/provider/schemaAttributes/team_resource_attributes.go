package schemaAttributes

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var TeamResourceAttributes = map[string]schema.Attribute{
	"description": schema.StringAttribute{
		Description: "The description of the team",
		Required:    true,
	},
	"display_name": schema.StringAttribute{
		Description: "The display name of the team",
		Required:    true,
	},
	"organization_id": schema.StringAttribute{
		Description: "The organization ID of the team",
		Required:    true,
	},
	"id": schema.StringAttribute{
		Description: "The ID of the team",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"site_id": schema.StringAttribute{
		Description: "The site ID of the team",
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 255),
		},
	},
	"team_type": schema.StringAttribute{
		Description: "The type of the team",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf(string(dto.OPEN), string(dto.MEMBER_INVITE), string(dto.EXTERNAL)),
		},
	},
	"user_permissions": schema.SingleNestedAttribute{
		Description: "The user permissions of the team",
		Computed:    true,
		Optional:    false,
		Required:    false,
		Attributes:  PublicApiUserPermissionsResourceAttributes,
	},
	"member": schema.SetNestedAttribute{
		Description: "The members of the team",
		Computed:    true,
		Optional:    true,
		Default: setdefault.StaticValue(
			types.SetValueMust(
				types.ObjectType{AttrTypes: dataModels.TeamMemberModelMap},
				[]attr.Value{},
			),
		),
		NestedObject: schema.NestedAttributeObject{
			Attributes: TeamMemberResourceAttributes,
		},
	},
}

var PublicApiUserPermissionsResourceAttributes = map[string]schema.Attribute{
	"add_members": schema.BoolAttribute{
		Description: "The permission to add members to the team",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
	"delete_team": schema.BoolAttribute{
		Description: "The permission to delete the team",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
	"remove_members": schema.BoolAttribute{
		Description: "The permission to remove members from the team",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
	"update_team": schema.BoolAttribute{
		Description: "The permission to update the team",
		Computed:    true,
		Optional:    false,
		Required:    false,
	},
}

var TeamMemberResourceAttributes = map[string]schema.Attribute{
	"account_id": schema.StringAttribute{
		Description: "The account ID of the user",
		Required:    true,
	},
}
