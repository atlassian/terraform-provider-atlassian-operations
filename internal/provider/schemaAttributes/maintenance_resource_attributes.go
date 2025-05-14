package schemaAttributes

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// MaintenanceResourceAttributes defines the schema for the Maintenance resource
var MaintenanceResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The unique identifier of the maintenance window",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"description": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "The description of the maintenance window",
	},
	"start_date": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The start date/time of the maintenance window in ISO8601 format (e.g., 2023-06-15T10:00:00Z)",
	},
	"end_date": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The end date/time of the maintenance window in ISO8601 format (e.g., 2023-06-15T14:00:00Z)",
	},
	"status": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The status of the maintenance window (e.g., scheduled, in_progress, completed, cancelled)",
	},
	"team_id": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "The ID of the team associated with this maintenance window",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
				if request.StateValue.ValueString() != request.ConfigValue.ValueString() {
					response.RequiresReplace = true
					return
				}
			},
				"Force replacement since method value updated",
				"Force replacement since method value updated"),
		},
	},
	"rules": schema.ListNestedAttribute{
		Required:            true,
		MarkdownDescription: "A list of rules defining what entities are affected during the maintenance window",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"state": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "The state to apply to the entity during maintenance (e.g., disabled, enabled, noMaintenance)",
					Validators: []validator.String{
						stringvalidator.OneOf("disabled", "enabled", "noMaintenance"),
					},
				},
				"entity": schema.SingleNestedAttribute{
					Required:            true,
					MarkdownDescription: "The entity affected by this maintenance rule",
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The identifier of the entity (e.g., integration ID, policy ID)",
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The type of the entity (e.g., integration, policy, sync)",
							Validators: []validator.String{
								stringvalidator.OneOf("integration", "policy", "sync"),
							},
						},
					},
				},
			},
		},
	},
}
