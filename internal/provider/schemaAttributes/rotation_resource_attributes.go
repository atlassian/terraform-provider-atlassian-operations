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
		Description: "The unique identifier of the rotation. This is automatically generated when the rotation is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"schedule_id": schema.StringAttribute{
		Description: "The ID of the schedule this rotation belongs to. This links the rotation to a specific on-call schedule.",
		Required:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the rotation. Must be at least 1 character long. This helps identify the rotation's purpose.",
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
		Description: "The date and time when this rotation begins, in RFC3339 format (e.g., '2024-01-01T00:00:00Z').",
		Required:    true,
		CustomType:  timetypes.RFC3339Type{},
	},
	"end_date": schema.StringAttribute{
		Description: "The date and time when this rotation ends, in RFC3339 format. If not specified, the rotation continues indefinitely.",
		Computed:    true,
		Optional:    true,
		CustomType:  timetypes.RFC3339Type{},
	},
	"type": schema.StringAttribute{
		Description: "The frequency of rotation. Valid values are 'daily' (rotate every day), 'weekly' (rotate every week), or 'hourly' (rotate every hour).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{"daily", "weekly", "hourly"}...),
		},
	},
	"length": schema.Int32Attribute{
		Description: "The duration of each rotation shift in units matching the rotation type (hours for hourly, days for daily, weeks for weekly). Defaults to 1.",
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
		Description: "The list of participants in this rotation. Can include users, teams, escalation policies, or empty slots (noone).",
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
		Description: "Optional time restrictions for when this rotation is active. Used to define specific hours or days when the rotation applies.",
		Attributes:  TimeRestrictionResourceAttributes,
		Optional:    true,
	},
}

var ResponderInfoResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the participant (user ID, team ID, or escalation policy ID). Required when type is 'user', 'team', or 'escalation'.",
		Optional:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of participant. Valid values are 'user' (individual user), 'team' (entire team), 'escalation' (escalation policy), or 'noone' (empty slot).",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{"user", "team", "escalation", "noone"}...),
		},
	},
}
