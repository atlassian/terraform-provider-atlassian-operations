package schemaAttributes

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes/customValidators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var NotificationRuleResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the notification rule.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the notification rule. This field is required and must be unique within the team.",
		Required:    true,
	},
	"action_type": schema.StringAttribute{
		Description: "The type of action that triggers this notification rule. Valid values are: create-alert, acknowledged-alert, closed-alert, assigned-alert, add-note, schedule-start, schedule-end, incoming-call-routing.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("create-alert", "acknowledged-alert", "closed-alert", "assigned-alert", "add-note", "schedule-start", "schedule-end", "incoming-call-routing"),
		},
	},
	"criteria": schema.SingleNestedAttribute{
		Description: "The criteria that determines when this notification rule should be triggered. Currently only supports 'match-all' type.",
		Optional:    true,
		Computed:    true,
		Validators: []validator.Object{
			customValidators.ListFieldNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("type"), "match-all"),
			customValidators.ListFieldNotNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("type"), "match-all-conditions"),
			customValidators.ListFieldNotNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("type"), "match-any-condition"),
		},
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "The type of criteria matching to use. Valid values are: 'match-all' (matches all incidents), 'match-all-conditions' (all conditions must match), or 'match-any-condition' (any condition can match).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("match-all", "match-all-conditions", "match-any-condition"),
				},
			},
			"conditions": schema.ListNestedAttribute{
				Description: "List of conditions that must be met for the routing rule to be applied. Required if type is 'match-all-conditions' or 'match-any-condition'.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "The incident field to evaluate (e.g., 'message', 'priority', 'tags').",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("message", "alias", "description", "source", "entity", "tags", "actions", "extra-properties", "priority", "details", "responders"),
							},
						},
						"operation": schema.StringAttribute{
							Description: "The comparison operation to perform (e.g., 'equals', 'contains', 'matches').",
							Required:    true,
						},
						"expected_value": schema.StringAttribute{
							Description: "The value to compare against the field value.",
							Optional:    true,
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "If field is set as extra-properties, key could be used for key-value pair.",
							Optional:    true,
							Computed:    true,
						},
						"not": schema.BoolAttribute{
							Description: "Indicates behaviour of the given operation.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"order": schema.Int64Attribute{
							Description: "Order of the condition in conditions list.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	},
	"notification_time": schema.ListAttribute{
		Description: "List of times when notifications should be sent. Valid values include: just-before, 15-minutes-ago, 1-hour-ago, 1-day-ago.",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
	"time_restriction": schema.SingleNestedAttribute{
		Description: "Time restrictions for when this notification rule should be active. Allows setting specific days of the week and time ranges.",
		Optional:    true,
		Computed:    true,
		Attributes:  TimeRestrictionResourceAttributes,
	},
	"schedules": schema.ListAttribute{
		Description: "List of schedule IDs that this notification rule applies to.",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
	"order": schema.Int64Attribute{
		Description: "The order in which this notification rule should be processed relative to other rules. Lower numbers are processed first.",
		Optional:    true,
		Computed:    true,
	},
	"steps": schema.ListNestedAttribute{
		Description: "List of notification steps that define who should be notified and when.",
		Optional:    true,
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"send_after": schema.Int64Attribute{
					Description: "The number of minutes to wait before sending this notification after the rule is triggered.",
					Required:    true,
				},
				"contact": schema.SingleNestedAttribute{
					Description: "The contact information for this notification step.",
					Required:    true,
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Description: "The method of contact (e.g., email, sms, voice, mobile).",
							Validators: []validator.String{
								stringvalidator.OneOf("email", "sms", "voice", "mobile"),
							},
							Required: true,
						},
						"to": schema.StringAttribute{
							Description: "The recipient of the notification (e.g., email address, phone number).",
							Required:    true,
						},
					},
				},
				"enabled": schema.BoolAttribute{
					Description: "Whether this notification step is enabled.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
	},
	"repeat": schema.SingleNestedAttribute{
		Description: "Configuration for repeating notifications.",
		Optional:    true,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"loop_after": schema.Int64Attribute{
				Description: "The number of minutes to wait before repeating the notification steps.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether notification repetition is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether this notification rule is enabled.",
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(true),
	},
}
