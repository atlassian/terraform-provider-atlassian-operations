package dataModels

import (
	"github.com/atlassian/jsm-ops-terraform-provider/internal/dto"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	TeamDataSourceModel struct {
		Description     types.String                             `tfsdk:"description"`
		DisplayName     types.String                             `tfsdk:"display_name"`
		OrganizationId  types.String                             `tfsdk:"organization_id"`
		TeamId          types.String                             `tfsdk:"team_id"`
		TeamType        types.String                             `tfsdk:"team_type"`
		UserPermissions *PublicApiUserPermissionsDataSourceModel `tfsdk:"user_permissions"`
		Member          []UserDataSourceModel                    `tfsdk:"member"`
	}
	TeamMemberDataSourceModel struct {
		Id       types.String `tfsdk:"id"`
		Username types.String `tfsdk:"username"`
		Role     types.String `tfsdk:"role"`
	}
	PublicApiUserPermissionsDataSourceModel struct {
		AddMembers    types.Bool `tfsdk:"add_members"`
		DeleteTeam    types.Bool `tfsdk:"delete_team"`
		RemoveMembers types.Bool `tfsdk:"remove_members"`
		UpdateTeam    types.Bool `tfsdk:"update_team"`
	}
)

func TeamDtoToModel(dto dto.TeamDto, usersDto []dto.UserDto) TeamDataSourceModel {
	model := TeamDataSourceModel{
		Description:     types.StringValue(dto.Description),
		DisplayName:     types.StringValue(dto.DisplayName),
		OrganizationId:  types.StringValue(dto.OrganizationId),
		TeamId:          types.StringValue(dto.TeamId),
		TeamType:        types.StringValue(string(dto.TeamType)),
		UserPermissions: PublicApiUserPermissionsDtoToModel(dto.UserPermissions),
		Member:          make([]UserDataSourceModel, len(usersDto)),
	}
	for i, member := range usersDto {
		model.Member[i] = UserDtoToModel(member)
	}
	return model
}

func PublicApiUserPermissionsDtoToModel(dto dto.PublicApiUserPermissions) *PublicApiUserPermissionsDataSourceModel {
	return &PublicApiUserPermissionsDataSourceModel{
		AddMembers:    types.BoolValue(dto.AddMembers),
		DeleteTeam:    types.BoolValue(dto.DeleteTeam),
		RemoveMembers: types.BoolValue(dto.RemoveMembers),
		UpdateTeam:    types.BoolValue(dto.UpdateTeam),
	}
}
