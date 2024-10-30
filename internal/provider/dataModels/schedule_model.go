package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ScheduleModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Timezone    types.String `tfsdk:"timezone"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	TeamId      types.String `tfsdk:"team_id"`
}

var ScheduleModelMap = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"timezone":    types.StringType,
	"enabled":     types.BoolType,
	"team_id":     types.StringType,
}

func (receiver *ScheduleModel) AsValue() types.Object {
	return types.ObjectValueMust(ScheduleModelMap, map[string]attr.Value{
		"id":          receiver.Id,
		"name":        receiver.Name,
		"description": receiver.Description,
		"timezone":    receiver.Timezone,
		"enabled":     receiver.Enabled,
		"team_id":     receiver.TeamId,
	})
}
