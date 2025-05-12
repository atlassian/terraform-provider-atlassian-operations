package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserContactModel struct {
	ID      types.String `tfsdk:"id"`
	Method  types.String `tfsdk:"method"`
	To      types.String `tfsdk:"to"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

var UserContactModelMap = map[string]attr.Type{
	"id":      types.StringType,
	"method":  types.StringType,
	"to":      types.StringType,
	"enabled": types.BoolType,
}

func (receiver *UserContactModel) AsValue() types.Object {
	return types.ObjectValueMust(UserContactModelMap, map[string]attr.Value{
		"id":      receiver.ID,
		"method":  receiver.Method,
		"to":      receiver.To,
		"enabled": receiver.Enabled,
	})
}
