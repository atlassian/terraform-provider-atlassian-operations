package dataModels

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PolicyOrderingModel owns explicit target positions and policy type, not the policies themselves.
type PolicyOrderingModel struct {
	ID            types.String `tfsdk:"id"`
	PolicyType    types.String `tfsdk:"policy_type"`
	TeamID        types.String `tfsdk:"team_id"`
	Policies      types.List   `tfsdk:"policies"`
	ObservedOrder types.List   `tfsdk:"observed_order"`
}

type PolicyOrderingEntry struct {
	PolicyID types.String `tfsdk:"policy_id"`
	Order    types.Int64  `tfsdk:"order"`
}

var PolicyOrderingEntryType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"policy_id": types.StringType,
	"order":     types.Int64Type,
}}
