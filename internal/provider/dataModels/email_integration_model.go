package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	EmailIntegrationModel struct {
		Id                          types.String `tfsdk:"id"`
		Name                        types.String `tfsdk:"name"`
		Enabled                     types.Bool   `tfsdk:"enabled"`
		TeamId                      types.String `tfsdk:"team_id"`
		TypeSpecificPropertiesModel types.Object `tfsdk:"type_specific_properties"`
	}

	TypeSpecificPropertiesModel struct {
		EmailUsername         types.String `tfsdk:"email_username"`
		SuppressNotifications types.Bool   `tfsdk:"suppress_notifications"`
	}
)

var TypeSpecificPropertiesModelMap = map[string]attr.Type{
	"email_username":         types.StringType,
	"suppress_notifications": types.BoolType,
}

var EmailIntegrationModelMap = map[string]attr.Type{
	"id":                       types.StringType,
	"name":                     types.StringType,
	"type":                     types.StringType,
	"timezone":                 types.StringType,
	"enabled":                  types.BoolType,
	"team_id":                  types.StringType,
	"type_specific_properties": types.ObjectType{AttrTypes: TypeSpecificPropertiesModelMap},
}

func (receiver *TypeSpecificPropertiesModel) AsValue() types.Object {
	return types.ObjectValueMust(TypeSpecificPropertiesModelMap, map[string]attr.Value{
		"email_username":         receiver.EmailUsername,
		"suppress_notifications": receiver.SuppressNotifications,
	})
}

func (receiver *EmailIntegrationModel) AsValue() types.Object {
	return types.ObjectValueMust(EmailIntegrationModelMap, map[string]attr.Value{
		"id":      receiver.Id,
		"name":    receiver.Name,
		"enabled": receiver.Enabled,
		"team_id": receiver.TeamId,
	})
}