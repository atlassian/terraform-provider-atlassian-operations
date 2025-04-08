package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CustomRoleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	GrantedRights    types.Set    `tfsdk:"granted_rights"`
	DisallowedRights types.Set    `tfsdk:"disallowed_rights"`
}
