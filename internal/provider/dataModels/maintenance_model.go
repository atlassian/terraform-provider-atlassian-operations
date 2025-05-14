package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MaintenanceModel represents the Terraform resource data model for a maintenance window
type MaintenanceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	StartDate   types.String `tfsdk:"start_date"`
	EndDate     types.String `tfsdk:"end_date"`
	Status      types.String `tfsdk:"status"`
	TeamID      types.String `tfsdk:"team_id"`
	Rules       types.List   `tfsdk:"rules"`
}

// MaintenanceRuleModel represents a rule within a maintenance window for Terraform
type MaintenanceRuleModel struct {
	State  types.String `tfsdk:"state"`
	Entity types.Object `tfsdk:"entity"`
}

// MaintenanceRuleEntityModel represents an entity affected by a maintenance rule
type MaintenanceRuleEntityModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

// MaintenanceRuleEntityObjectType defines the type for the entity object
var MaintenanceRuleEntityObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":   types.StringType,
		"type": types.StringType,
	},
}

// MaintenanceRuleObjectType defines the type for a rule object
var MaintenanceRuleObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"state": types.StringType,
		"entity": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":   types.StringType,
				"type": types.StringType,
			},
		},
	},
}
