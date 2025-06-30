package provider

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
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

	participants := make([]attr.Value, len(dto.Participants))
	if len(dto.Participants) != 0 {
		for i, participant := range dto.Participants {
			toModel := ResponderInfoDtoToModel(participant)
			participants[i] = toModel.AsValue()
		}
	}
	model.Participants = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.ResponderInfoModelMap}, participants)

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

func EmailIntegrationMaintenanceSourcesIntervalDtoToModel(dto dto.MaintenanceInterval) dataModels.MaintenanceIntervalModel {
	return dataModels.MaintenanceIntervalModel{
		StartTimeMillis: types.Int64Value(dto.StartTimeMillis),
		EndTimeMillis:   types.Int64Value(dto.EndTimeMillis),
	}
}

func EmailIntegrationMaintenanceSourcesDtoToModel(dto dto.MaintenanceSource) dataModels.MaintenanceSourceModel {
	model := dataModels.MaintenanceSourceModel{
		MaintenanceId: types.StringValue(dto.MaintenanceId),
		Enabled:       types.BoolValue(dto.Enabled),
	}

	responseIntervalModel := EmailIntegrationMaintenanceSourcesIntervalDtoToModel(dto.Interval)
	model.Interval = responseIntervalModel.AsValue()

	return model
}

func EmailIntegrationDtoToModel(dto dto.EmailIntegration) dataModels.EmailIntegrationModel {
	model := dataModels.EmailIntegrationModel{
		Id:       types.StringValue(dto.Id),
		Name:     types.StringValue(dto.Name),
		Enabled:  types.BoolValue(dto.Enabled),
		Advanced: types.BoolValue(dto.Advanced),
		TeamId:   types.StringValue(dto.TeamId),
	}

	toModel := EmailIntegrationTypeSpecificPropertiesDtoToModel(dto.TypeSpecificProperties)
	model.TypeSpecificPropertiesModel = toModel.AsValue()

	directions := make([]attr.Value, len(dto.Directions))
	for i, direction := range dto.Directions {
		directions[i] = types.StringValue(direction)
	}
	model.Directions, _ = types.ListValue(types.StringType, directions)

	domains := make([]attr.Value, len(dto.Domains))
	for i, domain := range dto.Domains {
		domains[i] = types.StringValue(domain)
	}
	model.Domains, _ = types.ListValue(types.StringType, domains)

	maintenanceSources := make([]attr.Value, len(dto.MaintenanceSources))
	for i, maintenanceSource := range dto.MaintenanceSources {
		toModel := EmailIntegrationMaintenanceSourcesDtoToModel(maintenanceSource)
		maintenanceSources[i] = toModel.AsValue()
	}
	model.MaintenanceSources, _ = types.ListValue(types.ObjectType{AttrTypes: dataModels.IntegrationMaintenanceSourcesResponseModelMap}, maintenanceSources)

	return model
}

func TeamDtoToModel(dto dto.TeamDto, membersDto []dto.TeamMember, deleteDefaultResources types.Bool) dataModels.TeamModel {
	model := dataModels.TeamModel{
		Description:            types.StringValue(dto.Description),
		DisplayName:            types.StringValue(dto.DisplayName),
		OrganizationId:         types.StringValue(dto.OrganizationId),
		Id:                     types.StringValue(dto.TeamId),
		SiteId:                 types.StringNull(),
		TeamType:               types.StringValue(string(dto.TeamType)),
		UserPermissions:        types.ObjectNull(dataModels.PublicApiUserPermissionsModelMap),
		Member:                 types.SetNull(types.ObjectType{AttrTypes: dataModels.TeamMemberModelMap}),
		DeleteDefaultResources: deleteDefaultResources,
	}

	if dto.SiteId != nil {
		model.SiteId = types.StringValue(*dto.SiteId)
	}

	if dto.UserPermissions != nil {
		model.UserPermissions = PublicApiUserPermissionsDtoToModel(*dto.UserPermissions).AsValue()
	}

	arr := make([]attr.Value, len(membersDto))
	if len(membersDto) != 0 {
		for i, member := range membersDto {
			toModel := TeamMemberDtoToModel(member)
			arr[i] = toModel.AsValue()
		}
	}
	model.Member = types.SetValueMust(types.ObjectType{AttrTypes: dataModels.TeamMemberModelMap}, arr)

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

func OrgUserDtoToModel(userDto dto.OrgUserDto, conf dataModels.UserModel) dataModels.UserModel {
	model := dataModels.UserModel{
		AccountId:        types.StringValue(userDto.AccountId),
		AccountType:      types.StringValue(string(userDto.AccountType)),
		ApplicationRoles: types.ListNull(types.ObjectType{AttrTypes: dataModels.ApplicationRoleModelMap}),
		EmailAddress:     types.StringValue(userDto.Email),
		OrganizationId:   conf.OrganizationId,
		Expand:           types.StringNull(),
		Groups:           types.ListNull(types.ObjectType{AttrTypes: dataModels.GroupNameModelMap}),
		Locale:           types.StringNull(),
		TimeZone:         types.StringNull(),
	}

	if userDto.Nickname != "" {
		model.DisplayName = types.StringValue(userDto.Nickname)
	} else {
		model.DisplayName = types.StringValue(userDto.Name)
	}

	if userDto.Avatar != "" {
		model.AvatarUrls = AvatarUrlsBeanDtoToModel(dto.AvatarUrlsBeanDto{
			A16x16: userDto.Avatar,
			A24x24: userDto.Avatar,
			A32x32: userDto.Avatar,
			A48x48: userDto.Avatar,
		}).AsValue()
	} else {
		model.AvatarUrls = AvatarUrlsBeanDtoToModel(dto.AvatarUrlsBeanDto{
			A16x16: userDto.Picture,
			A24x24: userDto.Picture,
			A32x32: userDto.Picture,
			A48x48: userDto.Picture,
		}).AsValue()
	}

	if userDto.AccountStatus == dto.OrgUserActive {
		model.Active = types.BoolValue(true)
	} else {
		model.Active = types.BoolValue(false)
	}

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

func ApiIntegrationMaintenanceSourceIntervalModelToDto(model dataModels.MaintenanceIntervalModel) dto.MaintenanceInterval {
	return dto.MaintenanceInterval{
		StartTimeMillis: model.StartTimeMillis.ValueInt64(),
		EndTimeMillis:   model.EndTimeMillis.ValueInt64(),
	}
}

func ApiIntegrationMaintenanceSourceModelToDto(ctx context.Context, model dataModels.MaintenanceSourceModel) dto.MaintenanceSource {
	intervalModel := dataModels.MaintenanceIntervalModel{}
	model.Interval.As(ctx, &intervalModel, basetypes.ObjectAsOptions{})

	return dto.MaintenanceSource{
		MaintenanceId: model.MaintenanceId.ValueString(),
		Enabled:       model.Enabled.ValueBool(),
		Interval:      ApiIntegrationMaintenanceSourceIntervalModelToDto(intervalModel),
	}
}

func ApiIntegrationModelToDto(ctx context.Context, model dataModels.ApiIntegrationModel) dto.ApiIntegration {
	maintenanceSources := make([]dataModels.MaintenanceSourceModel, len(model.MaintenanceSources.Elements()))
	model.MaintenanceSources.ElementsAs(ctx, &maintenanceSources, false)

	directions := make([]types.String, len(model.Directions.Elements()))
	model.Directions.ElementsAs(ctx, &directions, false)

	domains := make([]types.String, len(model.Domains.Elements()))
	model.Domains.ElementsAs(ctx, &domains, false)

	typeSpecificProperties := make(map[string]interface{})
	if !(model.TypeSpecificProperties.IsNull() || model.TypeSpecificProperties.IsUnknown()) {
		model.TypeSpecificProperties.Unmarshal(&typeSpecificProperties)
	}

	dtoObj := dto.ApiIntegration{
		Id:                     model.Id.ValueString(),
		Name:                   model.Name.ValueString(),
		ApiKey:                 model.ApiKey.ValueString(),
		Type:                   model.Type.ValueString(),
		Enabled:                model.Enabled.ValueBool(),
		TeamId:                 model.TeamId.ValueString(),
		Advanced:               model.Advanced.ValueBool(),
		MaintenanceSources:     make([]dto.MaintenanceSource, len(maintenanceSources)),
		Directions:             make([]string, len(directions)),
		Domains:                make([]string, len(domains)),
		TypeSpecificProperties: typeSpecificProperties,
	}

	for i, maintenanceSource := range maintenanceSources {
		dtoObj.MaintenanceSources[i] = ApiIntegrationMaintenanceSourceModelToDto(ctx, maintenanceSource)
	}

	for i, direction := range directions {
		dtoObj.Directions[i] = direction.ValueString()
	}

	for i, domain := range domains {
		dtoObj.Domains[i] = domain.ValueString()
	}

	return dtoObj
}

func ApiIntegrationMaintenanceSourceIntervalDtoToModel(dto dto.MaintenanceInterval) dataModels.MaintenanceIntervalModel {
	return dataModels.MaintenanceIntervalModel{
		StartTimeMillis: types.Int64Value(dto.StartTimeMillis),
		EndTimeMillis:   types.Int64Value(dto.EndTimeMillis),
	}
}

func ApiIntegrationMaintenanceSourceDtoToModel(dto dto.MaintenanceSource) dataModels.MaintenanceSourceModel {
	interval := ApiIntegrationMaintenanceSourceIntervalDtoToModel(dto.Interval)
	return dataModels.MaintenanceSourceModel{
		MaintenanceId: types.StringValue(dto.MaintenanceId),
		Enabled:       types.BoolValue(dto.Enabled),
		Interval:      interval.AsValue(),
	}
}

func ApiIntegrationDtoToModel(dtoObj dto.ApiIntegration, oldModel dataModels.ApiIntegrationModel) dataModels.ApiIntegrationModel {
	typeSpecificProperties, _ := json.Marshal(dtoObj.TypeSpecificProperties)
	model := dataModels.ApiIntegrationModel{
		Id:                     types.StringValue(dtoObj.Id),
		Name:                   types.StringValue(dtoObj.Name),
		Type:                   types.StringValue(dtoObj.Type),
		Enabled:                types.BoolValue(dtoObj.Enabled),
		TeamId:                 types.StringValue(dtoObj.TeamId),
		Advanced:               types.BoolValue(dtoObj.Advanced),
		MaintenanceSources:     types.ListNull(types.ObjectType{AttrTypes: dataModels.IntegrationMaintenanceSourcesResponseModelMap}),
		Directions:             types.ListNull(types.StringType),
		Domains:                types.ListNull(types.StringType),
		TypeSpecificProperties: jsontypes.NewExactValue(string(typeSpecificProperties)),
	}

	if dtoObj.ApiKey != "" {
		model.ApiKey = types.StringValue(dtoObj.ApiKey)
	} else if !(oldModel.ApiKey.IsNull() || oldModel.ApiKey.IsUnknown()) {
		model.ApiKey = types.StringValue(oldModel.ApiKey.ValueString())
	} else {
		model.ApiKey = types.StringNull()
	}

	maintenanceSources := make([]attr.Value, len(dtoObj.MaintenanceSources))
	if len(dtoObj.MaintenanceSources) != 0 {
		for i, maintenanceSource := range dtoObj.MaintenanceSources {
			maintenanceSourceModel := ApiIntegrationMaintenanceSourceDtoToModel(maintenanceSource)
			maintenanceSources[i] = maintenanceSourceModel.AsValue()
		}
	}
	model.MaintenanceSources = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.IntegrationMaintenanceSourcesResponseModelMap}, maintenanceSources)

	directions := make([]attr.Value, len(dtoObj.Directions))
	if len(dtoObj.Directions) != 0 {
		for i, direction := range dtoObj.Directions {
			directions[i] = types.StringValue(direction)
		}
	}
	model.Directions = types.ListValueMust(types.StringType, directions)

	domains := make([]attr.Value, len(dtoObj.Domains))
	if len(dtoObj.Domains) != 0 {
		for i, domain := range dtoObj.Domains {
			domains[i] = types.StringValue(domain)
		}
	}
	model.Domains = types.ListValueMust(types.StringType, domains)

	return model
}

