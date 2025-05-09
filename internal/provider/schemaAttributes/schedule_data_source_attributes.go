package schemaAttributes

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

var ScheduleDataSourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the schedule. This is automatically generated when the schedule is created.",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the schedule. This is used to look up the schedule and must be unique within your organization.",
		Required:    true,
	},
	"description": schema.StringAttribute{
		Description: "A detailed description of the schedule's purpose and coverage. This helps team members understand the schedule's role.",
		Computed:    true,
	},
	"timezone": schema.StringAttribute{
		Description: "The timezone in IANA format (e.g., 'America/New_York') that this schedule operates in. All times in the schedule are interpreted in this timezone.",
		Computed:    true,
	},
	"enabled": schema.BoolAttribute{
		Description: "Indicates whether the schedule is currently active and can be used for rotations and assignments.",
		Computed:    true,
	},
	"team_id": schema.StringAttribute{
		Description: "The unique identifier of the team that owns this schedule. Used for access control and organization.",
		Computed:    true,
	},
}
