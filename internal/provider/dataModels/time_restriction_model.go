package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	TimeRestrictionModel struct {
		Type         types.String `tfsdk:"type"`
		Restriction  types.Object `tfsdk:"restriction"`
		Restrictions types.List   `tfsdk:"restrictions"`
	}
	TimeOfDayTimeRestrictionSettingsModel struct {
		StartHour types.Int32 `tfsdk:"start_hour"`
		EndHour   types.Int32 `tfsdk:"end_hour"`
		StartMin  types.Int32 `tfsdk:"start_min"`
		EndMin    types.Int32 `tfsdk:"end_min"`
	}
	WeekdayTimeRestrictionSettingsModel struct {
		StartDay  types.String `tfsdk:"start_day"`
		EndDay    types.String `tfsdk:"end_day"`
		StartHour types.Int32  `tfsdk:"start_hour"`
		EndHour   types.Int32  `tfsdk:"end_hour"`
		StartMin  types.Int32  `tfsdk:"start_min"`
		EndMin    types.Int32  `tfsdk:"end_min"`
	}
)

var TimeRestrictionModelMap = map[string]attr.Type{
	"type": types.StringType,
	"restriction": types.ObjectType{
		AttrTypes: TimeOfDayTimeRestrictionSettingsModelMap,
	},
	"restrictions": types.ListType{ElemType: types.ObjectType{
		AttrTypes: WeekdayTimeRestrictionSettingsModelMap,
	}},
}

var WeekdayTimeRestrictionSettingsModelMap = map[string]attr.Type{
	"start_day":  types.StringType,
	"end_day":    types.StringType,
	"start_hour": types.Int32Type,
	"end_hour":   types.Int32Type,
	"start_min":  types.Int32Type,
	"end_min":    types.Int32Type,
}

var TimeOfDayTimeRestrictionSettingsModelMap = map[string]attr.Type{
	"start_hour": types.Int32Type,
	"end_hour":   types.Int32Type,
	"start_min":  types.Int32Type,
	"end_min":    types.Int32Type,
}

func (receiver *TimeRestrictionModel) AsValue() types.Object {
	return types.ObjectValueMust(TimeRestrictionModelMap, map[string]attr.Value{
		"type":         receiver.Type,
		"restriction":  receiver.Restriction,
		"restrictions": receiver.Restrictions,
	})
}

func (receiver *TimeOfDayTimeRestrictionSettingsModel) AsValue() types.Object {
	return types.ObjectValueMust(TimeOfDayTimeRestrictionSettingsModelMap, map[string]attr.Value{
		"start_hour": receiver.StartHour,
		"end_hour":   receiver.EndHour,
		"start_min":  receiver.StartMin,
		"end_min":    receiver.EndMin,
	})
}

func (receiver *WeekdayTimeRestrictionSettingsModel) AsValue() types.Object {
	return types.ObjectValueMust(WeekdayTimeRestrictionSettingsModelMap, map[string]attr.Value{
		"start_day":  receiver.StartDay,
		"end_day":    receiver.EndDay,
		"start_hour": receiver.StartHour,
		"end_hour":   receiver.EndHour,
		"start_min":  receiver.StartMin,
		"end_min":    receiver.EndMin,
	})
}
