package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
