package provider

import (
	"context"
	"github.com/atlassian/terraform-provider-jsm-ops/internal/dto"
	"github.com/atlassian/terraform-provider-jsm-ops/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func ResponderInfoDtoToModel(dto dto.ResponderInfo) dataModels.ResponderInfoModel {
	model := dataModels.ResponderInfoModel{
		Id:   types.StringNull(),
		Type: types.StringValue(string(dto.Type)),
	}
	if dto.Id != nil {
		model.Id = types.StringValue(*dto.Id)
	}
	return model
}

func RotationDtoToModel(scheduleId string, dto dto.Rotation) dataModels.RotationModel {
	model := dataModels.RotationModel{
		Id:              types.StringValue(dto.Id),
		ScheduleId:      types.StringValue(scheduleId),
		Name:            types.StringValue(dto.Name),
		StartDate:       types.StringValue(dto.StartDate),
		EndDate:         types.StringValue(dto.EndDate),
		Type:            types.StringValue(string(dto.Type)),
		Length:          types.Int32Value(dto.Length),
		TimeRestriction: types.ObjectNull(dataModels.TimeRestrictionModelMap),
		Participants: types.ListNull(types.ObjectType{
			AttrTypes: dataModels.ResponderInfoModelMap,
		}),
	}
	if len(dto.Participants) != 0 {
		participants := make([]attr.Value, len(dto.Participants))
		for i, participant := range dto.Participants {
			toModel := ResponderInfoDtoToModel(participant)
			participants[i] = toModel.AsValue()
		}
		model.Participants = types.ListValueMust(types.ObjectType{
			AttrTypes: dataModels.ResponderInfoModelMap,
		}, participants)
	}
	if dto.TimeRestriction != nil {
		attributes := map[string]attr.Value{
			"type":        types.StringValue(string(dto.TimeRestriction.Type)),
			"restriction": types.ObjectNull(dataModels.TimeOfDayTimeRestrictionSettingsModelMap),
			"restrictions": types.ListNull(
				types.ObjectType{AttrTypes: dataModels.WeekdayTimeRestrictionSettingsModelMap},
			),
		}

		if dto.TimeRestriction.TimeOfDayRestriction != nil {
			attributes["restriction"] = types.ObjectValueMust(
				dataModels.TimeOfDayTimeRestrictionSettingsModelMap,
				map[string]attr.Value{
					"start_hour": types.Int32Value(dto.TimeRestriction.TimeOfDayRestriction.StartHour),
					"end_hour":   types.Int32Value(dto.TimeRestriction.TimeOfDayRestriction.EndHour),
					"start_min":  types.Int32Value(dto.TimeRestriction.TimeOfDayRestriction.StartMin),
					"end_min":    types.Int32Value(dto.TimeRestriction.TimeOfDayRestriction.EndMin),
				},
			)
		}

		if dto.TimeRestriction.WeekAndTimeOfDayRestriction != nil {
			restrictions := make([]attr.Value, len(*dto.TimeRestriction.WeekAndTimeOfDayRestriction))
			for i, restriction := range *dto.TimeRestriction.WeekAndTimeOfDayRestriction {
				restrictions[i], _ = types.ObjectValue(
					dataModels.WeekdayTimeRestrictionSettingsModelMap,
					map[string]attr.Value{
						"start_day":  types.StringValue(string(restriction.StartDay)),
						"end_day":    types.StringValue(string(restriction.EndDay)),
						"start_hour": types.Int32Value(restriction.StartHour),
						"end_hour":   types.Int32Value(restriction.EndHour),
						"start_min":  types.Int32Value(restriction.StartMin),
						"end_min":    types.Int32Value(restriction.EndMin),
					},
				)
			}

			attributes["restrictions"] = types.ListValueMust(
				types.ObjectType{AttrTypes: dataModels.WeekdayTimeRestrictionSettingsModelMap},
				restrictions,
			)
		}

		model.TimeRestriction = types.ObjectValueMust(
			dataModels.TimeRestrictionModelMap,
			attributes,
		)
	}

	return model
}

func ScheduleDtoToModel(dto dto.Schedule) dataModels.ScheduleModel {
	model := dataModels.ScheduleModel{
		Id:          types.StringValue(dto.Id),
		Name:        types.StringValue(dto.Name),
		Description: types.StringValue(dto.Description),
		Timezone:    types.StringValue(dto.Timezone),
		Enabled:     types.BoolValue(dto.Enabled),
		TeamId:      types.StringValue(dto.TeamId),
	}
	rotations := make([]attr.Value, len(dto.Rotations))
	for i, rotation := range dto.Rotations {
		toModel := RotationDtoToModel(dto.Id, rotation)
		rotations[i] = toModel.AsValue()
	}
	model.Rotations = types.ListValueMust(
		types.ObjectType{AttrTypes: dataModels.RotationModelMap},
		rotations,
	)
	return model
}

