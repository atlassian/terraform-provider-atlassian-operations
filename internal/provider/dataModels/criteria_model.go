package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	CriteriaModel struct {
		Type       types.String `tfsdk:"type"`
		Conditions types.List   `tfsdk:"conditions"`
	}

	CriteriaConditionModel struct {
		Field         types.String `tfsdk:"field"`
		Operation     types.String `tfsdk:"operation"`
		ExpectedValue types.String `tfsdk:"expected_value"`
		Key           types.String `tfsdk:"key"`
		Not           types.Bool   `tfsdk:"not"`
		Order         types.Int64  `tfsdk:"order"`
	}
)

var ConditionModelMap = map[string]attr.Type{
	"field":          types.StringType,
	"operation":      types.StringType,
	"expected_value": types.StringType,
	"key":            types.StringType,
	"not":            types.BoolType,
	"order":          types.Int64Type,
}

var CriteriaModelMap = map[string]attr.Type{
	"type": types.StringType,
	"conditions": types.ListType{ElemType: types.ObjectType{
		AttrTypes: ConditionModelMap,
	}},
}

func (receiver *CriteriaConditionModel) AsValue() types.Object {
	return types.ObjectValueMust(ConditionModelMap, map[string]attr.Value{
		"field":          receiver.Field,
		"operation":      receiver.Operation,
		"expected_value": receiver.ExpectedValue,
		"key":            receiver.Key,
		"not":            receiver.Not,
		"order":          receiver.Order,
	})
}

func (receiver *CriteriaModel) AsValue() types.Object {
	return types.ObjectValueMust(CriteriaModelMap, map[string]attr.Value{
		"type":       receiver.Type,
		"conditions": receiver.Conditions,
	})
}
