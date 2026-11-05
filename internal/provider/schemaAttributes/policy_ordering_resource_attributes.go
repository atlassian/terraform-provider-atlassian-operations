package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var PolicyOrderingResourceAttributes = map[string]schema.Attribute{
	"policy_type": schema.StringAttribute{
		Required:      true,
		Description:   "Policy type: alert or notification. Notification policies require team_id. Changing type replaces the controller without deleting policies.",
		Validators:    []validator.String{stringvalidator.OneOf("alert", "notification")},
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	},
	"id": schema.StringAttribute{
		Computed:      true,
		Description:   "Logical ordering identity: team/<team_id>/<policy_type> or global/alert. This is not a remote policy ID.",
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"team_id": schema.StringAttribute{
		Optional:      true,
		Description:   "Team whose policies are ordered. Required for notification; omit only for global alert policies. Changing the scope replaces this ordering controller without deleting any policies.",
		Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	},
	"policies": schema.ListNestedAttribute{
		Required:    true,
		Description: "Policies and their explicit target positions. policy_id and order must each be unique; positions must cover 1 through N for N entries. Entry list order is irrelevant. Referenced policies are placed first; unlisted policies remain afterward. Reference resource IDs to wait for creation and omit top-level order on individual policy resources. Use one ordering owner per scope and policy type. An empty list controls no positions.",
		NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"policy_id": schema.StringAttribute{
				Required:    true,
				Description: "Policy ID in this team/global scope. Reference the individual policy resource's id attribute.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"order": schema.Int64Attribute{
				Required:    true,
				Description: "Desired one-based position. Must be unique and within 1..N for N configured policies; gaps are not supported.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
		}},
	},
	"observed_order": schema.ListAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "Observed first N policy IDs, where N is the number of policies entries. Compared with the configured IDs sorted by numeric order to detect drift. Do not use this observation to define desired positions.",
	},
}
