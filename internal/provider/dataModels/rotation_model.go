package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	RotationModel struct {
		Id              types.String `tfsdk:"id"`
		ScheduleId      types.String `tfsdk:"schedule_id"`
		Name            types.String `tfsdk:"name"`
		StartDate       types.String `tfsdk:"start_date"`
		EndDate         types.String `tfsdk:"end_date"`
		Type            types.String `tfsdk:"type"`
		Length          types.Int32  `tfsdk:"length"`
		Participants    types.List   `tfsdk:"participants"`
		TimeRestriction types.Object `tfsdk:"time_restriction"`
	}
	ResponderInfoModel struct {
		Id   types.String `tfsdk:"id"`
		Type types.String `tfsdk:"type"`
	}
)

var RotationModelMap = map[string]attr.Type{
	"id":          types.StringType,
	"schedule_id": types.StringType,
	"name":        types.StringType,
	"start_date":  types.StringType,
	"end_date":    types.StringType,
	"type":        types.StringType,
	"length":      types.Int32Type,
	"participants": types.ListType{ElemType: types.ObjectType{
		AttrTypes: ResponderInfoModelMap,
	}},
	"time_restriction": types.ObjectType{
		AttrTypes: TimeRestrictionModelMap,
	},
}

var ResponderInfoModelMap = map[string]attr.Type{
	"id":   types.StringType,
	"type": types.StringType,
}

func (receiver *ResponderInfoModel) AsValue() types.Object {
	return types.ObjectValueMust(ResponderInfoModelMap, map[string]attr.Value{
		"id":   receiver.Id,
		"type": receiver.Type,
	})
}

func (receiver *RotationModel) AsValue() types.Object {
	return types.ObjectValueMust(RotationModelMap, map[string]attr.Value{
		"id":               receiver.Id,
		"schedule_id":      receiver.ScheduleId,
		"name":             receiver.Name,
		"start_date":       receiver.StartDate,
		"end_date":         receiver.EndDate,
		"type":             receiver.Type,
		"length":           receiver.Length,
		"participants":     receiver.Participants,
		"time_restriction": receiver.TimeRestriction,
	})
}
