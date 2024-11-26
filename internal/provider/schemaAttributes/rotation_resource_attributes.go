package schemaAttributes

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes/customValidators"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	},
	"start_date": schema.StringAttribute{
		Description: "The start date of the rotation",
		Required:    true,
		CustomType:  timetypes.RFC3339Type{},
	},
	"end_date": schema.StringAttribute{
		Description: "The end date of the rotation",
		Computed:    true,
		Optional:    true,
		CustomType:  timetypes.RFC3339Type{},
	},
	"type": schema.StringAttribute{
		Description: "The type of the rotation",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{"daily", "weekly", "hourly"}...),
		},
	},
	"length": schema.Int32Attribute{
		Description: "The length of the rotation",
		Default:     int32default.StaticInt32(1),
		Computed:    true,
		Optional:    true,
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.UseStateForUnknown(),
		},
		Validators: []validator.Int32{
			int32validator.AtLeast(1),
		},
	},
	"participants": schema.ListNestedAttribute{
		Description: "The participants of the rotation",
		Optional:    true,
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: ResponderInfoResourceAttributes,
			Validators: []validator.Object{
				customValidators.StringFieldNotNullIfOtherField(path.MatchRelative().AtName("id"), path.MatchRelative().AtName("type"), "user"),
				customValidators.StringFieldNotNullIfOtherField(path.MatchRelative().AtName("id"), path.MatchRelative().AtName("type"), "team"),
				customValidators.StringFieldNotNullIfOtherField(path.MatchRelative().AtName("id"), path.MatchRelative().AtName("type"), "escalation"),
			},
		},
		Default: listdefault.StaticValue(
			types.ListValueMust(
				types.ObjectType{AttrTypes: dataModels.ResponderInfoModelMap},
				[]attr.Value{},
			),
		),
	},
	"time_restriction": schema.SingleNestedAttribute{
		Attributes: TimeRestrictionResourceAttributes,
		Optional:   true,
	},
}

var ResponderInfoResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the participant",
		Optional:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the participant",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{"user", "team", "escalation", "noone"}...),
		},
	},
}
