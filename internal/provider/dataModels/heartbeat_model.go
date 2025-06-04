package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HeartbeatModel maps our data source attributes
type HeartbeatModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Interval      types.Int64  `tfsdk:"interval"`
	IntervalUnit  types.String `tfsdk:"interval_unit"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Status        types.String `tfsdk:"status"`
	TeamID        types.String `tfsdk:"team_id"`
	AlertMessage  types.String `tfsdk:"alert_message"`
	AlertTags     types.Set    `tfsdk:"alert_tags"`
	AlertPriority types.String `tfsdk:"alert_priority"`
}
