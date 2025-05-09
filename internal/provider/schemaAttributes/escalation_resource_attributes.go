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
		Description: "The unique identifier of the escalation policy. This is automatically generated when the policy is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns this escalation policy. Used for access control and organization.",
		Required:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the escalation policy. This helps identify the policy's purpose and scope.",
		Required:    true,
	},
	"description": schema.StringAttribute{
		Description: "A detailed description of the escalation policy's purpose and behavior. Maximum length is 200 characters.",
		Computed:    true,
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthAtMost(200),
		},
		Default: stringdefault.StaticString(""),
	},
	"rules": schema.SetNestedAttribute{
		Description: "List of escalation rules that define how and when to escalate alerts. Each rule specifies conditions, delays, and recipients.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: EscalationRulesResponseResourceAttributes,
		},
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the escalation policy is active. When disabled, no escalations will be triggered. Defaults to true.",
		Computed:    true,
		Optional:    true,
		Default:     booldefault.StaticBool(true),
	},
	"repeat": schema.SingleNestedAttribute{
		Description: "Configuration for repeating escalations, including intervals, counts, and state management.",
		Attributes:  EscalationRepeatResourceAttributes,
		Optional:    true,
		Computed:    true,
	},
}

var EscalationRepeatResourceAttributes = map[string]schema.Attribute{
	"wait_interval": schema.Int32Attribute{
		Description: "The time to wait (in minutes) before repeating the escalation rules. Set to 0 to disable repeats. Required when configuring repeat behavior.",
		Computed:    true,
		Optional:    true,
		Validators: []validator.Int32{
			int32validator.AtLeast(0),
		},
		Default: int32default.StaticInt32(0),
	},
	"count": schema.Int32Attribute{
		Description: "The number of times to repeat the escalation rules. Must be between 1 and 20. Defaults to 1.",
		Computed:    true,
		Optional:    true,
		Validators: []validator.Int32{
			int32validator.Between(1, 20),
		},
		Default: int32default.StaticInt32(1),
	},
	"reset_recipient_states": schema.BoolAttribute{
		Description: "Whether to reset acknowledgment and seen states for recipients on each repeat cycle if the alert remains open. Defaults to false.",
		Computed:    true,
		Optional:    true,
		Default:     booldefault.StaticBool(false),
	},
	"close_alert_after_all": schema.BoolAttribute{
		Description: "Whether to automatically close the alert after all repeat cycles are completed. Defaults to false.",
		Computed:    true,
		Optional:    true,
		Default:     booldefault.StaticBool(false),
	},
}

var EscalationRulesResponseResourceAttributes = map[string]schema.Attribute{
	"condition": schema.StringAttribute{
		Description: "The condition that triggers this escalation rule. Valid values are 'if-not-acked' (escalate if alert is not acknowledged) or 'if-not-closed' (escalate if alert is not closed).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("if-not-acked", "if-not-closed"),
		},
	},
	"notify_type": schema.StringAttribute{
		Description: "How to select recipients for notification. Valid values are: 'default' (use default notification rules), 'next' (next in rotation), 'previous' (previous in rotation), 'users' (specific users), 'admins' (team admins), 'random' (random member), or 'all' (all members).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("default", "next", "previous", "users", "admins", "random", "all"),
		},
	},
	"delay": schema.Int64Attribute{
		Description: "The time to wait (in minutes) before executing this escalation rule. Must be 0 or greater.",
		Required:    true,
		Validators: []validator.Int64{
			int64validator.AtLeast(0),
		},
	},
	"recipient": schema.SingleNestedAttribute{
		Description: "The target recipient for this escalation rule. Can be a user, schedule, or team.",
		Required:    true,
		Attributes:  EscalationRuleRecipientResourceAttributes,
	},
}

var EscalationRuleRecipientResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the recipient (user ID, schedule ID, or team ID).",
		Optional:    true,
		Computed:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of recipient. Valid values are 'user' (individual user), 'schedule' (on-call schedule), or 'team' (entire team).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("user", "schedule", "team"),
		},
	},
}
