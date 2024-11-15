package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	EmailIntegrationModel struct {
		Id                          types.String `tfsdk:"id"`
		Name                        types.String `tfsdk:"name"`
		Enabled                     types.Bool   `tfsdk:"enabled"`
		TeamId                      types.String `tfsdk:"team_id"`
		Advanced                    types.Bool   `tfsdk:"advanced"`
		Directions                  types.List   `tfsdk:"directions"`
		Domains                     types.List   `tfsdk:"domains"`
		MaintenanceSources          types.List   `tfsdk:"maintenance_sources"`
		TypeSpecificPropertiesModel types.Object `tfsdk:"type_specific_properties"`
	}

	MaintenanceIntervalModel struct {
		StartTimeMillis types.Int64 `tfsdk:"start_time_millis"`
		EndTimeMillis   types.Int64 `tfsdk:"end_time_millis"`
	}

	MaintenanceSourceModel struct {
		MaintenanceId types.String `tfsdk:"maintenance_id"`
		Enabled       types.Bool   `tfsdk:"enabled"`
		Interval      types.Object `tfsdk:"interval"`
	}

	TypeSpecificPropertiesModel struct {
		EmailUsername         types.String `tfsdk:"email_username"`
		SuppressNotifications types.Bool   `tfsdk:"suppress_notifications"`
	}
)

var TypeSpecificPropertiesModelMap = map[string]attr.Type{
	"email_username":         types.StringType,
	"suppress_notifications": types.BoolType,
}

var EmailIntegrationMaintenanceSourcesIntervalResponseModelMap = map[string]attr.Type{
	"start_time_millis": types.Int64Type,
	"end_time_millis":   types.Int64Type,
}

var EmailIntegrationMaintenanceSourcesResponseModelMap = map[string]attr.Type{
	"maintenance_id": types.StringType,
	"enabled":        types.BoolType,
	"interval":       types.ObjectType{AttrTypes: EmailIntegrationMaintenanceSourcesIntervalResponseModelMap},
}

var EmailIntegrationModelMap = map[string]attr.Type{
	"id":                       types.StringType,
	"name":                     types.StringType,
	"type":                     types.StringType,
	"timezone":                 types.StringType,
	"enabled":                  types.BoolType,
	"advanced":                 types.BoolType,
	"team_id":                  types.StringType,
	"directions":               types.ListType{ElemType: types.StringType},
	"domains":                  types.ListType{ElemType: types.StringType},
	"maintenance_sources":      types.ListType{ElemType: types.ObjectType{AttrTypes: EmailIntegrationMaintenanceSourcesResponseModelMap}},
	"type_specific_properties": types.ObjectType{AttrTypes: TypeSpecificPropertiesModelMap},
}

func (receiver *MaintenanceIntervalModel) AsValue() types.Object {
	return types.ObjectValueMust(EmailIntegrationMaintenanceSourcesIntervalResponseModelMap, map[string]attr.Value{
		"start_time_millis": receiver.StartTimeMillis,
		"end_time_millis":   receiver.EndTimeMillis,
	})
}

func (receiver *MaintenanceSourceModel) AsValue() types.Object {
	return types.ObjectValueMust(EmailIntegrationMaintenanceSourcesResponseModelMap, map[string]attr.Value{
		"maintenance_id": receiver.MaintenanceId,
		"enabled":        receiver.Enabled,
	})
}

func (receiver *TypeSpecificPropertiesModel) AsValue() types.Object {
	return types.ObjectValueMust(TypeSpecificPropertiesModelMap, map[string]attr.Value{
		"email_username":         receiver.EmailUsername,
		"suppress_notifications": receiver.SuppressNotifications,
	})
}

func (receiver *EmailIntegrationModel) AsValue() types.Object {
	return types.ObjectValueMust(EmailIntegrationModelMap, map[string]attr.Value{
		"id":      receiver.Id,
		"name":    receiver.Name,
		"enabled": receiver.Enabled,
		"team_id": receiver.TeamId,
	})
}
