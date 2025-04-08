package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationPolicyModel struct {
	ID                  types.String  `tfsdk:"id"`
	Type                types.String  `tfsdk:"type"`
	Name                types.String  `tfsdk:"name"`
	Description         types.String  `tfsdk:"description"`
	TeamID              types.String  `tfsdk:"team_id"`
	Enabled             types.Bool    `tfsdk:"enabled"`
	Order               types.Float64 `tfsdk:"order"`
	Filter              types.Object  `tfsdk:"filter"`
	TimeRestriction     types.Object  `tfsdk:"time_restriction"`
	AutoRestartAction   types.Object  `tfsdk:"auto_restart_action"`
	AutoCloseAction     types.Object  `tfsdk:"auto_close_action"`
	DeduplicationAction types.Object  `tfsdk:"deduplication_action"`
	DelayAction         types.Object  `tfsdk:"delay_action"`
	Suppress            types.Bool    `tfsdk:"suppress"`
}

type NotificationPolicyTimeRestrictionModel struct {
	Enabled          types.Bool `tfsdk:"enabled"`
	TimeRestrictions types.List `tfsdk:"time_restrictions"`
}

type NotificationPolicyTimeRestrictionSettingsModel struct {
	StartHour   types.Int64 `tfsdk:"start_hour"`
	EndHour     types.Int64 `tfsdk:"end_hour"`
	StartMinute types.Int64 `tfsdk:"start_minute"`
	EndMinute   types.Int64 `tfsdk:"end_minute"`
}

type NotificationConditionModel struct {
	Field         types.String `tfsdk:"field"`
	Key           types.String `tfsdk:"key"`
	Not           types.Bool   `tfsdk:"not"`
	Operation     types.String `tfsdk:"operation"`
	ExpectedValue types.String `tfsdk:"expected_value"`
	Order         types.Int64  `tfsdk:"order"`
}

type NotificationFilterModel struct {
	Type       types.String `tfsdk:"type"`
	Conditions types.List   `tfsdk:"conditions"`
}

type AutoRestartActionModel struct {
	WaitDuration   types.Int64  `tfsdk:"wait_duration"`
	MaxRepeatCount types.Int64  `tfsdk:"max_repeat_count"`
	DurationFormat types.String `tfsdk:"duration_format"`
}

type AutoCloseActionModel struct {
	WaitDuration   types.Int64  `tfsdk:"wait_duration"`
	DurationFormat types.String `tfsdk:"duration_format"`
}

type DeduplicationActionModel struct {
	DeduplicationActionType types.String `tfsdk:"deduplication_action_type"`
	Frequency               types.Int64  `tfsdk:"frequency"`
	CountValueLimit         types.Int64  `tfsdk:"count_value_limit"`
	WaitDuration            types.Int64  `tfsdk:"wait_duration"`
	DurationFormat          types.String `tfsdk:"duration_format"`
}

type DelayActionModel struct {
	DelayTime      types.Object `tfsdk:"delay_time"`
	DelayOption    types.String `tfsdk:"delay_option"`
	WaitDuration   types.Int64  `tfsdk:"wait_duration"`
	DurationFormat types.String `tfsdk:"duration_format"`
}

type DelayActionDelayTimeModel struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}
