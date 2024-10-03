package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type (
	UserModel struct {
		AccountId        types.String `tfsdk:"account_id"`
		AccountType      types.String `tfsdk:"account_type"`
		Active           types.Bool   `tfsdk:"active"`
		ApplicationRoles types.List   `tfsdk:"application_roles"`
		AvatarUrls       types.Object `tfsdk:"avatar_urls"`
		DisplayName      types.String `tfsdk:"display_name"`
		EmailAddress     types.String `tfsdk:"email_address"`
		Expand           types.String `tfsdk:"expand"`
		Groups           types.List   `tfsdk:"groups"`
		Locale           types.String `tfsdk:"locale"`
		TimeZone         types.String `tfsdk:"timezone"`
	}
	ApplicationRoleModel struct {
		DefaultGroups        types.List   `tfsdk:"default_groups"`
		DefaultGroupsDetails types.List   `tfsdk:"default_groups_details"`
		Defined              types.Bool   `tfsdk:"defined"`
		GroupDetails         types.List   `tfsdk:"group_details"`
		Groups               types.List   `tfsdk:"groups"`
		HasUnlimitedSeats    types.Bool   `tfsdk:"has_unlimited_seats"`
		Key                  types.String `tfsdk:"key"`
		Name                 types.String `tfsdk:"name"`
		NumberOfSeats        types.Int32  `tfsdk:"number_of_seats"`
		Platform             types.Bool   `tfsdk:"platform"`
	}
	GroupNameModel struct {
		GroupId types.String `tfsdk:"group_id"`
		Name    types.String `tfsdk:"name"`
		Self    types.String `tfsdk:"self"`
	}
	AvatarUrlsBeanModel struct {
		A16x16 types.String `tfsdk:"a_16x16"`
		A24x24 types.String `tfsdk:"a_24x24"`
		A32x32 types.String `tfsdk:"a_32x32"`
		A48x48 types.String `tfsdk:"a_48x48"`
	}
)

var UserModelMap = map[string]attr.Type{
	"account_id":   types.StringType,
	"account_type": types.StringType,
	"active":       types.BoolType,
	"application_roles": types.ListType{ElemType: types.ObjectType{
		AttrTypes: ApplicationRoleModelMap,
	}},
	"avatar_urls": types.ObjectType{
		AttrTypes: AvatarUrlsBeanModelMap,
	},
	"display_name":  types.StringType,
	"email_address": types.StringType,
	"expand":        types.StringType,
	"groups": types.ListType{ElemType: types.ObjectType{
		AttrTypes: GroupNameModelMap,
	}},
	"locale":   types.StringType,
	"timezone": types.StringType,
}

var ApplicationRoleModelMap = map[string]attr.Type{
	"default_groups": types.ListType{ElemType: types.StringType},
	"default_groups_details": types.ListType{ElemType: types.ObjectType{
		AttrTypes: GroupNameModelMap,
	}},
	"defined": types.BoolType,
	"group_details": types.ListType{ElemType: types.ObjectType{
		AttrTypes: GroupNameModelMap,
	}},
	"groups":              types.ListType{ElemType: types.StringType},
	"has_unlimited_seats": types.BoolType,
	"key":                 types.StringType,
	"name":                types.StringType,
	"number_of_seats":     types.Int32Type,
	"platform":            types.BoolType,
}

var GroupNameModelMap = map[string]attr.Type{
	"group_id": types.StringType,
	"name":     types.StringType,
	"self":     types.StringType,
}

var AvatarUrlsBeanModelMap = map[string]attr.Type{
	"a_16x16": types.StringType,
	"a_24x24": types.StringType,
	"a_32x32": types.StringType,
	"a_48x48": types.StringType,
}

func (receiver *UserModel) AsValue() types.Object {
	return types.ObjectValueMust(UserModelMap, map[string]attr.Value{
		"account_id":        receiver.AccountId,
		"account_type":      receiver.AccountType,
		"active":            receiver.Active,
		"application_roles": receiver.ApplicationRoles,
		"avatar_urls":       receiver.AvatarUrls,
		"display_name":      receiver.DisplayName,
		"email_address":     receiver.EmailAddress,
		"expand":            receiver.Expand,
		"groups":            receiver.Groups,
		"locale":            receiver.Locale,
		"timezone":          receiver.TimeZone,
	})
}

func (receiver *ApplicationRoleModel) AsValue() types.Object {
	return types.ObjectValueMust(ApplicationRoleModelMap, map[string]attr.Value{
		"default_groups":         receiver.DefaultGroups,
		"default_groups_details": receiver.DefaultGroupsDetails,
		"defined":                receiver.Defined,
		"group_details":          receiver.GroupDetails,
		"groups":                 receiver.Groups,
		"has_unlimited_seats":    receiver.HasUnlimitedSeats,
		"key":                    receiver.Key,
		"name":                   receiver.Name,
		"number_of_seats":        receiver.NumberOfSeats,
		"platform":               receiver.Platform,
	})
}

func (receiver *GroupNameModel) AsValue() types.Object {
	return types.ObjectValueMust(GroupNameModelMap, map[string]attr.Value{
		"group_id": receiver.GroupId,
		"name":     receiver.Name,
		"self":     receiver.Self,
	})
}

func (receiver *AvatarUrlsBeanModel) AsValue() types.Object {
	return types.ObjectValueMust(AvatarUrlsBeanModelMap, map[string]attr.Value{
		"a_16x16": receiver.A16x16,
		"a_24x24": receiver.A24x24,
		"a_32x32": receiver.A32x32,
		"a_48x48": receiver.A48x48,
	})
}
