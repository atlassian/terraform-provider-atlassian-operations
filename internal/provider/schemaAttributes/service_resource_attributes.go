package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ServiceResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the JSM service",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the JSM service",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 100),
		},
	},
	"description": schema.StringAttribute{
		Description: "The description of the JSM service",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 1000),
		},
	},
	"tier": schema.Int32Attribute{
		Description: "The tier level of the JSM service",
		Required:    true,
		Validators: []validator.Int32{
			int32validator.OneOf(1, 2, 3, 4),
		},
	},
	"type": schema.StringAttribute{
		Description: "The type of the JSM service",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("SOFTWARE_SERVICES", "BUSINESS_SERVICES", "CAPABILITIES_SERVICES", "APPLICATIONS"),
		},
	},
	"owner": schema.StringAttribute{
		Description: "The owner team ID of the JSM service. Omit or set to an empty string to remove the owner.",
		Optional:    true,
		Computed:    true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(0, 500),
		},
	},
	"change_approvers": schema.SingleNestedAttribute{
		Description: "Change approvers configuration for the JSM service",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListAttribute{
				Description: "List of group IDs for change approvers. If you want to remove all group change approvers, set this to an empty list.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	},
	"responders": schema.SingleNestedAttribute{
		Description: "Responders configuration for the JSM service",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"users": schema.ListAttribute{
				Description: "List of user IDs for responders. If you want to remove all users responders, set this to an empty list.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"teams": schema.ListAttribute{
				Description: "List of team IDs for responders. If you want to remove all teams responders, set this to an empty list.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	},
	"stakeholders": schema.SingleNestedAttribute{
		Description: "Stakeholders configuration for the JSM service",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"users": schema.ListAttribute{
				Description: "List of user IDs for stakeholders. If you want to remove all user stakeholders, set this to an empty list.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	},
	"projects": schema.SingleNestedAttribute{
		Description: "Projects configuration for the JSM service",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				Description: "List of project IDs. If you want to remove all project IDs, set this to an empty list.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	},
}
