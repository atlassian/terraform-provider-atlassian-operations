package schemaAttributes

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes/customValidators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var IntegrationActionResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the integration action",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"integration_id": schema.StringAttribute{
		Description: "The ID of the integration that integration action is going to be",
		Required:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the integration action",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 250),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the integration action",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 250),
		},
	},
	"domain": schema.StringAttribute{
		Description: "The domain of the integration action",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("alert"),
		},
	},
	"direction": schema.StringAttribute{
		Description: "The direction of the integration action (incoming or outgoing)",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("incoming", "outgoing"),
		},
	},
	"group_type": schema.StringAttribute{
		Description: "The group type of the integration action",
		Optional:    true,
		Computed:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("forwarding", "updating", "checkbox"),
		},
	},
	"filter": schema.SingleNestedAttribute{
		Description: "The filter configuration for the integration action",
		Optional:    true,
		Computed:    true,
		Validators: []validator.Object{
			customValidators.ListFieldNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("condition_match_type"), "match-all"),
			customValidators.ListFieldNotNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("condition_match_type"), "match-all-conditions"),
			customValidators.ListFieldNotNullIfOtherField(path.MatchRelative().AtName("conditions"), path.MatchRelative().AtName("condition_match_type"), "match-any-condition"),
		},
		Attributes: map[string]schema.Attribute{
			"conditions_empty": schema.BoolAttribute{
				Description: "Whether the conditions list is empty",
				Required:    true,
			},
			"condition_match_type": schema.StringAttribute{
				Description: "The type of condition matching to apply",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("match-all", "match-any-condition", "match-all-conditions"),
				},
			},
			"conditions": schema.ListNestedAttribute{
				Description: "List of conditions for the filter",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "The incident field to evaluate (e.g., 'message', 'priority', 'tags').",
							Required:    true,
						},
						"operation": schema.StringAttribute{
							Description: "The comparison operation to perform (e.g., 'matches', 'contains', 'starts-with', 'ends-with', 'equals', 'contains-key', 'contains-value', 'greater-than', 'less-than', 'is-empty', 'equals-ignore-whitespace').",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("matches", "contains", "starts-with", "ends-with", "equals", "contains-key", "contains-value", "greater-than", "less-than", "is-empty", "equals-ignore-whitespace"),
							},
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
						"system_condition": schema.BoolAttribute{
							Description: "Whether the condition is a system condition",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	},
	"type_specific_properties": schema.StringAttribute{
		CustomType:  jsontypes.ExactType{},
		Description: "Type-specific properties for the integration action",
		Optional:    true,
		Computed:    true,
	},
	"field_mappings": schema.StringAttribute{
		CustomType:  jsontypes.ExactType{},
		Description: "Field mappings for the integration action",
		Optional:    true,
		Computed:    true,
	},
	"action_mapping": schema.SingleNestedAttribute{
		Description: "The action mapping configuration",
		Optional:    true,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "The type of action mapping",
				Required:    true,
			},
			"parameter": schema.StringAttribute{
				CustomType:  jsontypes.ExactType{},
				Description: "Parameters for the action mapping",
				Optional:    true,
				Computed:    true,
			},
		},
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the integration action is enabled",
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(true),
	},
}
