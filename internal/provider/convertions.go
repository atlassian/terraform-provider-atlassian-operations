package provider

import (
	"context"
	"github.com/atlassian/terraform-provider-jsm-ops/internal/dto"
	"github.com/atlassian/terraform-provider-jsm-ops/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
		Type:            types.StringValue(string(dto.Type)),
		Length:          types.Int32Value(dto.Length),
		TimeRestriction: types.ObjectNull(dataModels.TimeRestrictionModelMap),
		Participants: types.ListNull(types.ObjectType{
			AttrTypes: dataModels.ResponderInfoModelMap,
		}),
	}

	if dto.StartDate == "" {
		model.StartDate = timetypes.NewRFC3339Null()
	} else {
		model.StartDate = timetypes.NewRFC3339ValueMust(dto.StartDate)
	}

	if dto.EndDate == "" {
		model.EndDate = timetypes.NewRFC3339Null()
	} else {
		model.EndDate = timetypes.NewRFC3339ValueMust(dto.EndDate)
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
	return model
}

func EmailIntegrationTypeSpecificPropertiesModelToDto(model dataModels.TypeSpecificPropertiesModel) dto.TypeSpecificPropertiesDto {
	return dto.TypeSpecificPropertiesDto{
		EmailUsername:         model.EmailUsername.ValueString(),
		SuppressNotifications: model.SuppressNotifications.ValueBool(),
	}
}

func EmailIntegrationModelToDto(ctx context.Context, model dataModels.EmailIntegrationModel) dto.EmailIntegration {
	dtoObj := dto.EmailIntegration{
		Id:      model.Id.ValueString(),
		Name:    model.Name.ValueString(),
		Enabled: model.Enabled.ValueBool(),
		TeamId:  model.TeamId.ValueString(),
		Type:    "Email",
	}

	if !(model.TypeSpecificPropertiesModel.IsNull() || model.TypeSpecificPropertiesModel.IsUnknown()) {
		var typeSpecificProperties dataModels.TypeSpecificPropertiesModel
		model.TypeSpecificPropertiesModel.As(ctx, &typeSpecificProperties, basetypes.ObjectAsOptions{})

		dtoObj.TypeSpecificProperties = EmailIntegrationTypeSpecificPropertiesModelToDto(typeSpecificProperties)
	}

	return dtoObj
}

func EmailIntegrationTypeSpecificPropertiesDtoToModel(dto dto.TypeSpecificPropertiesDto) dataModels.TypeSpecificPropertiesModel {
	return dataModels.TypeSpecificPropertiesModel{
		EmailUsername:         types.StringValue(dto.EmailUsername),
		SuppressNotifications: types.BoolValue(dto.SuppressNotifications),
	}
}

func EmailIntegrationDtoToModel(dto dto.EmailIntegration) dataModels.EmailIntegrationModel {
	model := dataModels.EmailIntegrationModel{
		Id:      types.StringValue(dto.Id),
		Name:    types.StringValue(dto.Name),
		Enabled: types.BoolValue(dto.Enabled),
		TeamId:  types.StringValue(dto.TeamId),
	}

	toModel := EmailIntegrationTypeSpecificPropertiesDtoToModel(dto.TypeSpecificProperties)
	model.TypeSpecificPropertiesModel = toModel.AsValue()

	return model
}

func TeamDtoToModel(dto dto.TeamDto, membersDto []dto.TeamMember) dataModels.TeamModel {
	model := dataModels.TeamModel{
		Description:     types.StringValue(dto.Description),
		DisplayName:     types.StringValue(dto.DisplayName),
		OrganizationId:  types.StringValue(dto.OrganizationId),
		Id:              types.StringValue(dto.TeamId),
		SiteId:          types.StringNull(),
		TeamType:        types.StringValue(string(dto.TeamType)),
		UserPermissions: types.ObjectNull(dataModels.PublicApiUserPermissionsModelMap),
		Member:          types.SetNull(types.ObjectType{AttrTypes: dataModels.TeamMemberModelMap}),
	}

	if dto.SiteId != nil {
		model.SiteId = types.StringValue(*dto.SiteId)
	}

	if dto.UserPermissions != nil {
		model.UserPermissions = PublicApiUserPermissionsDtoToModel(*dto.UserPermissions).AsValue()
	}

	if len(membersDto) != 0 {
		arr := make([]attr.Value, len(membersDto))
		for i, member := range membersDto {
			toModel := TeamMemberDtoToModel(member)
			arr[i] = toModel.AsValue()
		}
		model.Member = types.SetValueMust(types.ObjectType{AttrTypes: dataModels.TeamMemberModelMap}, arr)
	}
	return model
}

