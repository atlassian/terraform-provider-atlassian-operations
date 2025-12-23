// Copyright (c) HashiCorp, Inc.

package schemaAttributes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var TeamRoleResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The unique identifier of the team role. This is automatically generated when the team role is created.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"team_id": schema.StringAttribute{
		Description: "The ID of the team that owns this role. This field is immutable after creation.",
		Required:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"name": schema.StringAttribute{
		Description: "The name of the team role. This helps identify the role's purpose and the permissions it grants.",
		Required:    true,
	},
	"granted_rights": schema.SetAttribute{
		Description: "Set of permissions that are granted to this team role. Valid values include: access-reports, delete-escalations, delete-heartbeats, delete-integrations, delete-maintenance, delete-policies, delete-routing-rules, delete-schedules, delete-team-roles, edit-escalations, edit-heartbeats, edit-integrations, edit-maintenance, edit-policies, edit-routing-rules, edit-schedules, edit-team-roles.",
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
		Validators: []validator.Set{
			&setStringElementValidator{
				allowedValues: []string{
					"access-reports",
					"delete-escalations",
					"delete-heartbeats",
					"delete-integrations",
					"delete-maintenance",
					"delete-policies",
					"delete-routing-rules",
					"delete-schedules",
					"delete-team-roles",
					"edit-escalations",
					"edit-heartbeats",
					"edit-integrations",
					"edit-maintenance",
					"edit-policies",
					"edit-routing-rules",
					"edit-schedules",
					"edit-team-roles",
				},
			},
		},
	},
	"disallowed_rights": schema.SetAttribute{
		Description: "Set of permissions that are explicitly denied for this team role. Valid values include: access-reports, delete-escalations, delete-heartbeats, delete-integrations, delete-maintenance, delete-policies, delete-routing-rules, delete-schedules, delete-team-roles, edit-escalations, edit-heartbeats, edit-integrations, edit-maintenance, edit-policies, edit-routing-rules, edit-schedules, edit-team-roles.",
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
		Validators: []validator.Set{
			&setStringElementValidator{
				allowedValues: []string{
					"access-reports",
					"delete-escalations",
					"delete-heartbeats",
					"delete-integrations",
					"delete-maintenance",
					"delete-policies",
					"delete-routing-rules",
					"delete-schedules",
					"delete-team-roles",
					"edit-escalations",
					"edit-heartbeats",
					"edit-integrations",
					"edit-maintenance",
					"edit-policies",
					"edit-routing-rules",
					"edit-schedules",
					"edit-team-roles",
				},
			},
		},
	},
}

// setStringElementValidator validates that all elements in a set are from an allowed list
type setStringElementValidator struct {
	allowedValues []string
}

func (v *setStringElementValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("value must be one of: %v", v.allowedValues)
}

func (v *setStringElementValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("value must be one of: `%v`", v.allowedValues)
}

func (v *setStringElementValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var elements []string
	diags := req.ConfigValue.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	allowedMap := make(map[string]bool)
	for _, allowed := range v.allowedValues {
		allowedMap[allowed] = true
	}

	for _, element := range elements {
		if !allowedMap[element] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Value",
				fmt.Sprintf("Value %q is not valid. Must be one of: %v", element, v.allowedValues),
			)
		}
	}
}
