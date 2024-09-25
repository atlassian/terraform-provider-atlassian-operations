package dataModels

import (
	"github.com/atlassian/terraform-provider-jsm-ops/internal/dto"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ScheduleDataSourceModel struct {
	Id          types.String              `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	Description types.String              `tfsdk:"description"`
	Timezone    types.String              `tfsdk:"timezone"`
	Enabled     types.Bool                `tfsdk:"enabled"`
	TeamId      types.String              `tfsdk:"team_id"`
	Rotations   []RotationDataSourceModel `tfsdk:"rotations"`
}

func ScheduleDtoToModel(dto dto.Schedule) ScheduleDataSourceModel {
	model := ScheduleDataSourceModel{
		Id:          types.StringValue(dto.Id),
		Name:        types.StringValue(dto.Name),
		Description: types.StringValue(dto.Description),
		Timezone:    types.StringValue(dto.Timezone),
		Enabled:     types.BoolValue(dto.Enabled),
		TeamId:      types.StringValue(dto.TeamId),
		Rotations:   make([]RotationDataSourceModel, len(dto.Rotations)),
	}
	for i, rotation := range dto.Rotations {
		model.Rotations[i] = RotationDtoToModel(rotation)
	}

	return model
}
