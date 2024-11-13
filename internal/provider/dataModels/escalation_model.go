package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	EscalationModel struct {
		Id          types.String `tfsdk:"id"`
		TeamId      types.String `tfsdk:"team_id"`
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
		Rules       types.Set    `tfsdk:"rules"`
		Enabled     types.Bool   `tfsdk:"enabled"`
		Repeat      types.Object `tfsdk:"repeat"`
	}
	EscalationRuleResponseModel struct {
		Condition  types.String `tfsdk:"condition"`
		NotifyType types.String `tfsdk:"notify_type"`
		Delay      types.Int64  `tfsdk:"delay"`
		Recipient  types.Object `tfsdk:"recipient"`
	}
	EscalationRuleResponseRecipientModel struct {
		Id   types.String `tfsdk:"id"`
		Type types.String `tfsdk:"type"`
	}
	EscalationRepeatModel struct {
		WaitInterval         types.Int32 `tfsdk:"wait_interval"`
		Count                types.Int32 `tfsdk:"count"`
		ResetRecipientStates types.Bool  `tfsdk:"reset_recipient_states"`
		CloseAlertAfterAll   types.Bool  `tfsdk:"close_alert_after_all"`
	}
)

var EscalationRuleResponseRecipientModelMap = map[string]attr.Type{
	"id":   types.StringType,
	"type": types.StringType,
}

var EscalationRuleResponseModelMap = map[string]attr.Type{
	"condition":   types.StringType,
	"notify_type": types.StringType,
	"delay":       types.Int64Type,
	"recipient":   types.ObjectType{AttrTypes: EscalationRuleResponseRecipientModelMap},
}

var EscalationRepeatModelMap = map[string]attr.Type{
	"wait_interval":          types.Int32Type,
	"count":                  types.Int32Type,
	"reset_recipient_states": types.BoolType,
	"close_alert_after_all":  types.BoolType,
}

var EscalationModelMap = map[string]attr.Type{
	"id":          types.StringType,
	"team_id":     types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"rules":       types.SetType{ElemType: types.ObjectType{AttrTypes: EscalationRuleResponseModelMap}},
	"enabled":     types.BoolType,
	"repeat":      types.ObjectType{AttrTypes: EscalationRepeatModelMap},
}

func (receiver *EscalationRuleResponseRecipientModel) AsValue() types.Object {
	return types.ObjectValueMust(EscalationRuleResponseRecipientModelMap, map[string]attr.Value{
		"id":   receiver.Id,
		"type": receiver.Type,
	})
}

func (receiver *EscalationRuleResponseModel) AsValue() types.Object {
	return types.ObjectValueMust(EscalationRuleResponseModelMap, map[string]attr.Value{
		"condition":   receiver.Condition,
		"notify_type": receiver.NotifyType,
		"delay":       receiver.Delay,
		"recipient":   receiver.Recipient,
	})
}

func (receiver *EscalationRepeatModel) AsValue() types.Object {
	return types.ObjectValueMust(EscalationRepeatModelMap, map[string]attr.Value{
		"wait_interval":          receiver.WaitInterval,
		"count":                  receiver.Count,
		"reset_recipient_states": receiver.ResetRecipientStates,
		"close_alert_after_all":  receiver.CloseAlertAfterAll,
	})
}

func (receiver *EscalationModel) AsValue() types.Object {
	return types.ObjectValueMust(EscalationModelMap, map[string]attr.Value{
		"id":          receiver.Id,
		"team_id":     receiver.TeamId,
		"name":        receiver.Name,
		"description": receiver.Description,
		"rules":       receiver.Rules,
		"enabled":     receiver.Enabled,
	})
}
