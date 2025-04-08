package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	notificationRuleType = "notification-rule"
)

type NotificationRuleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ActionType       types.String `tfsdk:"action_type"`
	Criteria         types.Object `tfsdk:"criteria"`
	NotificationTime types.List   `tfsdk:"notification_time"`
	TimeRestriction  types.Object `tfsdk:"time_restriction"`
	Schedules        types.List   `tfsdk:"schedules"`
	Order            types.Int64  `tfsdk:"order"`
	Steps            types.List   `tfsdk:"steps"`
	Repeat           types.Object `tfsdk:"repeat"`
	Enabled          types.Bool   `tfsdk:"enabled"`
}

func (m NotificationRuleModel) GetType() string {
	return notificationRuleType
}

func (m NotificationRuleModel) AsValue() types.Object {
	return types.ObjectValueMust(NotificationRuleModelMap, map[string]attr.Value{
		"id":                m.ID,
		"name":              m.Name,
		"action_type":       m.ActionType,
		"criteria":          m.Criteria,
		"notification_time": m.NotificationTime,
		"time_restriction":  m.TimeRestriction,
		"schedules":         m.Schedules,
		"order":             m.Order,
		"steps":             m.Steps,
		"repeat":            m.Repeat,
		"enabled":           m.Enabled,
	})
}

type NotificationRuleConditionModel struct {
	Field         types.String `tfsdk:"field"`
	Operation     types.String `tfsdk:"operation"`
	ExpectedValue types.String `tfsdk:"expected_value"`
	Key           types.String `tfsdk:"key"`
	Not           types.Bool   `tfsdk:"not"`
	Order         types.Int64  `tfsdk:"order"`
}

func (c NotificationRuleConditionModel) AsValue() types.Object {
	return types.ObjectValueMust(NotificationRuleConditionModelMap, map[string]attr.Value{
		"field":          c.Field,
		"operation":      c.Operation,
		"expected_value": c.ExpectedValue,
		"key":            c.Key,
		"not":            c.Not,
		"order":          c.Order,
	})
}

type NotificationRuleStepModel struct {
	SendAfter types.Int64  `tfsdk:"send_after"`
	Contact   types.Object `tfsdk:"contact"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

func (s NotificationRuleStepModel) AsValue() types.Object {
	return types.ObjectValueMust(NotificationRuleStepModelMap, map[string]attr.Value{
		"send_after": s.SendAfter,
		"contact":    s.Contact,
		"enabled":    s.Enabled,
	})
}

type NotificationContactModel struct {
	Method types.String `tfsdk:"method"`
	To     types.String `tfsdk:"to"`
}

func (c NotificationContactModel) AsValue() types.Object {
	return types.ObjectValueMust(NotificationContactModelMap, map[string]attr.Value{
		"method": c.Method,
		"to":     c.To,
	})
}

type NotificationRuleRepeatModel struct {
	LoopAfter types.Int64 `tfsdk:"loop_after"`
	Enabled   types.Bool  `tfsdk:"enabled"`
}

func (r NotificationRuleRepeatModel) AsValue() types.Object {
	return types.ObjectValueMust(NotificationRuleRepeatModelMap, map[string]attr.Value{
		"loop_after": r.LoopAfter,
		"enabled":    r.Enabled,
	})
}

var NotificationRuleModelMap = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"action_type": types.StringType,
	"criteria": types.ObjectType{
		AttrTypes: CriteriaModelMap,
	},
	"notification_time": types.ListType{
		ElemType: types.StringType,
	},
	"time_restriction": types.ObjectType{
		AttrTypes: TimeRestrictionModelMap,
	},
	"schedules": types.ListType{
		ElemType: types.StringType,
	},
	"order": types.Int64Type,
	"steps": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: NotificationRuleStepModelMap,
		},
	},
	"repeat": types.ObjectType{
		AttrTypes: NotificationRuleRepeatModelMap,
	},
	"enabled": types.BoolType,
}

var NotificationRuleConditionModelMap = map[string]attr.Type{
	"field":          types.StringType,
	"operation":      types.StringType,
	"expected_value": types.StringType,
	"key":            types.StringType,
	"not":            types.BoolType,
	"order":          types.Int64Type,
}

var NotificationRuleStepModelMap = map[string]attr.Type{
	"send_after": types.Int64Type,
	"contact": types.ObjectType{
		AttrTypes: NotificationContactModelMap,
	},
	"enabled": types.BoolType,
}

var NotificationContactModelMap = map[string]attr.Type{
	"method": types.StringType,
	"to":     types.StringType,
}

var NotificationRuleRepeatModelMap = map[string]attr.Type{
	"loop_after": types.Int64Type,
	"enabled":    types.BoolType,
}
