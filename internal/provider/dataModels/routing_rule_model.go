package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoutingRuleModel struct {
	ID              types.String `tfsdk:"id"`
	TeamID          types.String `tfsdk:"team_id"`
	Name            types.String `tfsdk:"name"`
	Order           types.Int64  `tfsdk:"order"`
	IsDefault       types.Bool   `tfsdk:"is_default"`
	Timezone        types.String `tfsdk:"timezone"`
	Criteria        types.Object `tfsdk:"criteria"`
	TimeRestriction types.Object `tfsdk:"time_restriction"`
	Notify          types.Object `tfsdk:"notify"`
}

type RoutingRuleNotifyModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

var RoutingRuleModelMap = map[string]attr.Type{
	"id":         types.StringType,
	"team_id":    types.StringType,
	"name":       types.StringType,
	"order":      types.Int64Type,
	"is_default": types.BoolType,
	"timezone":   types.StringType,
	"criteria": types.ObjectType{
		AttrTypes: CriteriaModelMap,
	},
	"time_restriction": types.ObjectType{
		AttrTypes: TimeRestrictionModelMap,
	},
	"notify": types.ObjectType{
		AttrTypes: RoutingRuleNotifyModelMap,
	},
}

var RoutingRuleNotifyModelMap = map[string]attr.Type{
	"type": types.StringType,
	"id":   types.StringType,
}

func (receiver *RoutingRuleNotifyModel) AsValue() types.Object {
	return types.ObjectValueMust(RoutingRuleNotifyModelMap, map[string]attr.Value{
		"type": receiver.Type,
		"id":   receiver.ID,
	})
}

func (receiver *RoutingRuleModel) AsValue() types.Object {
	return types.ObjectValueMust(RoutingRuleModelMap, map[string]attr.Value{
		"id":               receiver.ID,
		"team_id":          receiver.TeamID,
		"name":             receiver.Name,
		"order":            receiver.Order,
		"is_default":       receiver.IsDefault,
		"timezone":         receiver.Timezone,
		"criteria":         receiver.Criteria,
		"time_restriction": receiver.TimeRestriction,
		"notify":           receiver.Notify,
	})
}
