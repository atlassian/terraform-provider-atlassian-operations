package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	MaintenanceIntervalModel struct {
		StartTimeMillis types.Int64 `tfsdk:"start_time_millis"`
		EndTimeMillis   types.Int64 `tfsdk:"end_time_millis"`
	}

	MaintenanceSourceModel struct {
		MaintenanceId types.String `tfsdk:"maintenance_id"`
		Enabled       types.Bool   `tfsdk:"enabled"`
		Interval      types.Object `tfsdk:"interval"`
	}
)

var IntegrationMaintenanceSourcesIntervalResponseModelMap = map[string]attr.Type{
	"start_time_millis": types.Int64Type,
	"end_time_millis":   types.Int64Type,
}

var IntegrationMaintenanceSourcesResponseModelMap = map[string]attr.Type{
	"maintenance_id": types.StringType,
	"enabled":        types.BoolType,
	"interval":       types.ObjectType{AttrTypes: IntegrationMaintenanceSourcesIntervalResponseModelMap},
}

func (receiver *MaintenanceIntervalModel) AsValue() types.Object {
	return types.ObjectValueMust(IntegrationMaintenanceSourcesIntervalResponseModelMap, map[string]attr.Value{
		"start_time_millis": receiver.StartTimeMillis,
		"end_time_millis":   receiver.EndTimeMillis,
	})
}

func (receiver *MaintenanceSourceModel) AsValue() types.Object {
	return types.ObjectValueMust(IntegrationMaintenanceSourcesResponseModelMap, map[string]attr.Value{
		"maintenance_id": receiver.MaintenanceId,
		"enabled":        receiver.Enabled,
		"interval":       receiver.Interval,
	})
}