func TeamMemberDtoToModel(teamMember dto.TeamMember) dataModels.TeamMemberModel {
	return dataModels.TeamMemberModel{
		AccountId: types.StringValue(teamMember.AccountId),
	}
}

func PublicApiUserPermissionsDtoToModel(dto dto.PublicApiUserPermissions) dataModels.PublicApiUserPermissionsModel {
	return dataModels.PublicApiUserPermissionsModel{
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
	if !(model.Restriction.IsNull() || model.Restriction.IsUnknown()) {
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

func TeamModelToDto(ctx context.Context, model dataModels.TeamModel) (dto.TeamDto, []dto.TeamMember) {
	userPermissions := dataModels.PublicApiUserPermissionsModel{}
	model.UserPermissions.As(ctx, &userPermissions, basetypes.ObjectAsOptions{})

	teamDtoObj := dto.TeamDto{
		Description:     model.Description.ValueString(),
		DisplayName:     model.DisplayName.ValueString(),
		OrganizationId:  model.OrganizationId.ValueString(),
		TeamId:          model.Id.ValueString(),
		SiteId:          nil,
		TeamType:        dto.TeamType(model.TeamType.ValueString()),
		UserPermissions: nil,
	}

	if !(model.SiteId.IsNull() || model.SiteId.IsUnknown()) {
		teamDtoObj.SiteId = model.SiteId.ValueStringPointer()
	}

	if !(model.UserPermissions.IsNull() || model.UserPermissions.IsUnknown()) {
		model := PublicApiUserPermissionsModelToDto(userPermissions)
		teamDtoObj.UserPermissions = &model
	}

	membersModel := make([]dataModels.TeamMemberModel, len(model.Member.Elements()))
	model.Member.ElementsAs(ctx, &membersModel, false)

	membersDto := make([]dto.TeamMember, len(model.Member.Elements()))
	for i, member := range membersModel {
		membersDto[i] = TeamMemberModelToDto(member)
	}

	return teamDtoObj, membersDto
}

func TeamMemberModelToDto(memberModel dataModels.TeamMemberModel) dto.TeamMember {
	return dto.TeamMember{
		AccountId: memberModel.AccountId.ValueString(),
	}
}

func PublicApiUserPermissionsModelToDto(userPermissions dataModels.PublicApiUserPermissionsModel) dto.PublicApiUserPermissions {
	return dto.PublicApiUserPermissions{
		AddMembers:    userPermissions.AddMembers.ValueBool(),
		DeleteTeam:    userPermissions.DeleteTeam.ValueBool(),
		RemoveMembers: userPermissions.RemoveMembers.ValueBool(),
		UpdateTeam:    userPermissions.UpdateTeam.ValueBool(),
	}
}

func ScheduleModelToDto(model dataModels.ScheduleModel) dto.Schedule {
	dtoObj := dto.Schedule{
		Id:          model.Id.ValueString(),
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Timezone:    model.Timezone.ValueString(),
		Enabled:     model.Enabled.ValueBool(),
		TeamId:      model.TeamId.ValueString(),
	}

	return dtoObj
}

func EscalationRepeatModelToDto(model dataModels.EscalationRepeatModel) dto.EscalationRepeatDto {
	return dto.EscalationRepeatDto{
		WaitInterval:         model.WaitInterval.ValueInt32(),
		Count:                model.Count.ValueInt32(),
		ResetRecipientStates: model.ResetRecipientStates.ValueBool(),
		CloseAlertAfterAll:   model.CloseAlertAfterAll.ValueBool(),
	}
}

func EscalationRuleResponseRecipientModelToDto(model dataModels.EscalationRuleResponseRecipientModel) dto.EscalationRuleRecipientDto {
	return dto.EscalationRuleRecipientDto{
		Id:   model.Id.ValueString(),
		Type: model.Type.ValueString(),
	}
}

func EscalationRuleResponseModelToDto(ctx context.Context, model dataModels.EscalationRuleResponseModel) dto.EscalationRuleDto {
	var recipient dataModels.EscalationRuleResponseRecipientModel
	model.Recipient.As(ctx, &recipient, basetypes.ObjectAsOptions{})

	return dto.EscalationRuleDto{
		Condition:  model.Condition.ValueString(),
		NotifyType: model.NotifyType.ValueString(),
		Delay:      model.Delay.ValueInt64(),
		Recipient:  EscalationRuleResponseRecipientModelToDto(recipient),
	}
}

func EscalationModelToDto(ctx context.Context, model dataModels.EscalationModel) dto.EscalationDto {
	dtoObj := dto.EscalationDto{
		Id:          model.Id.ValueString(),
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Enabled:     model.Enabled.ValueBool(),
		Rules:       make([]dto.EscalationRuleDto, len(model.Rules.Elements())),
		Repeat:      nil,
	}

	if !(model.Repeat.IsNull() || model.Repeat.IsUnknown()) {
		var repeat dataModels.EscalationRepeatModel
		model.Repeat.As(ctx, &repeat, basetypes.ObjectAsOptions{})

		var dtoRepeat = EscalationRepeatModelToDto(repeat)
		dtoObj.Repeat = &dtoRepeat
	}

	rules := make([]dataModels.EscalationRuleResponseModel, len(model.Rules.Elements()))
	model.Rules.ElementsAs(ctx, &rules, false)

	for i, rule := range rules {
		dtoObj.Rules[i] = EscalationRuleResponseModelToDto(ctx, rule)
	}

	return dtoObj
}

func EscalationRepeatDtoToModel(dto dto.EscalationRepeatDto) dataModels.EscalationRepeatModel {
	return dataModels.EscalationRepeatModel{
		WaitInterval:         types.Int32Value(dto.WaitInterval),
		Count:                types.Int32Value(dto.Count),
		ResetRecipientStates: types.BoolValue(dto.ResetRecipientStates),
		CloseAlertAfterAll:   types.BoolValue(dto.CloseAlertAfterAll),
	}
}

func EscalationRuleResponseDtoToModel(dto dto.EscalationRuleDto) dataModels.EscalationRuleResponseModel {
	model := dataModels.EscalationRuleResponseModel{
		Condition:  types.StringValue(dto.Condition),
		NotifyType: types.StringValue(dto.NotifyType),
		Delay:      types.Int64Value(dto.Delay),
	}
	responseRecipientModel := EscalationRuleResponseRecipientDtoToModel(dto.Recipient)
	model.Recipient = responseRecipientModel.AsValue()

	return model
}

func EscalationRuleResponseRecipientDtoToModel(dto dto.EscalationRuleRecipientDto) dataModels.EscalationRuleResponseRecipientModel {
	return dataModels.EscalationRuleResponseRecipientModel{
		Id:   types.StringValue(dto.Id),
		Type: types.StringValue(dto.Type),
	}
}

func EscalationDtoToModel(teamId string, dto dto.EscalationDto) dataModels.EscalationModel {
	model := dataModels.EscalationModel{
		Id:          types.StringValue(dto.Id),
		TeamId:      types.StringValue(teamId),
		Name:        types.StringValue(dto.Name),
		Description: types.StringValue(dto.Description),
		Enabled:     types.BoolValue(dto.Enabled),
		Repeat:      types.ObjectNull(dataModels.EscalationRepeatModelMap),
	}

	if dto.Repeat != nil {
		toModel := EscalationRepeatDtoToModel(*dto.Repeat)
		model.Repeat = toModel.AsValue()
	}

	rules := make([]attr.Value, len(dto.Rules))
	for i, rule := range dto.Rules {
		toModel := EscalationRuleResponseDtoToModel(rule)
		rules[i] = toModel.AsValue()
	}
	model.Rules = types.SetValueMust(types.ObjectType{AttrTypes: dataModels.EscalationRuleResponseModelMap}, rules)

	return model
}