func TeamDtoToModel(dto dto.TeamDto, usersDto []dto.UserDto) dataModels.TeamModel {
	model := dataModels.TeamModel{
		Description:     types.StringValue(dto.Description),
		DisplayName:     types.StringValue(dto.DisplayName),
		OrganizationId:  types.StringValue(dto.OrganizationId),
		TeamId:          types.StringValue(dto.TeamId),
		TeamType:        types.StringValue(string(dto.TeamType)),
		UserPermissions: PublicApiUserPermissionsDtoToModel(dto.UserPermissions).AsValue(),
		Member:          types.ListNull(types.ObjectType{AttrTypes: dataModels.TeamMemberModelMap}),
	}
	if len(usersDto) != 0 {
		arr := make([]attr.Value, len(usersDto))
		for i, member := range usersDto {
			toModel := UserDtoToModel(member)
			arr[i] = toModel.AsValue()
		}
		model.Member = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.UserModelMap}, arr)
	}
	return model
}

func PublicApiUserPermissionsDtoToModel(dto dto.PublicApiUserPermissions) *dataModels.PublicApiUserPermissionsModel {
	return &dataModels.PublicApiUserPermissionsModel{
		AddMembers:    types.BoolValue(dto.AddMembers),
		DeleteTeam:    types.BoolValue(dto.DeleteTeam),
		RemoveMembers: types.BoolValue(dto.RemoveMembers),
		UpdateTeam:    types.BoolValue(dto.UpdateTeam),
	}
}

func UserDtoToModel(dto dto.UserDto) dataModels.UserModel {
	model := dataModels.UserModel{
		AccountId:    types.StringValue(dto.AccountId),
		AccountType:  types.StringValue(string(dto.AccountType)),
		Active:       types.BoolValue(dto.Active),
		AvatarUrls:   AvatarUrlsBeanDtoToModel(dto.AvatarUrls).AsValue(),
		DisplayName:  types.StringValue(dto.DisplayName),
		EmailAddress: types.StringValue(dto.EmailAddress),
		Expand:       types.StringValue(dto.Expand),
		Locale:       types.StringValue(dto.Locale),
		TimeZone:     types.StringValue(dto.TimeZone),
	}
	applicationRoles := make([]attr.Value, dto.ApplicationRoles.Size)
	for i, applicationRole := range dto.ApplicationRoles.Items {
		toModel := ApplicationRoleDtoToModel(applicationRole)
		applicationRoles[i] = toModel.AsValue()
	}
	model.ApplicationRoles = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.ApplicationRoleModelMap}, applicationRoles)

	groups := make([]attr.Value, dto.Groups.Size)
	for i, group := range dto.Groups.Items {
		toModel := GroupNameDtoToModel(group)
		groups[i] = toModel.AsValue()
	}
	model.Groups = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.GroupNameModelMap}, groups)

	return model
}

func AvatarUrlsBeanDtoToModel(dto dto.AvatarUrlsBeanDto) *dataModels.AvatarUrlsBeanModel {
	return &dataModels.AvatarUrlsBeanModel{
		A16x16: types.StringValue(dto.A16x16),
		A24x24: types.StringValue(dto.A24x24),
		A32x32: types.StringValue(dto.A32x32),
		A48x48: types.StringValue(dto.A48x48),
	}
}

func GroupNameDtoToModel(dto dto.GroupNameDto) dataModels.GroupNameModel {
	return dataModels.GroupNameModel{
		GroupId: types.StringValue(dto.GroupId),
		Name:    types.StringValue(dto.Name),
		Self:    types.StringValue(dto.Self),
	}
}

