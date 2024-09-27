package dataModels

import (
	"github.com/atlassian/terraform-provider-jsm-ops/internal/dto"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	UserDataSourceModel struct {
		AccountId        types.String                     `tfsdk:"account_id"`
		AccountType      types.String                     `tfsdk:"account_type"`
		Active           types.Bool                       `tfsdk:"active"`
		ApplicationRoles []ApplicationRoleDataSourceModel `tfsdk:"application_roles"`
		AvatarUrls       *AvatarUrlsBeanDataSourceModel   `tfsdk:"avatar_urls"`
		DisplayName      types.String                     `tfsdk:"display_name"`
		EmailAddress     types.String                     `tfsdk:"email_address"`
		Expand           types.String                     `tfsdk:"expand"`
		Groups           []GroupNameDataSourceModel       `tfsdk:"groups"`
		Locale           types.String                     `tfsdk:"locale"`
		TimeZone         types.String                     `tfsdk:"timezone"`
	}
	ApplicationRoleDataSourceModel struct {
		DefaultGroups        []types.String             `tfsdk:"default_groups"`
		DefaultGroupsDetails []GroupNameDataSourceModel `tfsdk:"default_groups_details"`
		Defined              types.Bool                 `tfsdk:"defined"`
		GroupDetails         []GroupNameDataSourceModel `tfsdk:"group_details"`
		Groups               []types.String             `tfsdk:"groups"`
		HasUnlimitedSeats    types.Bool                 `tfsdk:"has_unlimited_seats"`
		Key                  types.String               `tfsdk:"key"`
		Name                 types.String               `tfsdk:"name"`
		NumberOfSeats        types.Int32                `tfsdk:"number_of_seats"`
		Platform             types.Bool                 `tfsdk:"platform"`
	}
	GroupNameDataSourceModel struct {
		GroupId types.String `tfsdk:"group_id"`
		Name    types.String `tfsdk:"name"`
		Self    types.String `tfsdk:"self"`
	}
	AvatarUrlsBeanDataSourceModel struct {
		A16x16 types.String `tfsdk:"a_16x16"`
		A24x24 types.String `tfsdk:"a_24x24"`
		A32x32 types.String `tfsdk:"a_32x32"`
		A48x48 types.String `tfsdk:"a_48x48"`
	}
)

func UserDtoToModel(dto dto.UserDto) UserDataSourceModel {
	model := UserDataSourceModel{
		AccountId:        types.StringValue(dto.AccountId),
		AccountType:      types.StringValue(string(dto.AccountType)),
		Active:           types.BoolValue(dto.Active),
		ApplicationRoles: make([]ApplicationRoleDataSourceModel, dto.ApplicationRoles.Size),
		AvatarUrls:       AvatarUrlsBeanDtoToModel(dto.AvatarUrls),
		DisplayName:      types.StringValue(dto.DisplayName),
		EmailAddress:     types.StringValue(dto.EmailAddress),
		Expand:           types.StringValue(dto.Expand),
		Groups:           make([]GroupNameDataSourceModel, dto.Groups.Size),
		Locale:           types.StringValue(dto.Locale),
		TimeZone:         types.StringValue(dto.TimeZone),
	}
	for i, applicationRole := range dto.ApplicationRoles.Items {
		model.ApplicationRoles[i] = ApplicationRoleDtoToModel(applicationRole)
	}
	for i, group := range dto.Groups.Items {
		model.Groups[i] = GroupNameDtoToModel(group)
	}
	return model
}

func AvatarUrlsBeanDtoToModel(dto dto.AvatarUrlsBeanDto) *AvatarUrlsBeanDataSourceModel {
	return &AvatarUrlsBeanDataSourceModel{
		A16x16: types.StringValue(dto.A16x16),
		A24x24: types.StringValue(dto.A24x24),
		A32x32: types.StringValue(dto.A32x32),
		A48x48: types.StringValue(dto.A48x48),
	}
}

func GroupNameDtoToModel(dto dto.GroupNameDto) GroupNameDataSourceModel {
	return GroupNameDataSourceModel{
		GroupId: types.StringValue(dto.GroupId),
		Name:    types.StringValue(dto.Name),
		Self:    types.StringValue(dto.Self),
	}
}

func ApplicationRoleDtoToModel(dto dto.ApplicationRoleDto) ApplicationRoleDataSourceModel {
	model := ApplicationRoleDataSourceModel{
		DefaultGroups:        make([]types.String, len(dto.DefaultGroups)),
		DefaultGroupsDetails: make([]GroupNameDataSourceModel, len(dto.DefaultGroupsDetails)),
		Defined:              types.BoolValue(dto.Defined),
		GroupDetails:         make([]GroupNameDataSourceModel, len(dto.GroupDetails)),
		Groups:               make([]types.String, len(dto.Groups)),
		HasUnlimitedSeats:    types.BoolValue(dto.HasUnlimitedSeats),
		Key:                  types.StringValue(dto.Key),
		Name:                 types.StringValue(dto.Name),
		NumberOfSeats:        types.Int32Value(dto.NumberOfSeats),
		Platform:             types.BoolValue(dto.Platform),
	}
	for i, defaultGroup := range dto.DefaultGroups {
		model.DefaultGroups[i] = types.StringValue(defaultGroup)
	}
	for i, defaultGroupDetail := range dto.DefaultGroupsDetails {
		model.DefaultGroupsDetails[i] = GroupNameDtoToModel(defaultGroupDetail)
	}
	for i, group := range dto.GroupDetails {
		model.GroupDetails[i] = GroupNameDtoToModel(group)
	}
	for i, group := range dto.Groups {
		model.Groups[i] = types.StringValue(group)
	}
	return model
}
