package schemaAttributes

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes/customValidators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var NotificationPolicyResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"type": schema.StringAttribute{
		Required:    true,
		Description: "The type of the notification policy. Must be 'notification'.",
		Validators: []validator.String{
			stringvalidator.OneOf("notification"),
		},
	},
	"name": schema.StringAttribute{
		Required:    true,
		Description: "The name of the notification policy",
	},
	"description": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The description of the notification policy",
	},
	"team_id": schema.StringAttribute{
		Required:    true,
		Description: "The ID of the team this notification policy belongs to",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"enabled": schema.BoolAttribute{
		Required:    true,
		Description: "Whether the notification policy is enabled",
	},
	"order": schema.Float64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Order of the notification policy",
	},
	"filter": schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The filter configuration for the notification policy",
		Validators: []validator.Object{
			customValidators.ListFieldNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("type"), "match-all"),
			customValidators.ListFieldNotNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("type"), "match-all-conditions"),
			customValidators.ListFieldNotNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("type"), "match-any-condition"),
		},
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the filter",
			},
			"conditions": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of filter conditions",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Required:    true,
							Description: "The field to filter on",
						},
						"key": schema.StringAttribute{
							Optional:    true,
							Description: "The key to filter on",
						},
						"not": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether to negate the condition",
						},
						"operation": schema.StringAttribute{
							Required:    true,
							Description: "The operation to perform",
						},
						"expected_value": schema.StringAttribute{
							Required:    true,
							Description: "The expected value for the condition",
						},
						"order": schema.Int64Attribute{
							Optional:    true,
							Description: "The order of the condition",
						},
					},
				},
			},
		},
	},
	"time_restriction": schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Time restriction configuration for the notification policy",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether time restrictions are enabled",
			},
			"time_restrictions": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of time restriction periods",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"start_hour": schema.Int64Attribute{
							Required:    true,
							Description: "Start hour of the restriction period",
						},
						"start_minute": schema.Int64Attribute{
							Required:    true,
							Description: "Start minute of the restriction period",
						},
						"end_hour": schema.Int64Attribute{
							Required:    true,
							Description: "End hour of the restriction period",
						},
						"end_minute": schema.Int64Attribute{
							Required:    true,
							Description: "End minute of the restriction period",
						},
					},
				},
			},
		},
	},
	"suppress": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
		Description: "Whether to suppress notifications for this policy",
	},
	"auto_restart_action": schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Configuration for automatically restarting alerts",
		Attributes: map[string]schema.Attribute{
			"wait_duration": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Wait Duration amount for the auto-restart action",
				Default:     int64default.StaticInt64(1),
			},
			"max_repeat_count": schema.Int64Attribute{
				Required:    true,
				Description: "Maximum number of times to repeat the restart. Must be between 1 and 20.",
				Validators: []validator.Int64{
					int64validator.AtMost(20),
				},
			},
			"duration_format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Message to include with the auto-restart action. Possible values are 'nanos', 'micros', 'millis', 'seconds', 'minutes', 'hours', 'days'",
				Validators: []validator.String{
					stringvalidator.OneOf("nanos", "micros", "millis", "seconds", "minutes", "hours", "days"),
				},
				Default: stringdefault.StaticString("minutes"),
			},
		},
	},
	"auto_close_action": schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Configuration for automatically closing alerts",
		Attributes: map[string]schema.Attribute{
			"wait_duration": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Wait Duration amount for the auto-close action",
				Default:     int64default.StaticInt64(1),
			},
			"duration_format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Message to include with the auto-close action. Possible values are 'nanos', 'micros', 'millis', 'seconds', 'minutes', 'hours', 'days'",
				Validators: []validator.String{
					stringvalidator.OneOf("nanos", "micros", "millis", "seconds", "minutes", "hours", "days"),
				},
				Default: stringdefault.StaticString("minutes"),
			},
		},
	},
	"deduplication_action": schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Configuration for alert deduplication",
		Attributes: map[string]schema.Attribute{
			"deduplication_action_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of deduplication to perform",
				Validators: []validator.String{
					stringvalidator.OneOf("valueBased", "frequencyBased"),
				},
			},
			"frequency": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Duration in seconds for the deduplication window",
			},
			"count_value_limit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of alerts to trigger deduplication",
			},
			"wait_duration": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Wait Duration amount for the deduplication_action action",
				Default:     int64default.StaticInt64(1),
			},
			"duration_format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Message to include with the deduplication_action action. Possible values are 'nanos', 'micros', 'millis', 'seconds', 'minutes', 'hours', 'days'",
				Validators: []validator.String{
					stringvalidator.OneOf("nanos", "micros", "millis", "seconds", "minutes", "hours", "days"),
				},
				Default: stringdefault.StaticString("minutes"),
			},
		},
	},
	"delay_action": schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Configuration for delaying alert notifications",
		Attributes: map[string]schema.Attribute{
			"delay_time": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Duration in seconds to delay the alert",
				Attributes: map[string]schema.Attribute{
					"hours": schema.Int64Attribute{
						Required:    true,
						Description: "Number of hours to delay the alert",
					},
					"minutes": schema.Int64Attribute{
						Required:    true,
						Description: "Number of minutes to delay the alert",
					},
				},
			},
			"delay_option": schema.StringAttribute{
				Required:    true,
				Description: "Option for how to apply the delay",
				Validators: []validator.String{
					stringvalidator.OneOf("nextTime", "nextFriday", "nextMonday", "nextWeekday",
						"nextTuesday", "nextWednesday", "nextThursday", "nextSaturday",
						"nextSunday", "forMinutes"),
				},
			},
			"wait_duration": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Wait Duration amount for the delay_action action",
				Default:     int64default.StaticInt64(1),
			},
			"duration_format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Message to include with the delay_action action. Possible values are 'nanos', 'micros', 'millis', 'seconds', 'minutes', 'hours', 'days'",
				Validators: []validator.String{
					stringvalidator.OneOf("nanos", "micros", "millis", "seconds", "minutes", "hours", "days"),
				},
				Default: stringdefault.StaticString("minutes"),
			},
		},
	},
}