func CriteriaConditionModelToDto(model dataModels.CriteriaConditionModel) dto.CriteriaConditionDto {
	return dto.CriteriaConditionDto{
		Field:         model.Field.ValueString(),
		Operation:     model.Operation.ValueString(),
		ExpectedValue: model.ExpectedValue.ValueString(),
		Key:           model.Key.ValueString(),
		Not:           model.Not.ValueBool(),
		Order:         int(model.Order.ValueInt64()),
	}
}

func RoutingRuleNotifyModelToDto(model dataModels.RoutingRuleNotifyModel) *dto.RoutingRuleNotifyDto {
	return &dto.RoutingRuleNotifyDto{
		Type: model.Type.ValueString(),
		ID:   model.ID.ValueString(),
	}
}

func CriteriaModelToDto(ctx context.Context, model dataModels.CriteriaModel) *dto.CriteriaDto {
	dtoObj := dto.CriteriaDto{
		Type: dto.CriteriaType(model.Type.ValueString()),
	}

	if dtoObj.Type != dto.MatchAll {
		var conditions []dataModels.CriteriaConditionModel
		model.Conditions.ElementsAs(ctx, &conditions, false)

		arr := make([]dto.CriteriaConditionDto, len(conditions))
		for i, restriction := range conditions {
			arr[i] = CriteriaConditionModelToDto(restriction)
		}

		dtoObj.Conditions = &arr
	}

	return &dtoObj
}

func RoutingRuleModelToDto(ctx context.Context, model dataModels.RoutingRuleModel) dto.RoutingRuleDto {
	dtoObj := dto.RoutingRuleDto{
		ID:              model.ID.ValueString(),
		Name:            model.Name.ValueString(),
		Order:           model.Order.ValueInt64(),
		IsDefault:       model.IsDefault.ValueBool(),
		Timezone:        model.Timezone.ValueString(),
		Criteria:        nil,
		TimeRestriction: nil,
		Notify:          nil,
	}

	if !(model.TimeRestriction.IsNull() || model.TimeRestriction.IsUnknown()) {
		var timeRestriction dataModels.TimeRestrictionModel
		model.TimeRestriction.As(ctx, &timeRestriction, basetypes.ObjectAsOptions{})
		dtoObj.TimeRestriction = TimeRestrictionModelToDto(ctx, timeRestriction)
	}

	if !(model.Criteria.IsNull() || model.Criteria.IsUnknown()) {
		var criteria dataModels.CriteriaModel
		model.Criteria.As(ctx, &criteria, basetypes.ObjectAsOptions{})
		dtoObj.Criteria = CriteriaModelToDto(ctx, criteria)
	}

	if !(model.Notify.IsNull() || model.Notify.IsUnknown()) {
		var notify dataModels.RoutingRuleNotifyModel
		model.Notify.As(ctx, &notify, basetypes.ObjectAsOptions{})
		dtoObj.Notify = RoutingRuleNotifyModelToDto(notify)
	}

	return dtoObj
}