func ApplicationRoleDtoToModel(dto dto.ApplicationRoleDto) dataModels.ApplicationRoleModel {
	model := dataModels.ApplicationRoleModel{
		Defined:           types.BoolValue(dto.Defined),
		HasUnlimitedSeats: types.BoolValue(dto.HasUnlimitedSeats),
		Key:               types.StringValue(dto.Key),
		Name:              types.StringValue(dto.Name),
		NumberOfSeats:     types.Int32Value(dto.NumberOfSeats),
		Platform:          types.BoolValue(dto.Platform),
	}
	defaultGroups := make([]attr.Value, len(dto.DefaultGroups))
	for i, defaultGroup := range dto.DefaultGroups {
		defaultGroups[i] = types.StringValue(defaultGroup)
	}
	model.DefaultGroups = types.ListValueMust(types.StringType, defaultGroups)

	defaultGroupDetails := make([]attr.Value, len(dto.DefaultGroupsDetails))
	for i, defaultGroupDetail := range dto.DefaultGroupsDetails {
		toModel := GroupNameDtoToModel(defaultGroupDetail)
		defaultGroupDetails[i] = toModel.AsValue()
	}
	model.DefaultGroupsDetails = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.GroupNameModelMap}, defaultGroupDetails)

	groupDetails := make([]attr.Value, len(dto.GroupDetails))
	for i, group := range dto.GroupDetails {
		toModel := GroupNameDtoToModel(group)
		groupDetails[i] = toModel.AsValue()
	}
	model.GroupDetails = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.GroupNameModelMap}, groupDetails)

	groups := make([]attr.Value, len(dto.Groups))
	for i, group := range dto.Groups {
		groups[i] = types.StringValue(group)
	}
	model.Groups = types.ListValueMust(types.StringType, groups)

	return model
}

func RotationModelToDto(ctx context.Context, model dataModels.RotationModel) dto.Rotation {
	dtoObj := dto.Rotation{
		Id:              model.Id.ValueString(),
		Name:            model.Name.ValueString(),
		StartDate:       model.StartDate.ValueString(),
		EndDate:         model.EndDate.ValueString(),
		Type:            dto.RotationType(model.Type.ValueString()),
		Length:          model.Length.ValueInt32(),
		Participants:    make([]dto.ResponderInfo, len(model.Participants.Elements())),
		TimeRestriction: nil,
	}

	if !(model.TimeRestriction.IsNull() || model.TimeRestriction.IsUnknown()) {
		var timeRestriction dataModels.TimeRestrictionModel
		model.TimeRestriction.As(ctx, &timeRestriction, basetypes.ObjectAsOptions{})
		dtoObj.TimeRestriction = TimeRestrictionModelToDto(ctx, timeRestriction)
	}

	var participants []dataModels.ResponderInfoModel
	model.Participants.ElementsAs(ctx, &participants, false)

	for i, participant := range participants {
		dtoObj.Participants[i] = ResponderInfoModelToDto(participant)
	}

	return dtoObj
}

func ResponderInfoModelToDto(model dataModels.ResponderInfoModel) dto.ResponderInfo {
	return dto.ResponderInfo{
		Id:   model.Id.ValueStringPointer(),
		Type: dto.ResponderType(model.Type.ValueString()),
	}
}

func TimeRestrictionModelToDto(ctx context.Context, model dataModels.TimeRestrictionModel) *dto.TimeRestriction {
	dtoObj := dto.TimeRestriction{
		Type: dto.TimeRestrictionType(model.Type.ValueString()),
	}
	if len(model.Restrictions.Elements()) != 0 {
		var restrictions []dataModels.WeekdayTimeRestrictionSettingsModel
		model.Restrictions.ElementsAs(ctx, &restrictions, false)

		arr := make([]dto.WeekdayTimeRestrictionSettings, len(restrictions))
		for i, restriction := range restrictions {
			arr[i] = WeekdayTimeRestrictionSettingsModelToDto(restriction)
		}

		dtoObj.WeekAndTimeOfDayRestriction = &arr
	}
	if model.Restriction.IsNull() || model.Restriction.IsUnknown() {
		var restriction dataModels.TimeOfDayTimeRestrictionSettingsModel
		model.Restriction.As(ctx, &restriction, basetypes.ObjectAsOptions{})
		dtoObj.TimeOfDayRestriction = TimeOfDayTimeRestrictionSettingsModelToDto(restriction)
	}

	return &dtoObj
}

func TimeOfDayTimeRestrictionSettingsModelToDto(model dataModels.TimeOfDayTimeRestrictionSettingsModel) *dto.TimeOfDayTimeRestrictionSettings {
	return &dto.TimeOfDayTimeRestrictionSettings{
		StartHour: model.StartHour.ValueInt32(),
		EndHour:   model.EndHour.ValueInt32(),
		StartMin:  model.StartMin.ValueInt32(),
		EndMin:    model.EndMin.ValueInt32(),
	}
}

func WeekdayTimeRestrictionSettingsModelToDto(model dataModels.WeekdayTimeRestrictionSettingsModel) dto.WeekdayTimeRestrictionSettings {
	return dto.WeekdayTimeRestrictionSettings{
		StartDay:  dto.Weekday(model.StartDay.ValueString()),
		EndDay:    dto.Weekday(model.EndDay.ValueString()),
		StartHour: model.StartHour.ValueInt32(),
		EndHour:   model.EndHour.ValueInt32(),
		StartMin:  model.StartMin.ValueInt32(),
		EndMin:    model.EndMin.ValueInt32(),
	}
}
