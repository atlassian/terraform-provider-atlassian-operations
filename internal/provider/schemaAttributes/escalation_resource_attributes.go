package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var EscalationResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the escalation",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns the escalation",
		Required:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the escalation",
		Required:    true,
	},
	"description": schema.StringAttribute{
		Description: "The description of the escalation",
		Computed:    true,
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthAtMost(200),
		},
		Default: stringdefault.StaticString(""),
	},
	"rules": schema.SetNestedAttribute{
		Required: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: EscalationRulesResponseResourceAttributes,
		},
		Description: "List of the escalation rules.",
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the escalation is enabled",
		Computed:    true,
		Optional:    true,
		Default:     booldefault.StaticBool(true),
	},
	"repeat": schema.SingleNestedAttribute{
		Attributes:  EscalationRepeatResourceAttributes,
		Optional:    true,
		Computed:    true,
		Description: "Repeat preferences of the escalation including repeat interval, count, reverting acknowledge and seen states back and closing an alert automatically as soon as repeats are completed.",
	},
}

var EscalationRepeatResourceAttributes = map[string]schema.Attribute{
	"wait_interval": schema.Int32Attribute{
		Computed:    true,
		Optional:    true,
		Description: "The duration in minutes to repeat the escalation rules after processing the last escalation rule. It is mandatory if you would like to add or remove repeat option. 0 should be given as a value to disable repeat option.",
		Validators: []validator.Int32{
			int32validator.AtLeast(0),
		},
		Default: int32default.StaticInt32(0),
	},
	"count": schema.Int32Attribute{
		Computed:    true,
		Optional:    true,
		Description: "Repeat time indicating how many times the repeat action will be performed.",
		Validators: []validator.Int32{
			int32validator.Between(1, 20),
		},
		Default: int32default.StaticInt32(1),
	},
	"reset_recipient_states": schema.BoolAttribute{
		Computed:    true,
		Optional:    true,
		Description: "It is for reverting acknowledge and seen states back on each repeat turn if an alert is not closed.",
		Default:     booldefault.StaticBool(false),
	},
	"close_alert_after_all": schema.BoolAttribute{
		Computed:    true,
		Optional:    true,
		Description: "It is to close the alert automatically if escalation repeats are completed.",
		Default:     booldefault.StaticBool(false),
	},
}

var EscalationRulesResponseResourceAttributes = map[string]schema.Attribute{
	"condition": schema.StringAttribute{
		Description: "The condition for notifying the recipient of escalation rule that is based on the alert state.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("if-not-acked", "if-not-closed"),
		},
	},
	"notify_type": schema.StringAttribute{
		Description: "Recipient calculation logic for escalations.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("default", "next", "previous", "users", "admins", "random", "all"),
		},
	},
	"delay": schema.Int64Attribute{
		Description: "Time delay of the escalation rule in minutes.",
		Required:    true,
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
	},
	"recipient": schema.SingleNestedAttribute{
		Required:    true,
		Attributes:  EscalationRuleRecipientResourceAttributes,
		Description: "Object of schedule, team, or users which will be notified in escalation.",
	},
}

var EscalationRuleRecipientResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the recipient",
		Optional:    true,
		Computed:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the recipient",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("user", "schedule", "team"),
		},
	},
}