func RoutingRuleDtoToModel(teamId string, dto dto.RoutingRuleDto) dataModels.RoutingRuleModel {
	model := dataModels.RoutingRuleModel{
		ID:              types.StringValue(dto.ID),
		TeamID:          types.StringValue(teamId),
		Name:            types.StringValue(dto.Name),
		Order:           types.Int64Value(dto.Order),
		IsDefault:       types.BoolValue(dto.IsDefault),
		Timezone:        types.StringValue(dto.Timezone),
		TimeRestriction: types.ObjectNull(dataModels.TimeRestrictionModelMap),
		Criteria:        types.ObjectNull(dataModels.CriteriaModelMap),
		Notify:          types.ObjectNull(dataModels.RoutingRuleNotifyModelMap),
	}

	if dto.Criteria != nil {
		attributes := map[string]attr.Value{
			"type": types.StringValue(string(dto.Criteria.Type)),
			"conditions": types.ListNull(
				types.ObjectType{AttrTypes: dataModels.ConditionModelMap},
			),
		}

		if dto.Criteria.Conditions != nil {
			conditions := make([]attr.Value, len(*dto.Criteria.Conditions))
			for i, condition := range *dto.Criteria.Conditions {
				conditions[i], _ = types.ObjectValue(
					dataModels.ConditionModelMap,
					map[string]attr.Value{
						"field":          types.StringValue(condition.Field),
						"operation":      types.StringValue(condition.Operation),
						"expected_value": types.StringValue(condition.ExpectedValue),
						"key":            types.StringValue(condition.Key),
						"not":            types.BoolValue(condition.Not),
						"order":          types.Int64Value(int64(condition.Order)),
					},
				)
			}

			attributes["conditions"] = types.ListValueMust(
				types.ObjectType{AttrTypes: dataModels.ConditionModelMap},
				conditions,
			)
		}

		model.Criteria = types.ObjectValueMust(
			dataModels.CriteriaModelMap,
			attributes,
		)
	}

	if dto.Notify != nil {
		model.Notify = types.ObjectValueMust(
			dataModels.RoutingRuleNotifyModelMap,
			map[string]attr.Value{
				"type": types.StringValue(strings.ToLower(dto.Notify.Type)),
				"id":   types.StringValue(dto.Notify.ID),
			},
		)
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

func CriteriaDtoToModel(dto *dto.CriteriaDto) dataModels.CriteriaModel {
	model := dataModels.CriteriaModel{
		Type: types.StringValue(string(dto.Type)),
		Conditions: types.ListNull(
			types.ObjectType{AttrTypes: dataModels.ConditionModelMap},
		),
	}

	if dto.Conditions != nil {
		conditions := make([]attr.Value, len(*dto.Conditions))
		for i, condition := range *dto.Conditions {
			conditions[i] = types.ObjectValueMust(
				dataModels.ConditionModelMap,
				map[string]attr.Value{
					"field":          types.StringValue(condition.Field),
					"operation":      types.StringValue(condition.Operation),
					"expected_value": types.StringValue(condition.ExpectedValue),
					"key":            types.StringValue(condition.Key),
					"not":            types.BoolValue(condition.Not),
					"order":          types.Int64Value(int64(condition.Order)),
				},
			)
		}

		model.Conditions = types.ListValueMust(
			types.ObjectType{AttrTypes: dataModels.ConditionModelMap},
			conditions,
		)
	}

	return model
}

func NotificationRuleModelToDto(ctx context.Context, model dataModels.NotificationRuleModel) (dto.NotificationRuleDto, error) {
	var notificationTimes []string
	if !(model.NotificationTime.IsNull() || model.NotificationTime.IsUnknown()) {
		if !(model.ActionType.ValueString() == "schedule-start" || model.ActionType.ValueString() == "schedule-end") {
			return dto.NotificationRuleDto{}, errors.New("schedules are only available for schedule-start and schedule-end action types")

		}
		var times []string
		diags := model.NotificationTime.ElementsAs(ctx, &times, false)
		if diags.HasError() {
			return dto.NotificationRuleDto{}, errors.New("failed to convert notification times")
		}
		notificationTimes = times
	}

	var schedules []string
	if !(model.Schedules.IsNull() || model.Schedules.IsUnknown()) {
		if !(model.ActionType.ValueString() == "schedule-start" || model.ActionType.ValueString() == "schedule-end") {
			return dto.NotificationRuleDto{}, errors.New("schedules are only available for schedule-start and schedule-end action types")

		}
		var scheds []string
		diags := model.Schedules.ElementsAs(ctx, &scheds, false)
		if diags.HasError() {
			return dto.NotificationRuleDto{}, errors.New("failed to convert schedules")
		}
		schedules = scheds
	}

	var timeRestriction *dto.TimeRestriction
	if !(model.TimeRestriction.IsNull() || model.TimeRestriction.IsUnknown()) {
		var timeRestrictionModel dataModels.TimeRestrictionModel
		diags := model.TimeRestriction.As(ctx, &timeRestrictionModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return dto.NotificationRuleDto{}, errors.New("failed to convert timeRestriction")
		}
		timeRestriction = TimeRestrictionModelToDto(ctx, timeRestrictionModel)
	}

	var steps []dto.NotificationRuleStep
	if !(model.Steps.IsNull() || model.Steps.IsUnknown()) {
		var stepsList []dataModels.NotificationRuleStepModel
		diags := model.Steps.ElementsAs(ctx, &stepsList, false)
		if diags.HasError() {
			return dto.NotificationRuleDto{}, errors.New("failed to convert steps")
		}

		for _, step := range stepsList {
			var contact dataModels.NotificationContactModel
			diags := step.Contact.As(ctx, &contact, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return dto.NotificationRuleDto{}, errors.New("failed to convert contact")
			}
			if !(model.ActionType.ValueString() == "create-alert" || model.ActionType.ValueString() == "assigned-alert") &&
				!(step.SendAfter.IsNull() || step.SendAfter.IsUnknown()) {
				return dto.NotificationRuleDto{}, errors.New("contact method is only available for create-alert and assigned-alert action types")
			}

			steps = append(steps, dto.NotificationRuleStep{
				Contact: dto.NotificationContact{
					Method: contact.Method.ValueString(),
					To:     contact.To.ValueString(),
				},
				SendAfter: step.SendAfter.ValueInt64Pointer(),
				Enabled:   step.Enabled.ValueBool(),
			})
		}
	}

	var repeat *dto.NotificationRuleRepeat
	if !(model.Repeat.IsNull() || model.Repeat.IsUnknown()) {
		if !(model.ActionType.ValueString() == "create-alert" || model.ActionType.ValueString() == "assigned-alert") {
			return dto.NotificationRuleDto{}, errors.New("repeat is only available for create-alert and assigned-alert action types")
		}

		var repeatModel dataModels.NotificationRuleRepeatModel
		diags := model.Repeat.As(ctx, &repeatModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return dto.NotificationRuleDto{}, errors.New("failed to convert repeat")
		}

		repeat = &dto.NotificationRuleRepeat{
			LoopAfter: int(repeatModel.LoopAfter.ValueInt64()),
			Enabled:   repeatModel.Enabled.ValueBool(),
		}
	}

	var criteria *dto.CriteriaDto
	if !(model.Criteria.IsNull() || model.Criteria.IsUnknown()) {
		var criteriaModel dataModels.CriteriaModel
		model.Criteria.As(ctx, &criteriaModel, basetypes.ObjectAsOptions{})
		criteria = CriteriaModelToDto(ctx, criteriaModel)
	}

	return dto.NotificationRuleDto{
		ID:               model.ID.ValueString(),
		Name:             model.Name.ValueString(),
		ActionType:       model.ActionType.ValueString(),
		NotificationTime: notificationTimes,
		TimeRestriction:  timeRestriction,
		Schedules:        schedules,
		Order:            int(model.Order.ValueInt64()),
		Steps:            steps,
		Repeat:           repeat,
		Enabled:          model.Enabled.ValueBool(),
		Criteria:         criteria,
	}, nil
}

func NotificationRuleDtoToModel(_ context.Context, dto dto.NotificationRuleDto) dataModels.NotificationRuleModel {
	var notificationTime types.Set
	if dto.NotificationTime != nil {
		elements := make([]attr.Value, len(dto.NotificationTime))
		for i, time := range dto.NotificationTime {
			elements[i] = types.StringValue(time)
		}
		notificationTime = types.SetValueMust(types.StringType, elements)
	} else {
		notificationTime = types.SetNull(types.StringType)
	}

	var schedules types.List
	if dto.Schedules != nil {
		elements := make([]attr.Value, len(dto.Schedules))
		for i, schedule := range dto.Schedules {
			elements[i] = types.StringValue(schedule)
		}
		schedules = types.ListValueMust(types.StringType, elements)
	} else {
		schedules = types.ListNull(types.StringType)
	}

	var timeRestriction types.Object
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
				restrictions[i] = types.ObjectValueMust(
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

		timeRestriction = types.ObjectValueMust(
			dataModels.TimeRestrictionModelMap,
			attributes,
		)
	} else {
		timeRestriction = types.ObjectNull(dataModels.TimeRestrictionModelMap)
	}

	var steps types.List
	if dto.Steps != nil {
		elements := make([]attr.Value, len(dto.Steps))
		for i, step := range dto.Steps {
			contact := types.ObjectValueMust(
				dataModels.NotificationContactModelMap,
				map[string]attr.Value{
					"method": types.StringValue(step.Contact.Method),
					"to":     types.StringValue(step.Contact.To),
				},
			)

			elements[i] = types.ObjectValueMust(
				dataModels.NotificationRuleStepModelMap,
				map[string]attr.Value{
					"contact":    contact,
					"send_after": types.Int64PointerValue(step.SendAfter),
					"enabled":    types.BoolValue(step.Enabled),
				},
			)
		}
		steps = types.ListValueMust(types.ObjectType{AttrTypes: dataModels.NotificationRuleStepModelMap}, elements)
	} else {
		steps = types.ListNull(types.ObjectType{AttrTypes: dataModels.NotificationRuleStepModelMap})
	}

	var repeat types.Object
	if dto.Repeat != nil {
		repeat = types.ObjectValueMust(
			dataModels.NotificationRuleRepeatModelMap,
			map[string]attr.Value{
				"loop_after": types.Int64Value(int64(dto.Repeat.LoopAfter)),
				"enabled":    types.BoolValue(dto.Repeat.Enabled),
			},
		)
	} else {
		repeat = types.ObjectNull(dataModels.NotificationRuleRepeatModelMap)
	}

	var criteria types.Object
	if dto.Criteria != nil {
		model := CriteriaDtoToModel(dto.Criteria)
		criteria = model.AsValue()
	} else {
		criteria = types.ObjectNull(dataModels.CriteriaModelMap)
	}

	return dataModels.NotificationRuleModel{
		ID:               types.StringValue(dto.ID),
		Name:             types.StringValue(dto.Name),
		ActionType:       types.StringValue(dto.ActionType),
		NotificationTime: notificationTime,
		TimeRestriction:  timeRestriction,
		Schedules:        schedules,
		Order:            types.Int64Value(int64(dto.Order)),
		Steps:            steps,
		Repeat:           repeat,
		Enabled:          types.BoolValue(dto.Enabled),
		Criteria:         criteria,
	}
}
func UserContactModelToDto(model *dataModels.UserContactModel) dto.UserContactDto {
	return dto.UserContactDto{
		ID:      model.ID.ValueString(),
		Method:  model.Method.ValueString(),
		To:      model.To.ValueString(),
		Enabled: model.Enabled.ValueBool(),
	}
}

func UserContactDtoToModel(dto *dto.UserContactDto) *dataModels.UserContactModel {
	return &dataModels.UserContactModel{
		ID:      types.StringValue(dto.ID),
		Method:  types.StringValue(dto.Method),
		To:      types.StringValue(dto.To),
		Enabled: types.BoolValue(dto.Enabled),
	}
}

func UserContactCUDDtoToModel(dto *dto.UserContactCUDResponseDto, data *dataModels.UserContactModel) *dataModels.UserContactModel {
	return &dataModels.UserContactModel{
		ID:      types.StringValue(dto.Data.ID),
		Method:  data.Method,
		To:      data.To,
		Enabled: data.Enabled,
	}
}

func UserContactReadDtoToModel(dto *dto.UserContactDataReadResponseDto) *dataModels.UserContactModel {
	return &dataModels.UserContactModel{
		ID:      types.StringValue(dto.ID),
		Method:  types.StringValue(dto.Method),
		To:      types.StringValue(dto.To),
		Enabled: types.BoolValue(dto.Status.Enabled),
	}
}

func AlertPolicyModelToDto(ctx context.Context, model *dataModels.AlertPolicyModel) (*dto.AlertPolicyDto, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model == nil {
		return nil, diags
	}

	// Convert Filter
	var filter *dto.AlertFilterDto
	if !(model.Filter.IsNull() || model.Filter.IsUnknown()) {
		var filterModel dataModels.AlertFilterModel
		diags.Append(model.Filter.As(ctx, &filterModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		filter = &dto.AlertFilterDto{
			Type: filterModel.Type.ValueString(),
		}

		// Convert Conditions
		if !(filterModel.Conditions.IsNull() || filterModel.Conditions.IsUnknown()) {
			var conditions []dataModels.AlertConditionModel
			diags.Append(filterModel.Conditions.ElementsAs(ctx, &conditions, false)...)
			if diags.HasError() {
				return nil, diags
			}

			filter.Conditions = make([]dto.AlertConditionDto, len(conditions))
			for i, condition := range conditions {
				filter.Conditions[i] = dto.AlertConditionDto{
					Field:         condition.Field.ValueString(),
					Key:           condition.Key.ValueString(),
					Not:           condition.Not.ValueBool(),
					Operation:     condition.Operation.ValueString(),
					ExpectedValue: condition.ExpectedValue.ValueString(),
					Order:         int(condition.Order.ValueInt64()),
				}
			}
		}
	}

	// Convert TimeRestriction
	var timeRestriction *dto.TimeRestrictionDto
	if !(model.TimeRestriction.IsNull() || model.TimeRestriction.IsUnknown()) {
		var trModel dataModels.AlertTimeRestrictionModel
		diags.Append(model.TimeRestriction.As(ctx, &trModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		timeRestriction = &dto.TimeRestrictionDto{
			Enabled: trModel.Enabled.ValueBool(),
		}

		if !(trModel.TimeRestrictions.IsNull() || trModel.TimeRestrictions.IsUnknown()) {
			var periods []dataModels.AlertTimeRestrictionPeriodModel
			diags.Append(trModel.TimeRestrictions.ElementsAs(ctx, &periods, false)...)
			if diags.HasError() {
				return nil, diags
			}

			timeRestriction.TimeRestrictions = make([]dto.TimeRestrictionPeriodDto, len(periods))
			for i, period := range periods {
				timeRestriction.TimeRestrictions[i] = dto.TimeRestrictionPeriodDto{
					StartHour:   int(period.StartHour.ValueInt64()),
					StartMinute: int(period.StartMinute.ValueInt64()),
					EndHour:     int(period.EndHour.ValueInt64()),
					EndMinute:   int(period.EndMinute.ValueInt64()),
				}
			}
		}
	}

	// Convert Responders
	var responders []dto.ResponderDto
	if !(model.Responders.IsNull() || model.Responders.IsUnknown()) {
		var responderModels []dataModels.AlertResponderModel
		diags.Append(model.Responders.ElementsAs(ctx, &responderModels, false)...)
		if diags.HasError() {
			return nil, diags
		}

		responders = make([]dto.ResponderDto, len(responderModels))
		for i, responder := range responderModels {
			responders[i] = dto.ResponderDto{
				Type: responder.Type.ValueString(),
				ID:   responder.ID.ValueString(),
			}
		}
	}

	// Convert Actions
	var actions []string
	if !(model.Actions.IsNull() || model.Actions.IsUnknown()) {
		var actionModels []types.String
		diags.Append(model.Actions.ElementsAs(ctx, &actionModels, false)...)
		if diags.HasError() {
			return nil, diags
		}

		actions = make([]string, len(actionModels))
		for i, tag := range actionModels {
			actions[i] = tag.ValueString()
		}
	}

	// Convert Tags
	var tags []string
	if !(model.Tags.IsNull() || model.Tags.IsUnknown()) {
		var tagValues []types.String
		diags.Append(model.Tags.ElementsAs(ctx, &tagValues, false)...)
		if diags.HasError() {
			return nil, diags
		}

		tags = make([]string, len(tagValues))
		for i, tag := range tagValues {
			tags[i] = tag.ValueString()
		}
	}

	// Convert Details
	details := make(map[string]interface{})
	if !(model.Details.IsNull() || model.Details.IsUnknown()) {
		for k, v := range model.Details.Elements() {
			if strVal, ok := v.(basetypes.StringValue); ok {
				details[k] = strVal.ValueString()
			}
		}
	}

	alertPolicyDto := dto.AlertPolicyDto{
		ID:                     model.ID.ValueString(),
		Type:                   model.Type.ValueString(),
		Name:                   model.Name.ValueString(),
		Description:            model.Description.ValueString(),
		Enabled:                model.Enabled.ValueBool(),
		Filter:                 filter,
		TimeRestriction:        timeRestriction,
		Alias:                  model.Alias.ValueString(),
		Message:                model.Message.ValueString(),
		AlertDescription:       model.AlertDescription.ValueString(),
		Source:                 model.Source.ValueString(),
		Entity:                 model.Entity.ValueString(),
		Responders:             responders,
		Actions:                actions,
		Tags:                   tags,
		Details:                details,
		Continue:               model.Continue.ValueBool(),
		UpdatePriority:         model.UpdatePriority.ValueBool(),
		PriorityValue:          model.PriorityValue.ValueString(),
		KeepOriginalResponders: model.KeepOriginalResponders.ValueBool(),
		KeepOriginalDetails:    model.KeepOriginalDetails.ValueBool(),
		KeepOriginalActions:    model.KeepOriginalActions.ValueBool(),
		KeepOriginalTags:       model.KeepOriginalTags.ValueBool(),
	}

	if !(model.TeamID.IsNull() || model.TeamID.IsUnknown()) {
		alertPolicyDto.TeamID = model.TeamID.ValueString()
	}

	return &alertPolicyDto, diags
}

func AlertPolicyDtoToModel(_ context.Context, dto *dto.AlertPolicyDto) (*dataModels.AlertPolicyModel, error) {

	if dto == nil {
		return nil, errors.New("dto can not be nil")
	}

	// Convert Filter
	var filter types.Object
	if dto.Filter != nil {
		conditions := make([]attr.Value, len(dto.Filter.Conditions))
		for i, condition := range dto.Filter.Conditions {
			conditionCustomKey := types.StringNull()
			if condition.Key != "" {
				conditionCustomKey = types.StringValue(condition.Key)
			}
			conditions[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"field":          types.StringType,
					"key":            types.StringType,
					"not":            types.BoolType,
					"operation":      types.StringType,
					"expected_value": types.StringType,
					"order":          types.Int64Type,
				},
				map[string]attr.Value{
					"field":          types.StringValue(condition.Field),
					"key":            conditionCustomKey,
					"not":            types.BoolValue(condition.Not),
					"operation":      types.StringValue(condition.Operation),
					"expected_value": types.StringValue(condition.ExpectedValue),
					"order":          types.Int64Value(int64(condition.Order)),
				},
			)
		}

		filter = types.ObjectValueMust(
			map[string]attr.Type{
				"type": types.StringType,
				"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"field":          types.StringType,
					"key":            types.StringType,
					"not":            types.BoolType,
					"operation":      types.StringType,
					"expected_value": types.StringType,
					"order":          types.Int64Type,
				}}},
			},
			map[string]attr.Value{
				"type": types.StringValue(dto.Filter.Type),
				"conditions": types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
					"field":          types.StringType,
					"key":            types.StringType,
					"not":            types.BoolType,
					"operation":      types.StringType,
					"expected_value": types.StringType,
					"order":          types.Int64Type,
				}}, conditions),
			},
		)
	} else {
		filter = types.ObjectNull(map[string]attr.Type{
			"type": types.StringType,
			"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"field":          types.StringType,
				"key":            types.StringType,
				"not":            types.BoolType,
				"operation":      types.StringType,
				"expected_value": types.StringType,
				"order":          types.Int64Type,
			}}},
		})
	}

	// Convert TimeRestriction
	var timeRestriction types.Object
	if dto.TimeRestriction != nil {
		periods := make([]attr.Value, len(dto.TimeRestriction.TimeRestrictions))
		for i, period := range dto.TimeRestriction.TimeRestrictions {
			periods[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"start_hour":   types.Int64Type,
					"start_minute": types.Int64Type,
					"end_hour":     types.Int64Type,
					"end_minute":   types.Int64Type,
				},
				map[string]attr.Value{
					"start_hour":   types.Int64Value(int64(period.StartHour)),
					"start_minute": types.Int64Value(int64(period.StartMinute)),
					"end_hour":     types.Int64Value(int64(period.EndHour)),
					"end_minute":   types.Int64Value(int64(period.EndMinute)),
				},
			)
		}

		timeRestriction = types.ObjectValueMust(
			map[string]attr.Type{
				"enabled": types.BoolType,
				"time_restrictions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"start_hour":   types.Int64Type,
					"start_minute": types.Int64Type,
					"end_hour":     types.Int64Type,
					"end_minute":   types.Int64Type,
				}}},
			},
			map[string]attr.Value{
				"enabled": types.BoolValue(dto.TimeRestriction.Enabled),
				"time_restrictions": types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
					"start_hour":   types.Int64Type,
					"start_minute": types.Int64Type,
					"end_hour":     types.Int64Type,
					"end_minute":   types.Int64Type,
				}}, periods),
			},
		)
	} else {
		timeRestriction = types.ObjectNull(map[string]attr.Type{
			"enabled": types.BoolType,
			"time_restrictions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"start_hour":   types.Int64Type,
				"start_minute": types.Int64Type,
				"end_hour":     types.Int64Type,
				"end_minute":   types.Int64Type,
			}}},
		})
	}

	// Convert Responders
	var responders types.List
	if len(dto.Responders) > 0 {
		responderValues := make([]attr.Value, len(dto.Responders))
		for i, responder := range dto.Responders {
			responderValues[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"type": types.StringType,
					"id":   types.StringType,
				},
				map[string]attr.Value{
					"type": types.StringValue(responder.Type),
					"id":   types.StringValue(responder.ID),
				},
			)
		}
		responders = types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"id":   types.StringType,
		}}, responderValues)
	} else {
		responders = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"id":   types.StringType,
		}})
	}

	// Convert Actions
	var actions types.List
	if len(dto.Actions) > 0 {
		actionValue := make([]attr.Value, len(dto.Actions))
		for i, tag := range dto.Actions {
			actionValue[i] = types.StringValue(tag)
		}
		actions = types.ListValueMust(types.StringType, actionValue)
	} else {
		actions = types.ListNull(types.StringType)
	}

	// Convert Tags
	var tags types.List
	if len(dto.Tags) > 0 {
		tagValues := make([]attr.Value, len(dto.Tags))
		for i, tag := range dto.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		tags = types.ListNull(types.StringType)
	}

	// Convert Details
	var details types.Map
	if len(dto.Details) > 0 {
		detailValues := make(map[string]attr.Value)
		for k, v := range dto.Details {
			if str, ok := v.(string); ok {
				detailValues[k] = types.StringValue(str)
			}
		}
		details = types.MapValueMust(types.StringType, detailValues)
	} else {
		details = types.MapNull(types.StringType)
	}

	priorityValue := types.StringNull()
	if dto.PriorityValue != "" {
		priorityValue = types.StringValue(dto.PriorityValue)
	}

	alertPolicyModel := dataModels.AlertPolicyModel{
		ID:                     types.StringValue(dto.ID),
		Type:                   types.StringValue(strings.ToLower(dto.Type)),
		Name:                   types.StringValue(dto.Name),
		Description:            types.StringValue(dto.Description),
		Enabled:                types.BoolValue(dto.Enabled),
		Filter:                 filter,
		TimeRestriction:        timeRestriction,
		Alias:                  types.StringValue(dto.Alias),
		Message:                types.StringValue(dto.Message),
		AlertDescription:       types.StringValue(dto.AlertDescription),
		Source:                 types.StringValue(dto.Source),
		Entity:                 types.StringValue(dto.Entity),
		Responders:             responders,
		Actions:                actions,
		Tags:                   tags,
		Details:                details,
		Continue:               types.BoolValue(dto.Continue),
		UpdatePriority:         types.BoolValue(dto.UpdatePriority),
		PriorityValue:          priorityValue,
		KeepOriginalResponders: types.BoolValue(dto.KeepOriginalResponders),
		KeepOriginalDetails:    types.BoolValue(dto.KeepOriginalDetails),
		KeepOriginalActions:    types.BoolValue(dto.KeepOriginalActions),
		KeepOriginalTags:       types.BoolValue(dto.KeepOriginalTags),
	}

	if dto.TeamID == "" {
		alertPolicyModel.TeamID = types.StringNull()
	} else {
		alertPolicyModel.TeamID = types.StringValue(dto.TeamID)
	}

	return &alertPolicyModel, nil
}

