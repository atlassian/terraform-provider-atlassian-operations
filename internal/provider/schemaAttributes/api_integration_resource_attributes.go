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
		Description: "The ID of the escalation",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 250),
		},
	},
	"type": schema.StringAttribute{
		Required: true,
	},
	"enabled": schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	},
	"team_id": schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"advanced": schema.BoolAttribute{
		Computed: true,
		Optional: false,
	},
	"maintenance_sources": schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: ApiIntegrationResourceMaintenanceSourceAttributes,
		},
		Computed: true,
		Optional: false,
	},
	"directions": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Optional:    false,
	},
	"domains": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Optional:    false,
	},
	"type_specific_properties": schema.StringAttribute{
		CustomType:  jsontypes.ExactType{},
		Computed:    true,
		Optional:    true,
		Description: "Integration specific properties may be provided to this object.",
	},
}

var ApiIntegrationResourceMaintenanceSourceAttributes = map[string]schema.Attribute{
	"maintenance_id": schema.StringAttribute{
		Description: "The ID of the maintenance",
		Computed:    true,
		Optional:    false,
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the maintenance is enabled",
		Computed:    true,
		Optional:    false,
	},
	"interval": schema.SingleNestedAttribute{
		Attributes: ApiIntegrationResourceMaintenanceSourceIntervalAttributes,
		Computed:   true,
		Optional:   false,
	},
}

var ApiIntegrationResourceMaintenanceSourceIntervalAttributes = map[string]schema.Attribute{
	"start_time_millis": schema.Int64Attribute{
		Description: "The start time of the maintenance",
		Computed:    true,
		Optional:    false,
	},
	"end_time_millis": schema.Int64Attribute{
		Description: "The end time of the maintenance",
		Computed:    true,
		Optional:    false,
	},
}
