package dataModels

import (
	"github.com/atlassian/terraform-provider-jsm-ops/internal/dto"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	RotationDataSourceModel struct {
		Id              types.String                    `tfsdk:"id"`
		Name            types.String                    `tfsdk:"name"`
		StartDate       types.String                    `tfsdk:"start_date"`
		EndDate         types.String                    `tfsdk:"end_date"`
		Type            types.String                    `tfsdk:"type"`
		Length          types.Int32                     `tfsdk:"length"`
		Participants    []ResponderInfoDataSourceModel  `tfsdk:"participants"`
		TimeRestriction *TimeRestrictionDataSourceModel `tfsdk:"time_restriction"`
	}
	ResponderInfoDataSourceModel struct {
		Id   types.String `tfsdk:"id"`
		Type types.String `tfsdk:"type"`
	}
)

func ResponderInfoDtoToModel(dto dto.ResponderInfo) ResponderInfoDataSourceModel {
	model := ResponderInfoDataSourceModel{
		Id:   types.StringValue(""),
		Type: types.StringValue(string(dto.Type)),
	}
	if dto.Id != nil {
		model.Id = types.StringValue(*dto.Id)
	}
	return model
}

func RotationDtoToModel(dto dto.Rotation) RotationDataSourceModel {
	model := RotationDataSourceModel{
		Id:           types.StringValue(dto.Id),
		Name:         types.StringValue(dto.Name),
		StartDate:    types.StringValue(dto.StartDate),
		EndDate:      types.StringValue(dto.EndDate),
		Type:         types.StringValue(string(dto.Type)),
		Length:       types.Int32Value(dto.Length),
		Participants: make([]ResponderInfoDataSourceModel, len(dto.Participants)),
	}
	if dto.TimeRestriction != nil {
		model.TimeRestriction = TimeRestrictionDtoToModel(*dto.TimeRestriction)
	}
	for i, participant := range dto.Participants {
		model.Participants[i] = ResponderInfoDtoToModel(participant)
	}

	return model
}
