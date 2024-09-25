package schemaAttributes

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

var ScheduleDataSourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the schedule",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the schedule",
		Required:    true,
	},
	"description": schema.StringAttribute{
		Description: "The description of the schedule",
		Computed:    true,
	},
	"timezone": schema.StringAttribute{
		Description: "The timezone of the schedule",
		Computed:    true,
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the schedule is enabled",
		Computed:    true,
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns the schedule",
		Computed:    true,
	},
	"rotations": schema.ListNestedAttribute{
		Description: "The rotations of the schedule",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: RotationDataSourceAttributes,
		},
	},
}
