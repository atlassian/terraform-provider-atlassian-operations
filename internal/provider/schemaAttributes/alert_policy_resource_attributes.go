package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var AlertPolicyResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"type": schema.StringAttribute{
		Required:    true,
		Description: "The type of the alert policy. Must be 'alert'.",
		Validators: []validator.String{
			stringvalidator.OneOf("alert"),
		},
	},
	"name": schema.StringAttribute{
		Required:    true,
		Description: "The name of the alert policy",
	},
	"description": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The description of the alert policy",
	},
	"team_id": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The ID of the team this alert policy belongs to",
	},
	"enabled": schema.BoolAttribute{
		Required:    true,
		Description: "Whether the alert policy is enabled",
	},
	"order": schema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "The order of the alert policy",
	},
	"filter": schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The filter configuration for the alert policy",
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
		Description: "Time restriction configuration for the alert policy",
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
	"alias": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert alias template",
	},
	"message": schema.StringAttribute{
		Required:    true,
		Description: "Alert message template",
	},
	"alert_description": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert description template",
	},
	"source": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert source template",
	},
	"entity": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert entity template",
	},
	"responders": schema.ListNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "List of responders for the alert",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Required:    true,
					Description: "The type of the responder",
				},
				"id": schema.StringAttribute{
					Optional:    true,
					Description: "The ID of the responder",
				},
			},
		},
	},
	"actions": schema.ListAttribute{
		Optional:    true,
		Computed:    true,
		Description: "List of actions for the alert",
		ElementType: types.StringType,
	},
	"tags": schema.ListAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Description: "List of tags for the alert",
	},
	"details": schema.MapAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Description: "Additional details for the alert",
	},
	"continue": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to continue processing after this policy",
		Default:     booldefault.StaticBool(false),
	},
	"update_priority": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to update the priority of the alert",
		Default:     booldefault.StaticBool(false),
	},
	"priority_value": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "If update priorty is enabled, this is the value to set the priority to",
		Validators: []validator.String{
			stringvalidator.OneOf("P1", "P2", "P3", "P4", "P5"),
		},
	},
	"keep_original_responders": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original responders",
	},
	"keep_original_details": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original details",
	},
	"keep_original_actions": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original actions",
	},
	"keep_original_tags": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original tags",
	},
}
