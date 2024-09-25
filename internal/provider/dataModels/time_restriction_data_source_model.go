package dataModels

import (
	"github.com/atlassian/jsm-ops-terraform-provider/internal/dto"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	TimeRestrictionDataSourceModel struct {
		Type         types.String                                     `tfsdk:"type"`
		Restriction  *TimeOfDayTimeRestrictionSettingsDataSourceModel `tfsdk:"restriction"`
		Restrictions []WeekdayTimeRestrictionSettingsDataSourceModel  `tfsdk:"restrictions"`
	}
	TimeOfDayTimeRestrictionSettingsDataSourceModel struct {
		StartHour types.Int32 `tfsdk:"start_hour"`
		EndHour   types.Int32 `tfsdk:"end_hour"`
		StartMin  types.Int32 `tfsdk:"start_min"`
		EndMin    types.Int32 `tfsdk:"end_min"`
	}
	WeekdayTimeRestrictionSettingsDataSourceModel struct {
		StartDay  types.String `tfsdk:"start_day"`
		EndDay    types.String `tfsdk:"end_day"`
		StartHour types.Int32  `tfsdk:"start_hour"`
		EndHour   types.Int32  `tfsdk:"end_hour"`
		StartMin  types.Int32  `tfsdk:"start_min"`
		EndMin    types.Int32  `tfsdk:"end_min"`
	}
)

func TimeRestrictionDtoToModel(dto dto.TimeRestriction) *TimeRestrictionDataSourceModel {
	model := TimeRestrictionDataSourceModel{
		Type: types.StringValue(string(dto.Type)),
	}
	if dto.WeekAndTimeOfDayRestriction != nil {
		model.Restrictions = make([]WeekdayTimeRestrictionSettingsDataSourceModel, len(*dto.WeekAndTimeOfDayRestriction))
		for i, restriction := range *dto.WeekAndTimeOfDayRestriction {
			model.Restrictions[i] = WeekdayTimeRestrictionSettingsDtoToModel(restriction)
		}
	}
	if dto.TimeOfDayRestriction != nil {
		model.Restriction = TimeOfDayTimeRestrictionSettingsDtoToModel(*dto.TimeOfDayRestriction)
	}

	return &model
}

func TimeOfDayTimeRestrictionSettingsDtoToModel(dto dto.TimeOfDayTimeRestrictionSettings) *TimeOfDayTimeRestrictionSettingsDataSourceModel {
	return &TimeOfDayTimeRestrictionSettingsDataSourceModel{
		StartHour: types.Int32Value(dto.StartHour),
		EndHour:   types.Int32Value(dto.EndHour),
		StartMin:  types.Int32Value(dto.StartMin),
		EndMin:    types.Int32Value(dto.EndMin),
	}
}

func WeekdayTimeRestrictionSettingsDtoToModel(dto dto.WeekdayTimeRestrictionSettings) WeekdayTimeRestrictionSettingsDataSourceModel {
	return WeekdayTimeRestrictionSettingsDataSourceModel{
		StartDay:  types.StringValue(string(dto.StartDay)),
		EndDay:    types.StringValue(string(dto.EndDay)),
		StartHour: types.Int32Value(dto.StartHour),
		EndHour:   types.Int32Value(dto.EndHour),
		StartMin:  types.Int32Value(dto.StartMin),
		EndMin:    types.Int32Value(dto.EndMin),
	}
}
