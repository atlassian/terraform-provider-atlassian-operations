// Copyright (c) HashiCorp, Inc.

package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TeamRoleModel struct {
	ID               types.String `tfsdk:"id"`
	TeamId           types.String `tfsdk:"team_id"`
	Name             types.String `tfsdk:"name"`
	GrantedRights    types.Set    `tfsdk:"granted_rights"`
	DisallowedRights types.Set    `tfsdk:"disallowed_rights"`
}

var TeamRoleModelMap = map[string]attr.Type{
	"id":                types.StringType,
	"team_id":           types.StringType,
	"name":              types.StringType,
	"granted_rights":    types.SetType{ElemType: types.StringType},
	"disallowed_rights": types.SetType{ElemType: types.StringType},
}

func (receiver *TeamRoleModel) AsValue() types.Object {
	return types.ObjectValueMust(TeamRoleModelMap, map[string]attr.Value{
		"id":                receiver.ID,
		"team_id":           receiver.TeamId,
		"name":              receiver.Name,
		"granted_rights":    receiver.GrantedRights,
		"disallowed_rights": receiver.DisallowedRights,
	})
}
