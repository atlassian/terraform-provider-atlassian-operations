package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var EmailIntegrationResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the email integration. This is automatically generated when the integration is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the email integration. Must be between 1 and 250 characters.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 250),
		},
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the email integration is enabled. When disabled, the integration will not process any emails. Defaults to true.",
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(true),
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns this email integration. Used for access control and organization.",
		Optional:    true,
		Computed:    true,
	},
	"advanced": schema.BoolAttribute{
		Description: "Indicates whether this is an advanced email integration with additional configuration options.",
		Optional:    false,
		Computed:    true,
	},
	"maintenance_sources": schema.ListNestedAttribute{
		Description: "List of maintenance windows associated with this email integration. These define when the integration is under maintenance.",
		Optional:    false,
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: EmailIntegrationMaintenanceSourceResourceAttributes,
		},
	},
	"directions": schema.ListAttribute{
		Description: "The communication directions supported by this integration. Valid values are 'incoming' for receiving emails and 'outgoing' for sending emails.",
		ElementType: types.StringType,
		Optional:    false,
		Computed:    true,
	},
	"domains": schema.ListAttribute{
		Description: "The domains this integration operates on. Currently supports 'alert' for alert-related email communications.",
		ElementType: types.StringType,
		Optional:    false,
		Computed:    true,
	},
	"type_specific_properties": schema.SingleNestedAttribute{
		Description: "Configuration properties specific to email integrations, such as email address and notification settings.",
		Attributes:  EmailIntegrationTypeSpecificAttributesResourceAttributes,
		Required:    true,
	},
}

var EmailIntegrationTypeSpecificAttributesResourceAttributes = map[string]schema.Attribute{
	"email_username": schema.StringAttribute{
		Description: "The email address used for this integration. This will be used as the sender/receiver address depending on the integration direction.",
		Required:    true,
	},
	"suppress_notifications": schema.BoolAttribute{
		Description: "Whether to suppress email notifications from this integration. When true, no notification emails will be sent. Defaults to false.",
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
	},
}

var EmailIntegrationMaintenanceSourceResourceAttributes = map[string]schema.Attribute{
	"maintenance_id": schema.StringAttribute{
		Description: "The unique identifier of the maintenance window. This is automatically generated when the maintenance window is created.",
		Optional:    false,
		Required:    false,
		Computed:    true,
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the maintenance window is active. When enabled, the integration behavior may be modified during the maintenance period.",
		Optional:    false,
		Required:    false,
		Computed:    true,
	},
	"interval": schema.SingleNestedAttribute{
		Description: "The time interval during which the maintenance window is active.",
		Attributes:  EmailIntegrationMaintenanceSourceMaintenanceIntervalResourceAttributes,
		Optional:    false,
		Required:    false,
		Computed:    true,
	},
}

var EmailIntegrationMaintenanceSourceMaintenanceIntervalResourceAttributes = map[string]schema.Attribute{
	"start_time_millis": schema.Int64Attribute{
		Description: "The start time of the maintenance window in Unix milliseconds (UTC).",
		Optional:    false,
		Computed:    true,
	},
	"end_time_millis": schema.Int64Attribute{
		Description: "The end time of the maintenance window in Unix milliseconds (UTC).",
		Optional:    false,
		Computed:    true,
	},
}
