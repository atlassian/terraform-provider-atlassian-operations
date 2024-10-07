package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var RotationResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the rotation",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"schedule_id": schema.StringAttribute{
		Description: "The ID of the schedule",
		Required:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the rotation",
		Computed:    true,
		Optional:    true,
	},
	"start_date": schema.StringAttribute{
		Description: "The start date of the rotation",
		Required:    true,
	},
	"end_date": schema.StringAttribute{
		Description: "The end date of the rotation",
		Computed:    true,
		Optional:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the rotation",
		Required:    true,
	},
	"length": schema.Int32Attribute{
		Description: "The length of the rotation",
		Default:     int32default.StaticInt32(1),
		Computed:    true,
		Optional:    true,
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.UseStateForUnknown(),
		},
	},
	"participants": schema.ListNestedAttribute{
		Description: "The participants of the rotation",
		Computed:    true,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: ResponderInfoResourceAttributes,
		},
	},
	"time_restriction": schema.SingleNestedAttribute{
		Attributes: TimeRestrictionResourceAttributes,
		Computed:   true,
		Optional:   true,
	},
}

var ResponderInfoResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the participant",
		Computed:    true,
		Optional:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the participant",
		Required:    true,
	},
}