func CustomRoleModelToDto(ctx context.Context, model *dataModels.CustomRoleModel) dto.CustomRoleDto {
	var grantedRights []string
	if !(model.GrantedRights.IsNull() || model.GrantedRights.IsUnknown()) {
		var rights []string
		diags := model.GrantedRights.ElementsAs(ctx, &rights, false)
		if diags.HasError() {
			return dto.CustomRoleDto{}
		}
		grantedRights = rights
	}

	var disallowedRights []string
	if !(model.DisallowedRights.IsNull() || model.DisallowedRights.IsUnknown()) {
		var rights []string
		diags := model.DisallowedRights.ElementsAs(ctx, &rights, false)
		if diags.HasError() {
			return dto.CustomRoleDto{}
		}
		disallowedRights = rights
	}

	return dto.CustomRoleDto{
		ID:               model.ID.ValueString(),
		Name:             model.Name.ValueString(),
		GrantedRights:    grantedRights,
		DisallowedRights: disallowedRights,
	}
}

func CustomRoleDtoToModel(dto *dto.CustomRoleDto) *dataModels.CustomRoleModel {
	var grantedRights types.Set
	if dto.GrantedRights != nil {
		elements := make([]attr.Value, len(dto.GrantedRights))
		for i, time := range dto.GrantedRights {
			elements[i] = types.StringValue(time)
		}
		grantedRights = types.SetValueMust(types.StringType, elements)
	} else {
		grantedRights = types.SetNull(types.StringType)
	}

	var disallowedRights types.Set
	if dto.DisallowedRights != nil {
		elements := make([]attr.Value, len(dto.DisallowedRights))
		for i, time := range dto.DisallowedRights {
			elements[i] = types.StringValue(time)
		}
		disallowedRights = types.SetValueMust(types.StringType, elements)
	} else {
		disallowedRights = types.SetNull(types.StringType)
	}

	return &dataModels.CustomRoleModel{
		ID:               types.StringValue(dto.ID),
		Name:             types.StringValue(dto.Name),
		GrantedRights:    grantedRights,
		DisallowedRights: disallowedRights,
	}
}

