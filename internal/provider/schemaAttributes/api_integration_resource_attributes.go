package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ApiIntegrationResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the API integration. This is automatically generated when the integration is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the API integration. Must be between 1 and 250 characters.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 250),
		},
	},
	"type": schema.StringAttribute{
		Description: "The type of API integration.",
		Required:    true,
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the API integration is enabled. When disabled, the integration will not process any requests. Defaults to false.",
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns this API integration. Cannot be changed after creation.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"advanced": schema.BoolAttribute{
		Description: "Indicates whether this is an advanced API integration with additional configuration options.",
		Computed:    true,
		Optional:    false,
	},
	"maintenance_sources": schema.ListNestedAttribute{
		Description: "List of maintenance windows associated with this API integration. These define when the integration is under maintenance.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: ApiIntegrationResourceMaintenanceSourceAttributes,
		},
		Computed: true,
		Required: false,
		Optional: false,
	},
	"directions": schema.ListAttribute{
		Description: "List of supported communication directions for this integration (e.g., 'inbound', 'outbound').",
		ElementType: types.StringType,
		Computed:    true,
		Optional:    false,
	},
	"domains": schema.ListAttribute{
		Description: "List of domains associated with this API integration. Used for routing and security purposes.",
		ElementType: types.StringType,
		Computed:    true,
		Optional:    false,
	},
	"type_specific_properties": schema.StringAttribute{
		Description: "JSON object containing integration-specific configuration properties. The schema depends on the integration type.",
		CustomType:  jsontypes.ExactType{},
		Computed:    true,
		Optional:    true,
	},
}

var ApiIntegrationResourceMaintenanceSourceAttributes = map[string]schema.Attribute{
	"maintenance_id": schema.StringAttribute{
		Description: "The unique identifier of the maintenance window. This is automatically generated when the maintenance window is created.",
		Computed:    true,
		Optional:    false,
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the maintenance window is active. When enabled, the integration behavior may be modified during the maintenance period.",
		Computed:    true,
		Optional:    false,
	},
	"interval": schema.SingleNestedAttribute{
		Description: "The time interval during which the maintenance window is active.",
		Attributes:  ApiIntegrationResourceMaintenanceSourceIntervalAttributes,
		Computed:    true,
		Optional:    false,
	},
}

var ApiIntegrationResourceMaintenanceSourceIntervalAttributes = map[string]schema.Attribute{
	"start_time_millis": schema.Int64Attribute{
		Description: "The start time of the maintenance window in Unix milliseconds (UTC).",
		Computed:    true,
		Optional:    false,
	},
	"end_time_millis": schema.Int64Attribute{
		Description: "The end time of the maintenance window in Unix milliseconds (UTC).",
		Computed:    true,
		Optional:    false,
	},
}
