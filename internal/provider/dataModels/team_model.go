package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	TeamModel struct {
		Description     types.String `tfsdk:"description"`
		DisplayName     types.String `tfsdk:"display_name"`
		OrganizationId  types.String `tfsdk:"organization_id"`
		TeamId          types.String `tfsdk:"team_id"`
		TeamType        types.String `tfsdk:"team_type"`
		UserPermissions types.Object `tfsdk:"user_permissions"`
		Member          types.List   `tfsdk:"member"`
	}
	TeamMemberModel struct {
		Id       types.String `tfsdk:"id"`
		Username types.String `tfsdk:"username"`
		Role     types.String `tfsdk:"role"`
	}
	PublicApiUserPermissionsModel struct {
		AddMembers    types.Bool `tfsdk:"add_members"`
		DeleteTeam    types.Bool `tfsdk:"delete_team"`
		RemoveMembers types.Bool `tfsdk:"remove_members"`
		UpdateTeam    types.Bool `tfsdk:"update_team"`
	}
)

var TeamModelMap = map[string]attr.Type{
	"description":     types.StringType,
	"display_name":    types.StringType,
	"organization_id": types.StringType,
	"team_id":         types.StringType,
	"team_type":       types.StringType,
	"user_permissions": types.ObjectType{
		AttrTypes: PublicApiUserPermissionsModelMap,
	},
	"member": types.ListType{ElemType: types.ObjectType{
		AttrTypes: TeamMemberModelMap,
	}},
}

var TeamMemberModelMap = map[string]attr.Type{
	"id":       types.StringType,
	"username": types.StringType,
	"role":     types.StringType,
}

var PublicApiUserPermissionsModelMap = map[string]attr.Type{
	"add_members":    types.BoolType,
	"delete_team":    types.BoolType,
	"remove_members": types.BoolType,
	"update_team":    types.BoolType,
}

func (receiver *TeamModel) AsValue() types.Object {
	return types.ObjectValueMust(TeamModelMap, map[string]attr.Value{
		"description":      receiver.Description,
		"display_name":     receiver.DisplayName,
		"organization_id":  receiver.OrganizationId,
		"team_id":          receiver.TeamId,
		"team_type":        receiver.TeamType,
		"user_permissions": receiver.UserPermissions,
		"member":           receiver.Member,
	})
}

func (receiver *TeamMemberModel) AsValue() types.Object {
	return types.ObjectValueMust(TeamMemberModelMap, map[string]attr.Value{
		"id":       receiver.Id,
		"username": receiver.Username,
		"role":     receiver.Role,
	})
}

func (receiver *PublicApiUserPermissionsModel) AsValue() types.Object {
	return types.ObjectValueMust(PublicApiUserPermissionsModelMap, map[string]attr.Value{
		"add_members":    receiver.AddMembers,
		"delete_team":    receiver.DeleteTeam,
		"remove_members": receiver.RemoveMembers,
		"update_team":    receiver.UpdateTeam,
	})
}