func CustomRoleCUDDtoToModel(dto *dto.CustomRoleCUDResponseDto, data *dataModels.CustomRoleModel) *dataModels.CustomRoleModel {
	return &dataModels.CustomRoleModel{
		ID: types.StringValue(dto.Data.ID),
		// JSM OPS API does not return the updated name of the custom role
		Name:             data.Name,
		GrantedRights:    data.GrantedRights,
		DisallowedRights: data.DisallowedRights,
	}
}

func NotificationPolicyModelToDto(ctx context.Context, model *dataModels.NotificationPolicyModel) (*dto.NotificationPolicyDto, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model == nil {
		return nil, diags
	}

	// Convert Filter
	var filter *dto.NotificationFilterDto
	if !(model.Filter.IsNull() || model.Filter.IsUnknown()) {
		var filterModel dataModels.NotificationFilterModel
		diags.Append(model.Filter.As(ctx, &filterModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		filter = &dto.NotificationFilterDto{
			Type: filterModel.Type.ValueString(),
		}

		if !(filterModel.Conditions.IsNull() || filterModel.Conditions.IsUnknown()) {
			var conditions []dataModels.NotificationConditionModel
			diags.Append(filterModel.Conditions.ElementsAs(ctx, &conditions, false)...)
			if diags.HasError() {
				return nil, diags
			}

			filter.Conditions = make([]dto.NotificationConditionDto, len(conditions))
			for i, condition := range conditions {
				filter.Conditions[i] = dto.NotificationConditionDto{
					Field:         condition.Field.ValueString(),
					Key:           condition.Key.ValueString(),
					Not:           condition.Not.ValueBool(),
					Operation:     condition.Operation.ValueString(),
					ExpectedValue: condition.ExpectedValue.ValueString(),
					Order:         int(condition.Order.ValueInt64()),
				}
			}
		}
	}

	// Convert TimeRestriction
	var timeRestriction *dto.NotificationPolicyTimeRestrictionDto
	if !(model.TimeRestriction.IsNull() || model.TimeRestriction.IsUnknown()) {
		var trModel dataModels.NotificationPolicyTimeRestrictionModel
		diags.Append(model.TimeRestriction.As(ctx, &trModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		timeRestriction = &dto.NotificationPolicyTimeRestrictionDto{
			Enabled: trModel.Enabled.ValueBool(),
		}

		if !(trModel.TimeRestrictions.IsNull() || trModel.TimeRestrictions.IsUnknown()) {
			var periods []dataModels.NotificationPolicyTimeRestrictionSettingsModel
			diags.Append(trModel.TimeRestrictions.ElementsAs(ctx, &periods, false)...)
			if diags.HasError() {
				return nil, diags
			}

			timeRestriction.TimeRestrictions = make([]dto.NotificationPolicyTimeRestrictionSettingsDto, len(periods))
			for i, period := range periods {
				timeRestriction.TimeRestrictions[i] = dto.NotificationPolicyTimeRestrictionSettingsDto{
					StartHour:   int(period.StartHour.ValueInt64()),
					StartMinute: int(period.StartMinute.ValueInt64()),
					EndHour:     int(period.EndHour.ValueInt64()),
					EndMinute:   int(period.EndMinute.ValueInt64()),
				}
			}
		}
	}

	// Convert AutoRestartAction
	var autoRestartAction *dto.AutoRestartActionDto
	if !(model.AutoRestartAction.IsNull() || model.AutoRestartAction.IsUnknown()) {
		var actionModel dataModels.AutoRestartActionModel
		diags.Append(model.AutoRestartAction.As(ctx, &actionModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		autoRestartAction = &dto.AutoRestartActionDto{
			WaitDuration:   int(actionModel.WaitDuration.ValueInt64()),
			MaxRepeatCount: int(actionModel.MaxRepeatCount.ValueInt64()),
			DurationFormat: actionModel.DurationFormat.ValueString(),
		}
	}

	// Convert AutoCloseAction
	var autoCloseAction *dto.AutoCloseActionDto
	if !(model.AutoCloseAction.IsNull() || model.AutoCloseAction.IsUnknown()) {
		var actionModel dataModels.AutoCloseActionModel
		diags.Append(model.AutoCloseAction.As(ctx, &actionModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		autoCloseAction = &dto.AutoCloseActionDto{
			WaitDuration:   int(actionModel.WaitDuration.ValueInt64()),
			DurationFormat: actionModel.DurationFormat.ValueString(),
		}
	}

	// Convert DeduplicationAction
	var deduplicationAction *dto.DeduplicationActionDto
	if !(model.DeduplicationAction.IsNull() || model.DeduplicationAction.IsUnknown()) {
		var actionModel dataModels.DeduplicationActionModel
		diags.Append(model.DeduplicationAction.As(ctx, &actionModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		deduplicationAction = &dto.DeduplicationActionDto{
			DeduplicationActionType: actionModel.DeduplicationActionType.ValueString(),
			Frequency:               int(actionModel.Frequency.ValueInt64()),
			CountValueLimit:         int(actionModel.CountValueLimit.ValueInt64()),
			WaitDuration:            int(actionModel.WaitDuration.ValueInt64()),
			DurationFormat:          actionModel.DurationFormat.ValueString(),
		}
	}

	// Convert DelayAction
	var delayAction *dto.DelayActionDto
	if !(model.DelayAction.IsNull() || model.DelayAction.IsUnknown()) {
		var actionModel dataModels.DelayActionModel
		diags.Append(model.DelayAction.As(ctx, &actionModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		var delayTime *dto.DelayTimeDto
		if !(actionModel.DelayTime.IsNull() || actionModel.DelayTime.IsUnknown()) {
			var delayTimeModel dataModels.DelayActionDelayTimeModel
			diags.Append(actionModel.DelayTime.As(ctx, &delayTimeModel, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}

			delayTime = &dto.DelayTimeDto{
				Hours:   int(delayTimeModel.Hours.ValueInt64()),
				Minutes: int(delayTimeModel.Minutes.ValueInt64()),
			}
		} else {
			delayTime = &dto.DelayTimeDto{}
		}

		delayAction = &dto.DelayActionDto{
			DelayTime:      delayTime,
			DelayOption:    actionModel.DelayOption.ValueString(),
			WaitDuration:   int(actionModel.WaitDuration.ValueInt64()),
			DurationFormat: actionModel.DurationFormat.ValueString(),
		}
	}

	return &dto.NotificationPolicyDto{
		ID:                  model.ID.ValueString(),
		Type:                model.Type.ValueString(),
		Name:                model.Name.ValueString(),
		Description:         model.Description.ValueString(),
		TeamID:              model.TeamID.ValueString(),
		Enabled:             model.Enabled.ValueBool(),
		Order:               model.Order.ValueFloat64(),
		Filter:              filter,
		TimeRestriction:     timeRestriction,
		AutoRestartAction:   autoRestartAction,
		AutoCloseAction:     autoCloseAction,
		DeduplicationAction: deduplicationAction,
		DelayAction:         delayAction,
		Suppress:            model.Suppress.ValueBool(),
	}, diags
}

func NotificationPolicyDtoToModel(ctx context.Context, dto *dto.NotificationPolicyDto) (*dataModels.NotificationPolicyModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if dto == nil {
		return nil, diags
	}

	// Convert Filter
	var filter types.Object
	if dto.Filter != nil {
		conditions := make([]attr.Value, len(dto.Filter.Conditions))
		for i, condition := range dto.Filter.Conditions {
			conditionCustomKey := types.StringNull()
			if condition.Key != "" {
				conditionCustomKey = types.StringValue(condition.Key)
			}
			conditions[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"field":          types.StringType,
					"key":            types.StringType,
					"not":            types.BoolType,
					"operation":      types.StringType,
					"expected_value": types.StringType,
					"order":          types.Int64Type,
				},
				map[string]attr.Value{
					"field":          types.StringValue(condition.Field),
					"key":            conditionCustomKey,
					"not":            types.BoolValue(condition.Not),
					"operation":      types.StringValue(condition.Operation),
					"expected_value": types.StringValue(condition.ExpectedValue),
					"order":          types.Int64Value(int64(condition.Order)),
				},
			)
		}

		filter = types.ObjectValueMust(
			map[string]attr.Type{
				"type": types.StringType,
				"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"field":          types.StringType,
					"key":            types.StringType,
					"not":            types.BoolType,
					"operation":      types.StringType,
					"expected_value": types.StringType,
					"order":          types.Int64Type,
				}}},
			},
			map[string]attr.Value{
				"type": types.StringValue(dto.Filter.Type),
				"conditions": types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
					"field":          types.StringType,
					"key":            types.StringType,
					"not":            types.BoolType,
					"operation":      types.StringType,
					"expected_value": types.StringType,
					"order":          types.Int64Type,
				}}, conditions),
			},
		)
	} else {
		filter = types.ObjectNull(map[string]attr.Type{
			"type": types.StringType,
			"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"field":          types.StringType,
				"key":            types.StringType,
				"not":            types.BoolType,
				"operation":      types.StringType,
				"expected_value": types.StringType,
				"order":          types.Int64Type,
			}}},
		})
	}

	// Convert TimeRestriction
	var timeRestriction types.Object
	if dto.TimeRestriction != nil {
		periods := make([]attr.Value, len(dto.TimeRestriction.TimeRestrictions))
		for i, period := range dto.TimeRestriction.TimeRestrictions {
			periods[i] = types.ObjectValueMust(
				map[string]attr.Type{
					"start_hour":   types.Int64Type,
					"start_minute": types.Int64Type,
					"end_hour":     types.Int64Type,
					"end_minute":   types.Int64Type,
				},
				map[string]attr.Value{
					"start_hour":   types.Int64Value(int64(period.StartHour)),
					"start_minute": types.Int64Value(int64(period.StartMinute)),
					"end_hour":     types.Int64Value(int64(period.EndHour)),
					"end_minute":   types.Int64Value(int64(period.EndMinute)),
				},
			)
		}

		timeRestriction = types.ObjectValueMust(
			map[string]attr.Type{
				"enabled": types.BoolType,
				"time_restrictions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"start_hour":   types.Int64Type,
					"start_minute": types.Int64Type,
					"end_hour":     types.Int64Type,
					"end_minute":   types.Int64Type,
				}}},
			},
			map[string]attr.Value{
				"enabled": types.BoolValue(dto.TimeRestriction.Enabled),
				"time_restrictions": types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
					"start_hour":   types.Int64Type,
					"start_minute": types.Int64Type,
					"end_hour":     types.Int64Type,
					"end_minute":   types.Int64Type,
				}}, periods),
			},
		)
	} else {
		timeRestriction = types.ObjectNull(map[string]attr.Type{
			"enabled": types.BoolType,
			"time_restrictions": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"start_hour":   types.Int64Type,
				"start_minute": types.Int64Type,
				"end_hour":     types.Int64Type,
				"end_minute":   types.Int64Type,
			}}},
		})
	}

	// Convert AutoRestartAction
	var autoRestartAction types.Object
	if dto.AutoRestartAction != nil {
		autoRestartAction = types.ObjectValueMust(
			map[string]attr.Type{
				"wait_duration":    types.Int64Type,
				"max_repeat_count": types.Int64Type,
				"duration_format":  types.StringType,
			},
			map[string]attr.Value{
				"wait_duration":    types.Int64Value(int64(dto.AutoRestartAction.WaitDuration)),
				"max_repeat_count": types.Int64Value(int64(dto.AutoRestartAction.MaxRepeatCount)),
				"duration_format":  types.StringValue(dto.AutoRestartAction.DurationFormat),
			},
		)
	} else {
		autoRestartAction = types.ObjectNull(map[string]attr.Type{
			"wait_duration":    types.Int64Type,
			"max_repeat_count": types.Int64Type,
			"duration_format":  types.StringType,
		})
	}

	// Convert AutoCloseAction
	var autoCloseAction types.Object
	if dto.AutoCloseAction != nil {
		autoCloseAction = types.ObjectValueMust(
			map[string]attr.Type{
				"wait_duration":   types.Int64Type,
				"duration_format": types.StringType,
			},
			map[string]attr.Value{
				"wait_duration":   types.Int64Value(int64(dto.AutoCloseAction.WaitDuration)),
				"duration_format": types.StringValue(dto.AutoCloseAction.DurationFormat),
			},
		)
	} else {
		autoCloseAction = types.ObjectNull(map[string]attr.Type{
			"wait_duration":   types.Int64Type,
			"duration_format": types.StringType,
		})
	}

	// Convert DeduplicationAction
	var deduplicationAction types.Object
	if dto.DeduplicationAction != nil {
		deduplicationAction = types.ObjectValueMust(
			map[string]attr.Type{
				"deduplication_action_type": types.StringType,
				"frequency":                 types.Int64Type,
				"count_value_limit":         types.Int64Type,
				"wait_duration":             types.Int64Type,
				"duration_format":           types.StringType,
			},
			map[string]attr.Value{
				"deduplication_action_type": types.StringValue(dto.DeduplicationAction.DeduplicationActionType),
				"frequency":                 types.Int64Value(int64(dto.DeduplicationAction.Frequency)),
				"count_value_limit":         types.Int64Value(int64(dto.DeduplicationAction.CountValueLimit)),
				"wait_duration":             types.Int64Value(int64(dto.DeduplicationAction.WaitDuration)),
				"duration_format":           types.StringValue(dto.DeduplicationAction.DurationFormat),
			},
		)
	} else {
		deduplicationAction = types.ObjectNull(map[string]attr.Type{
			"deduplication_action_type": types.StringType,
			"frequency":                 types.Int64Type,
			"count_value_limit":         types.Int64Type,
			"wait_duration":             types.Int64Type,
			"duration_format":           types.StringType,
		})
	}

	// Convert DelayAction
	var delayAction types.Object
	if dto.DelayAction != nil {
		var delayTime types.Object
		if dto.DelayAction.DelayTime != nil {
			delayTime = types.ObjectValueMust(
				map[string]attr.Type{
					"hours":   types.Int64Type,
					"minutes": types.Int64Type,
				},
				map[string]attr.Value{
					"hours":   types.Int64Value(int64(dto.DelayAction.DelayTime.Hours)),
					"minutes": types.Int64Value(int64(dto.DelayAction.DelayTime.Minutes)),
				},
			)
		}
		delayAction = types.ObjectValueMust(
			map[string]attr.Type{
				"delay_time": types.ObjectType{AttrTypes: map[string]attr.Type{
					"hours":   types.Int64Type,
					"minutes": types.Int64Type,
				}},
				"delay_option":    types.StringType,
				"wait_duration":   types.Int64Type,
				"duration_format": types.StringType,
			},
			map[string]attr.Value{
				"delay_time":      delayTime,
				"delay_option":    types.StringValue(dto.DelayAction.DelayOption),
				"wait_duration":   types.Int64Value(int64(dto.DelayAction.WaitDuration)),
				"duration_format": types.StringValue(dto.DelayAction.DurationFormat),
			},
		)
	} else {
		delayAction = types.ObjectNull(map[string]attr.Type{
			"delay_time": types.ObjectType{AttrTypes: map[string]attr.Type{
				"hours":   types.Int64Type,
				"minutes": types.Int64Type,
			}},
			"delay_option":    types.StringType,
			"wait_duration":   types.Int64Type,
			"duration_format": types.StringType,
		})
	}

	return &dataModels.NotificationPolicyModel{
		ID:                  types.StringValue(dto.ID),
		Type:                types.StringValue(strings.ToLower(dto.Type)),
		Name:                types.StringValue(dto.Name),
		Description:         types.StringValue(dto.Description),
		TeamID:              types.StringValue(dto.TeamID),
		Enabled:             types.BoolValue(dto.Enabled),
		Order:               types.Float64Value(dto.Order),
		Filter:              filter,
		TimeRestriction:     timeRestriction,
		AutoRestartAction:   autoRestartAction,
		AutoCloseAction:     autoCloseAction,
		DeduplicationAction: deduplicationAction,
		DelayAction:         delayAction,
		Suppress:            types.BoolValue(dto.Suppress),
	}, diags
}

