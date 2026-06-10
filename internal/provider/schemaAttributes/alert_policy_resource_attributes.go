package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlertPolicyResourceAttributes defines the schema attributes for the atlassian-operations_alert_policy resource.
//
// All Optional+Computed attributes carry [UseStateForUnknown] plan modifiers to keep the resource idempotent
// across applies. Without this modifier, the Terraform Plugin Framework re-marks server-derived attributes as
// (known after apply) on every plan when the user has not set them in config, producing a perpetual no-op
// update diff. Each attribute uses the plan modifier package that matches its type (stringplanmodifier,
// boolplanmodifier, int64planmodifier, listplanmodifier, mapplanmodifier, objectplanmodifier, setplanmodifier).
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
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"team_id": schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
		Description: "The ID of the team this alert policy belongs to",
	},
	"enabled": schema.BoolAttribute{
		Required:    true,
		Description: "Whether the alert policy is enabled",
	},
	"order": schema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "The order of the alert policy. Must be >= 1 (1 means first position, 2 means second, etc.). Requires explicit depends_on to ensure sequential creation.",
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"filter": schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The filter configuration for the alert policy",
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
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
		Description: "Time restriction configuration for the alert policy",
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
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
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"message": schema.StringAttribute{
		Required:    true,
		Description: "Alert message template",
	},
	"alert_description": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert description template",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"source": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert source template",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"entity": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Alert entity template",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"responders": schema.ListNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "List of responders for the alert",
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
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
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	},
	"tags": schema.SetAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Description: "Set of tags for the alert. Stored as an unordered set so that API-returned tag order changes do not produce a perpetual diff.",
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	},
	"details": schema.MapAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Description: "Additional details for the alert",
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	},
	"continue": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to continue processing after this policy",
		Default:     booldefault.StaticBool(false),
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"update_priority": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to update the priority of the alert",
		Default:     booldefault.StaticBool(false),
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"priority_value": schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "If update priorty is enabled, this is the value to set the priority to",
		Validators: []validator.String{
			stringvalidator.OneOf("P1", "P2", "P3", "P4", "P5"),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"keep_original_responders": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original responders",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"keep_original_details": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original details",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"keep_original_actions": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original actions",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"keep_original_tags": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to keep the original tags",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
}
