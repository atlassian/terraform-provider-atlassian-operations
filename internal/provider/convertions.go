package provider

import (
	"context"
	"encoding/json"

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

func ApiIntegrationDtoToModel(dtoObj dto.ApiIntegration) dataModels.ApiIntegrationModel {
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

func NotificationRuleModelToDto(ctx context.Context, model dataModels.NotificationRuleModel) dto.NotificationRuleDto {
	var notificationTimes []string
	if !(model.NotificationTime.IsNull() || model.NotificationTime.IsUnknown()) {
		var times []string
		diags := model.NotificationTime.ElementsAs(ctx, &times, false)
		if diags.HasError() {
			return dto.NotificationRuleDto{}
		}
		notificationTimes = times
	}

	var schedules []string
	if !(model.Schedules.IsNull() || model.Schedules.IsUnknown()) {
		var scheds []string
		diags := model.Schedules.ElementsAs(ctx, &scheds, false)
		if diags.HasError() {
			return dto.NotificationRuleDto{}
		}
		schedules = scheds
	}

	var timeRestriction *dto.TimeRestriction
	if !(model.TimeRestriction.IsNull() || model.TimeRestriction.IsUnknown()) {
		var timeRestrictionModel dataModels.TimeRestrictionModel
		diags := model.TimeRestriction.As(ctx, &timeRestrictionModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return dto.NotificationRuleDto{}
		}
		timeRestriction = TimeRestrictionModelToDto(ctx, timeRestrictionModel)
	}

	var steps []dto.NotificationRuleStep
	if !(model.Steps.IsNull() || model.Steps.IsUnknown()) {
		var stepsList []dataModels.NotificationRuleStepModel
		diags := model.Steps.ElementsAs(ctx, &stepsList, false)
		if diags.HasError() {
			return dto.NotificationRuleDto{}
		}

		for _, step := range stepsList {
			var contact dataModels.NotificationContactModel
			diags := step.Contact.As(ctx, &contact, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return dto.NotificationRuleDto{}
			}

			steps = append(steps, dto.NotificationRuleStep{
				Contact: dto.NotificationContact{
					Method: contact.Method.ValueString(),
					To:     contact.To.ValueString(),
				},
				SendAfter: int(step.SendAfter.ValueInt64()),
				Enabled:   step.Enabled.ValueBool(),
			})
		}
	}

	var repeat *dto.NotificationRuleRepeat
	if !(model.Repeat.IsNull() || model.Repeat.IsUnknown()) {
		var repeatModel dataModels.NotificationRuleRepeatModel
		diags := model.Repeat.As(ctx, &repeatModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return dto.NotificationRuleDto{}
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
	}
}

func NotificationRuleDtoToModel(ctx context.Context, dto dto.NotificationRuleDto) dataModels.NotificationRuleModel {
	var notificationTime types.List
	if dto.NotificationTime != nil {
		elements := make([]attr.Value, len(dto.NotificationTime))
		for i, time := range dto.NotificationTime {
			elements[i] = types.StringValue(time)
		}
		notificationTime = types.ListValueMust(types.StringType, elements)
	} else {
		notificationTime = types.ListNull(types.StringType)
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
					"send_after": types.Int64Value(int64(step.SendAfter)),
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
