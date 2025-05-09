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
		Description: "The unique identifier of the schedule. This is automatically generated when the schedule is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the schedule. This helps identify the schedule's purpose and the team or service it covers.",
		Required:    true,
	},
	"description": schema.StringAttribute{
		Description: "A detailed description of the schedule's purpose, coverage, and any special instructions. Defaults to empty string.",
		Computed:    true,
		Optional:    true,
		Default:     stringdefault.StaticString(""),
	},
	"timezone": schema.StringAttribute{
		Description: "The timezone in IANA format (e.g., 'America/New_York') that this schedule operates in. All rotations and shifts are interpreted in this timezone. Defaults to 'America/New_York'.",
		Computed:    true,
		Optional:    true,
		Default:     stringdefault.StaticString("America/New_York"),
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the schedule is active and can be used for on-call rotations. When disabled, no notifications will be sent to participants. Defaults to true.",
		Computed:    true,
		Optional:    true,
		Default:     booldefault.StaticBool(true),
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns this schedule. Used for access control and organization of schedules.",
		Required:    true,
	},
}
