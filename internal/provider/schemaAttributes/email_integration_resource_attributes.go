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

var EmailIntegrationResourceAttributes = map[string]schema.Attribute{
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
	"enabled": schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(true),
	},
	"team_id": schema.StringAttribute{
		Optional: true,
		Computed: true,
	},
	"advanced": schema.BoolAttribute{
		Optional: true,
		Required: false,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	},
	"maintenance_sources": schema.ListNestedAttribute{
		Optional: true,
		Required: false,
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: EmailIntegrationMaintenanceSourceResourceAttributes,
		},
	},
	"directions": schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Required:    false,
		Computed:    true,
		Description: "Direction of the action. It can be incoming or outgoing",
	},
	"domains": schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Required:    false,
		Computed:    true,
		Description: "Domain of the action. It can be alert",
	},
	"type_specific_properties": schema.SingleNestedAttribute{
		Attributes:  EmailIntegrationTypeSpecificAttributesResourceAttributes,
		Required:    true,
		Description: "Integration specific properties may be provided to this object.",
	},
}

var EmailIntegrationTypeSpecificAttributesResourceAttributes = map[string]schema.Attribute{
	"email_username": schema.StringAttribute{
		Required: true,
	},
	"suppress_notifications": schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	},
}

var EmailIntegrationMaintenanceSourceResourceAttributes = map[string]schema.Attribute{
	"maintenance_id": schema.StringAttribute{
		Optional: false,
		Required: false,
		Computed: true,
	}, "enabled": schema.BoolAttribute{
		Optional: false,
		Required: false,
		Computed: true,
	}, "interval": schema.SingleNestedAttribute{
		Attributes: EmailIntegrationMaintenanceSourceMaintenanceIntervalResourceAttributes,
		Optional:   false,
		Required:   false,
		Computed:   true,
	},
}

var EmailIntegrationMaintenanceSourceMaintenanceIntervalResourceAttributes = map[string]schema.Attribute{
	"start_time_millis": schema.Int32Attribute{
		Optional: false,
		Required: false,
		Computed: true,
	},
	"end_time_millis": schema.Int32Attribute{
		Optional: false,
		Required: false,
		Computed: true,
	},
}
