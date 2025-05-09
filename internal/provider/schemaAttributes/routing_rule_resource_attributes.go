package schemaAttributes

import (
	"context"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes/customValidators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var RoutingRuleResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the routing rule. This is automatically generated when the rule is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"team_id": schema.StringAttribute{
		Description: "The unique identifier of the team that owns this routing rule. This field is required and cannot be changed after creation.",
		Required:    true,
	},
	"name": schema.StringAttribute{
		Description: "A descriptive name for the routing rule. This helps identify the rule's purpose and should be unique within the team.",
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString(""),
	},
	"order": schema.Int64Attribute{
		Description: "The order of the team routing rule within the rules. Order value is actually the index of the team routing rule whose minimum value is 0 and whose maximum value is n-1 (number of team routing rules is n).",
		Optional:    true,
		Computed:    true,
		Validators: []validator.Int64{
			int64validator.AtLeast(0),
			int64validator.AtMost(100),
		},
	},
	"is_default": schema.BoolAttribute{
		Description: "Indicates whether this is the default routing rule for the team. Default rules are used when no other rules match.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.BoolRequest, response *boolplanmodifier.RequiresReplaceIfFuncResponse) {
				if !request.StateValue.ValueBool() && request.ConfigValue.ValueBool() {
					response.RequiresReplace = true
					return
				}
			},
				"Force replacement since is_default value updated to true",
				"Force replacement since is_default value updated to true"),
		},
	},
	"timezone": schema.StringAttribute{
		Description: "The timezone used for time-based routing decisions (e.g., 'America/New_York', 'Europe/London'). Must be a valid IANA timezone identifier.",
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 50),
		},
		Computed: true,
	},
	"criteria": schema.SingleNestedAttribute{
		Description: "The conditions that determine when this routing rule should be applied to an incident.",
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
					stringvalidator.OneOf("match-all", "match-any-condition", "match-all-conditions"),
				},
			},
			"conditions": schema.ListNestedAttribute{
				Description: "List of conditions that must be met for the routing rule to be applied. Required if type is 'match-all-conditions' or 'match-any-condition'.",
				Optional:    true,
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
	"time_restriction": schema.SingleNestedAttribute{
		Description: "Time-based restrictions for when this routing rule should be active. Allows defining specific time windows and days of the week.",
		Optional:    true,
		Computed:    true,
		Attributes:  TimeRestrictionResourceAttributes,
	},
	"notify": schema.SingleNestedAttribute{
		Description: "Configuration for how incidents matching this rule should be handled.",
		Required:    true,
		Validators: []validator.Object{
			customValidators.StringFieldNotNullIfOtherField(path.MatchRelative().AtName("id"), path.MatchRelative().AtName("type"), "escalation"),
			customValidators.StringFieldNotNullIfOtherField(path.MatchRelative().AtName("id"), path.MatchRelative().AtName("type"), "schedule"),
		},
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "The type of notification to send. Valid values are: 'none' (no notification), 'escalation' (use escalation policy), 'schedule'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "escalation", "schedule"),
				},
			},
			"id": schema.StringAttribute{
				Description: "The ID of the escalation policy to use. Required when type is 'escalation' or 'schedule'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	},
}
