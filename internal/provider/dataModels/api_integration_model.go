package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	ApiIntegrationModel struct {
		Id                     types.String    `tfsdk:"id"`
		Name                   types.String    `tfsdk:"name"`
		ApiKey                 types.String    `tfsdk:"api_key"`
		Type                   types.String    `tfsdk:"type"`
		Enabled                types.Bool      `tfsdk:"enabled"`
		TeamId                 types.String    `tfsdk:"team_id"`
		Advanced               types.Bool      `tfsdk:"advanced"`
		MaintenanceSources     types.List      `tfsdk:"maintenance_sources"`
		Directions             types.List      `tfsdk:"directions"`
		Domains                types.List      `tfsdk:"domains"`
		TypeSpecificProperties jsontypes.Exact `tfsdk:"type_specific_properties"`
	}
)

var ApiIntegrationModelMap = map[string]attr.Type{
	"id":                       types.StringType,
	"name":                     types.StringType,
	"api_key":                  types.StringType,
	"type":                     types.StringType,
	"enabled":                  types.BoolType,
	"team_id":                  types.StringType,
	"advanced":                 types.BoolType,
	"maintenance_sources":      types.ListType{ElemType: types.ObjectType{AttrTypes: IntegrationMaintenanceSourcesResponseModelMap}},
	"directions":               types.ListType{ElemType: types.StringType},
	"domains":                  types.ListType{ElemType: types.StringType},
	"type_specific_properties": jsontypes.ExactType{},
}

func (receiver *ApiIntegrationModel) AsValue() types.Object {
	return types.ObjectValueMust(ApiIntegrationModelMap, map[string]attr.Value{
		"id":                       receiver.Id,
		"name":                     receiver.Name,
		"api_key":                  receiver.ApiKey,
		"type":                     receiver.Type,
		"enabled":                  receiver.Enabled,
		"team_id":                  receiver.TeamId,
		"advanced":                 receiver.Advanced,
		"maintenance_sources":      receiver.MaintenanceSources,
		"directions":               receiver.Directions,
		"domains":                  receiver.Domains,
		"type_specific_properties": receiver.TypeSpecificProperties,
	})
}
