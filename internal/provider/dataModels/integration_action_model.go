package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationActionModel struct {
	ID                     types.String    `tfsdk:"id"`
	IntegrationID          types.String    `tfsdk:"integration_id"`
	Type                   types.String    `tfsdk:"type"`
	Name                   types.String    `tfsdk:"name"`
	Domain                 types.String    `tfsdk:"domain"`
	Direction              types.String    `tfsdk:"direction"`
	GroupType              types.String    `tfsdk:"group_type"`
	Filter                 types.Object    `tfsdk:"filter"`
	TypeSpecificProperties jsontypes.Exact `tfsdk:"type_specific_properties"`
	FieldMappings          jsontypes.Exact `tfsdk:"field_mappings"`
	ActionMapping          types.Object    `tfsdk:"action_mapping"`
	Enabled                types.Bool      `tfsdk:"enabled"`
}

type FilterModel struct {
	ConditionsEmpty    types.Bool   `tfsdk:"conditions_empty"`
	ConditionMatchType types.String `tfsdk:"condition_match_type"`
	Conditions         types.List   `tfsdk:"conditions"`
}

type FilterConditionModel struct {
	Field           types.String `tfsdk:"field"`
	Operation       types.String `tfsdk:"operation"`
	ExpectedValue   types.String `tfsdk:"expected_value"`
	Key             types.String `tfsdk:"key"`
	Not             types.Bool   `tfsdk:"not"`
	Order           types.Int64  `tfsdk:"order"`
	SystemCondition types.Bool   `tfsdk:"system_condition"`
}

type ActionMappingModel struct {
	Type      types.String    `tfsdk:"type"`
	Parameter jsontypes.Exact `tfsdk:"parameter"`
}

var FilterConditionModelMap = map[string]attr.Type{
	"field":            types.StringType,
	"operation":        types.StringType,
	"expected_value":   types.StringType,
	"key":              types.StringType,
	"not":              types.BoolType,
	"order":            types.Int64Type,
	"system_condition": types.BoolType,
}

var FilterModelMap = map[string]attr.Type{
	"conditions_empty":     types.BoolType,
	"condition_match_type": types.StringType,
	"conditions":           types.ListType{ElemType: types.ObjectType{AttrTypes: FilterConditionModelMap}},
}

var ActionMappingModelMap = map[string]attr.Type{
	"type":      types.StringType,
	"parameter": jsontypes.ExactType{},
}

var IntegrationActionModelMap = map[string]attr.Type{
	"id":                       types.StringType,
	"integration_id":           types.StringType,
	"type":                     types.StringType,
	"name":                     types.StringType,
	"domain":                   types.StringType,
	"direction":                types.StringType,
	"group_type":               types.StringType,
	"filter":                   types.ObjectType{AttrTypes: FilterModelMap},
	"type_specific_properties": jsontypes.ExactType{},
	"field_mappings":           jsontypes.ExactType{},
	"action_mapping":           types.ObjectType{AttrTypes: ActionMappingModelMap},
	"enabled":                  types.BoolType,
}

func (m *IntegrationActionModel) AsValue() types.Object {
	return types.ObjectValueMust(IntegrationActionModelMap, map[string]attr.Value{
		"id":                       m.ID,
		"integration_id":           m.IntegrationID,
		"type":                     m.Type,
		"name":                     m.Name,
		"domain":                   m.Domain,
		"direction":                m.Direction,
		"group_type":               m.GroupType,
		"filter":                   m.Filter,
		"type_specific_properties": m.TypeSpecificProperties,
		"field_mappings":           m.FieldMappings,
		"action_mapping":           m.ActionMapping,
		"enabled":                  m.Enabled,
	})
}

func (m *FilterModel) AsValue() types.Object {
	return types.ObjectValueMust(FilterModelMap, map[string]attr.Value{
		"conditions_empty":     m.ConditionsEmpty,
		"condition_match_type": m.ConditionMatchType,
		"conditions":           m.Conditions,
	})
}

func (m *FilterConditionModel) AsValue() types.Object {
	return types.ObjectValueMust(FilterConditionModelMap, map[string]attr.Value{
		"field":          m.Field,
		"operation":      m.Operation,
		"expected_value": m.ExpectedValue,
	})
}

func (m *ActionMappingModel) AsValue() types.Object {
	return types.ObjectValueMust(ActionMappingModelMap, map[string]attr.Value{
		"type":      m.Type,
		"parameter": m.Parameter,
	})
}
