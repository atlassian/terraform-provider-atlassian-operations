package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AlertPolicyModel struct {
	ID                     types.String `tfsdk:"id"`
	Type                   types.String `tfsdk:"type"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	TeamID                 types.String `tfsdk:"team_id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	Order                  types.Int64  `tfsdk:"order"`
	Filter                 types.Object `tfsdk:"filter"`
	TimeRestriction        types.Object `tfsdk:"time_restriction"`
	Alias                  types.String `tfsdk:"alias"`
	Message                types.String `tfsdk:"message"`
	AlertDescription       types.String `tfsdk:"alert_description"`
	Source                 types.String `tfsdk:"source"`
	Entity                 types.String `tfsdk:"entity"`
	Responders             types.List   `tfsdk:"responders"`
	Actions                types.List   `tfsdk:"actions"`
	Tags                   types.List   `tfsdk:"tags"`
	Details                types.Map    `tfsdk:"details"`
	Continue               types.Bool   `tfsdk:"continue"`
	UpdatePriority         types.Bool   `tfsdk:"update_priority"`
	PriorityValue          types.String `tfsdk:"priority_value"`
	KeepOriginalResponders types.Bool   `tfsdk:"keep_original_responders"`
	KeepOriginalDetails    types.Bool   `tfsdk:"keep_original_details"`
	KeepOriginalActions    types.Bool   `tfsdk:"keep_original_actions"`
	KeepOriginalTags       types.Bool   `tfsdk:"keep_original_tags"`
}

type AlertConditionModel struct {
	Field         types.String `tfsdk:"field"`
	Key           types.String `tfsdk:"key"`
	Not           types.Bool   `tfsdk:"not"`
	Operation     types.String `tfsdk:"operation"`
	ExpectedValue types.String `tfsdk:"expected_value"`
	Order         types.Int64  `tfsdk:"order"`
}

type AlertFilterModel struct {
	Type       types.String `tfsdk:"type"`
	Conditions types.List   `tfsdk:"conditions"`
}

type AlertTimeRestrictionModel struct {
	Enabled          types.Bool `tfsdk:"enabled"`
	TimeRestrictions types.List `tfsdk:"time_restrictions"`
}

type AlertTimeRestrictionPeriodModel struct {
	StartHour   types.Int64 `tfsdk:"start_hour"`
	StartMinute types.Int64 `tfsdk:"start_minute"`
	EndHour     types.Int64 `tfsdk:"end_hour"`
	EndMinute   types.Int64 `tfsdk:"end_minute"`
}

type AlertResponderModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

type AlertActionModel struct {
	Type       types.String `tfsdk:"type"`
	Parameters types.Map    `tfsdk:"parameters"`
}
