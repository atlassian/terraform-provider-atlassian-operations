package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var ScheduleResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the schedule",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the schedule",
		Required:    true,
	},
	"description": schema.StringAttribute{
		Description: "The description of the schedule",
		Computed:    true,
		Optional:    true,
		Default:     stringdefault.StaticString(""),
	},
	"timezone": schema.StringAttribute{
		Description: "The timezone of the schedule",
		Computed:    true,
		Optional:    true,
		Default:     stringdefault.StaticString("America/New_York"),
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the schedule is enabled",
		Computed:    true,
		Optional:    true,
		Default:     booldefault.StaticBool(true),
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns the schedule",
		Required:    true,
	},
}
