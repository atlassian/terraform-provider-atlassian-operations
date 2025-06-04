package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var HeartbeatResourceAttributes = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		Description: "The name of the heartbeat. This is a unique identifier and cannot be changed after creation.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"description": schema.StringAttribute{
		Description: "Description of the heartbeat.",
		Optional:    true,
	},
	"interval": schema.Int64Attribute{
		Description: "The interval value for the heartbeat check.",
		Required:    true,
	},
	"interval_unit": schema.StringAttribute{
		Description: "The unit for the interval (e.g., 'minutes', 'hours', 'days').",
		Required:    true,
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the heartbeat is enabled or not.",
		Optional:    true,
	},
	"status": schema.StringAttribute{
		Description: "The current status of the heartbeat.",
		Computed:    true,
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns the heartbeat.",
		Required:    true,
	},
	"alert_message": schema.StringAttribute{
		Description: "The message to be displayed when an alert is triggered due to missed heartbeat.",
		Optional:    true,
	},
	"alert_tags": schema.SetAttribute{
		Description: "Tags to be associated with the alert when triggered.",
		Optional:    true,
		ElementType: types.StringType,
	},
	"alert_priority": schema.StringAttribute{
		Description: "The priority of the alert to be created when heartbeat is missed (e.g., 'P1', 'P2').",
		Optional:    true,
	},
}