func HeartbeatModelToDto(ctx context.Context, model *dataModels.HeartbeatModel) (*dto.HeartbeatDto, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model == nil {
		return nil, diags
	}

	// Convert alert tags
	var alertTags []string
	if !model.AlertTags.IsNull() && !model.AlertTags.IsUnknown() {
		diags.Append(model.AlertTags.ElementsAs(ctx, &alertTags, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &dto.HeartbeatDto{
		Name:          model.Name.ValueString(),
		Description:   model.Description.ValueString(),
		Interval:      int(model.Interval.ValueInt64()),
		IntervalUnit:  model.IntervalUnit.ValueString(),
		Enabled:       model.Enabled.ValueBool(),
		AlertMessage:  model.AlertMessage.ValueString(),
		AlertTags:     alertTags,
		AlertPriority: model.AlertPriority.ValueString(),
	}, diags
}

func HeartbeatDtoToModel(ctx context.Context, dto *dto.HeartbeatDto, teamID string) (*dataModels.HeartbeatModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if dto == nil {
		return nil, diags
	}

	// Convert alert tags
	var alertTagsList types.Set
	if len(dto.AlertTags) > 0 {
		alertTagsValues := []attr.Value{}
		for _, tag := range dto.AlertTags {
			alertTagsValues = append(alertTagsValues, types.StringValue(tag))
		}

		var listDiags diag.Diagnostics
		alertTagsList, listDiags = types.SetValueFrom(ctx, types.StringType, alertTagsValues)
		diags.Append(listDiags...)
		if diags.HasError() {
			return nil, diags
		}
	} else {
		alertTagsList = types.SetNull(types.StringType)
	}

	return &dataModels.HeartbeatModel{
		Name:          types.StringValue(dto.Name),
		Description:   types.StringValue(dto.Description),
		Interval:      types.Int64Value(int64(dto.Interval)),
		IntervalUnit:  types.StringValue(dto.IntervalUnit),
		Enabled:       types.BoolValue(dto.Enabled),
		Status:        types.StringValue(dto.Status),
		TeamID:        types.StringValue(teamID),
		AlertMessage:  types.StringValue(dto.AlertMessage),
		AlertTags:     alertTagsList,
		AlertPriority: types.StringValue(dto.AlertPriority),
	}, diags
}

func IntegrationActionModelToDto(ctx context.Context, model *dataModels.IntegrationActionModel) (*dto.IntegrationActionDto, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model == nil {
		return nil, diags
	}

	var filter *dto.FilterDto
	if !(model.Filter.IsNull() || model.Filter.IsUnknown()) {
		var filterModel dataModels.FilterModel
		diags.Append(model.Filter.As(ctx, &filterModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		var conditions []dto.FilterConditionDto
		if !(filterModel.Conditions.IsNull() || filterModel.Conditions.IsUnknown()) {
			elements := filterModel.Conditions.Elements()
			conditions = make([]dto.FilterConditionDto, 0, len(elements))
			for _, element := range elements {
				var condition dataModels.FilterConditionModel
				if obj, ok := element.(types.Object); ok {
					diags.Append(obj.As(ctx, &condition, basetypes.ObjectAsOptions{})...)
					if diags.HasError() {
						return nil, diags
					}
					conditions = append(conditions, dto.FilterConditionDto{
						Field:           condition.Field.ValueString(),
						Operation:       condition.Operation.ValueString(),
						ExpectedValue:   condition.ExpectedValue.ValueString(),
						Key:             condition.Key.ValueString(),
						Not:             condition.Not.ValueBool(),
						Order:           condition.Order.ValueInt64(),
						SystemCondition: condition.SystemCondition.ValueBool(),
					})
				}
			}
		}

		filter = &dto.FilterDto{
			ConditionsEmpty:    filterModel.ConditionsEmpty.ValueBool(),
			ConditionMatchType: filterModel.ConditionMatchType.ValueString(),
			Conditions:         conditions,
		}
	}

	var actionMapping *dto.ActionMappingDto
	if !(model.ActionMapping.IsNull() || model.ActionMapping.IsUnknown()) {
		var actionMappingModel dataModels.ActionMappingModel
		diags.Append(model.ActionMapping.As(ctx, &actionMappingModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		parameter := make(map[string]interface{})
		if !(actionMappingModel.Parameter.IsNull() || actionMappingModel.Parameter.IsUnknown()) {
			actionMappingModel.Parameter.Unmarshal(&parameter)
		}

		actionMapping = &dto.ActionMappingDto{
			Type:      actionMappingModel.Type.ValueString(),
			Parameter: parameter,
		}
	}

	typeSpecificProperties := make(map[string]interface{})
	if !(model.TypeSpecificProperties.IsNull() || model.TypeSpecificProperties.IsUnknown()) {
		model.TypeSpecificProperties.Unmarshal(&typeSpecificProperties)
	}

	fieldMappings := make(map[string]interface{})
	if !(model.FieldMappings.IsNull() || model.FieldMappings.IsUnknown()) {
		model.FieldMappings.Unmarshal(&fieldMappings)
	}

	return &dto.IntegrationActionDto{
		ID:                     model.ID.ValueString(),
		Type:                   model.Type.ValueString(),
		Name:                   model.Name.ValueString(),
		Domain:                 model.Domain.ValueString(),
		Direction:              model.Direction.ValueString(),
		GroupType:              model.GroupType.ValueString(),
		Filter:                 filter,
		TypeSpecificProperties: typeSpecificProperties,
		FieldMappings:          fieldMappings,
		ActionMapping:          actionMapping,
		Enabled:                model.Enabled.ValueBoolPointer(),
	}, diags
}

func IntegrationActionDtoToModel(ctx context.Context, dto *dto.IntegrationActionDto, integrationID types.String, model *dataModels.IntegrationActionModel) (*dataModels.IntegrationActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if dto == nil {
		return nil, diags
	}

	var filter types.Object
	if dto.Filter != nil {
		conditions := make([]attr.Value, 0, len(dto.Filter.Conditions))
		for _, c := range dto.Filter.Conditions {
			conditions = append(conditions, types.ObjectValueMust(
				dataModels.FilterConditionModelMap,
				map[string]attr.Value{
					"field":            types.StringValue(c.Field),
					"operation":        types.StringValue(c.Operation),
					"expected_value":   types.StringValue(c.ExpectedValue),
					"key":              types.StringValue(c.Key),
					"not":              types.BoolValue(c.Not),
					"order":            types.Int64Value(int64(c.Order)),
					"system_condition": types.BoolValue(c.SystemCondition),
				},
			))
		}

		filter = types.ObjectValueMust(
			dataModels.FilterModelMap,
			map[string]attr.Value{
				"conditions_empty":     types.BoolValue(dto.Filter.ConditionsEmpty),
				"condition_match_type": types.StringValue(dto.Filter.ConditionMatchType),
				"conditions":           types.ListValueMust(types.ObjectType{AttrTypes: dataModels.FilterConditionModelMap}, conditions),
			},
		)
	} else {
		filter = types.ObjectNull(dataModels.FilterModelMap)
	}

	var actionMapping types.Object
	if dto.ActionMapping != nil {
		parameterMap, _ := json.Marshal(dto.ActionMapping.Parameter)

		actionMapping = types.ObjectValueMust(
			dataModels.ActionMappingModelMap,
			map[string]attr.Value{
				"type":      types.StringValue(dto.ActionMapping.Type),
				"parameter": jsontypes.NewExactValue(string(parameterMap)),
			},
		)
	} else {
		actionMapping = types.ObjectNull(dataModels.ActionMappingModelMap)
	}

	var groupType types.String
	if dto.GroupType != "" {
		groupType = types.StringValue(dto.GroupType)
	} else {
		if model != nil && !model.GroupType.IsNull() && !model.GroupType.IsUnknown() {
			groupType = model.GroupType
		} else {
			groupType = types.StringNull() // Default group type if not specified
		}
	}

	var enabled types.Bool
	if dto.Enabled == nil {
		if (model != nil) && !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
			enabled = model.Enabled
		} else {
			enabled = types.BoolNull() // Default to true if not specified
		}
	} else {
		enabled = types.BoolPointerValue(dto.Enabled)
	}

	typeSpecificPropsMap, _ := json.Marshal(dto.TypeSpecificProperties)

	fieldMappingsMap, _ := json.Marshal(dto.FieldMappings)

	return &dataModels.IntegrationActionModel{
		ID:                     types.StringValue(dto.ID),
		IntegrationID:          integrationID,
		Type:                   types.StringValue(dto.Type),
		Name:                   types.StringValue(dto.Name),
		Domain:                 types.StringValue(dto.Domain),
		Direction:              types.StringValue(dto.Direction),
		GroupType:              groupType,
		Filter:                 filter,
		TypeSpecificProperties: jsontypes.NewExactValue(string(typeSpecificPropsMap)),
		FieldMappings:          jsontypes.NewExactValue(string(fieldMappingsMap)),
		ActionMapping:          actionMapping,
		Enabled:                enabled,
	}, diags
}

func MaintenanceModelToDto(ctx context.Context, model *dataModels.MaintenanceModel) (*dto.MaintenanceDto, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model == nil {
		return nil, diags
	}

	// Convert Rules
	var rules []dto.MaintenanceRuleDto
	if !(model.Rules.IsNull() || model.Rules.IsUnknown()) {
		var rulesObjects []dataModels.MaintenanceRuleModel
		diags.Append(model.Rules.ElementsAs(ctx, &rulesObjects, false)...)
		if diags.HasError() {
			return nil, diags
		}

		rules = make([]dto.MaintenanceRuleDto, len(rulesObjects))
		for i, ruleObj := range rulesObjects {
			var entityObj dataModels.MaintenanceRuleEntityModel
			diags.Append(ruleObj.Entity.As(ctx, &entityObj, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			rules[i] = dto.MaintenanceRuleDto{
				State: ruleObj.State.ValueString(),
				Entity: dto.MaintenanceRuleEntityDto{
					ID:   entityObj.ID.ValueString(),
					Type: entityObj.Type.ValueString(),
				},
			}
		}
	}

	return &dto.MaintenanceDto{
		ID:          model.ID.ValueString(),
		Status:      model.Status.ValueString(),
		Description: model.Description.ValueString(),
		StartDate:   model.StartDate.ValueString(),
		EndDate:     model.EndDate.ValueString(),
		TeamID:      model.TeamID.ValueString(),
		Rules:       rules,
	}, diags
}

func MaintenanceDtoToModel(ctx context.Context, dtoObj *dto.MaintenanceDto) (*dataModels.MaintenanceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if dtoObj == nil {
		return nil, diags
	}

	// Convert Rules
	var rules []attr.Value
	for _, rule := range dtoObj.Rules {
		// Create entity object
		entityObj, entityDiags := types.ObjectValue(
			dataModels.MaintenanceRuleEntityObjectType.AttrTypes,
			map[string]attr.Value{
				"id":   types.StringValue(rule.Entity.ID),
				"type": types.StringValue(rule.Entity.Type),
			},
		)
		diags.Append(entityDiags...)
		if diags.HasError() {
			return nil, diags
		}

		// Create rule object
		ruleObj, ruleDiags := types.ObjectValue(
			dataModels.MaintenanceRuleObjectType.AttrTypes,
			map[string]attr.Value{
				"state":  types.StringValue(rule.State),
				"entity": entityObj,
			},
		)
		diags.Append(ruleDiags...)
		if diags.HasError() {
			return nil, diags
		}

		rules = append(rules, ruleObj)
	}

	// Create rules list
	rulesList, rulesDiags := types.ListValue(
		dataModels.MaintenanceRuleObjectType,
		rules,
	)
	diags.Append(rulesDiags...)
	if diags.HasError() {
		return nil, diags
	}

	teamId := types.StringNull()
	if dtoObj.TeamID != "" {
		teamId = types.StringValue(dtoObj.TeamID)
	}

	return &dataModels.MaintenanceModel{
		ID:          types.StringValue(dtoObj.ID),
		Status:      types.StringValue(dtoObj.Status),
		Description: types.StringValue(dtoObj.Description),
		StartDate:   types.StringValue(dtoObj.StartDate),
		EndDate:     types.StringValue(dtoObj.EndDate),
		TeamID:      teamId,
		Rules:       rulesList,
	}, diags
}
