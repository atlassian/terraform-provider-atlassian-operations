package provider

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *PolicyOrderingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data dataModels.PolicyOrderingModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyTypeStr := data.PolicyType.ValueString()
	if !data.PolicyType.IsUnknown() && (data.PolicyType.IsNull() || (policyTypeStr != "alert" && policyTypeStr != "notification")) {
		resp.Diagnostics.AddAttributeError(path.Root("policy_type"), "Invalid Policy Type", "policy_type must be alert or notification.")
	}

	// Validate notification requires team
	if policyTypeStr == "notification" && data.TeamID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("team_id"),
			"Missing Required Team",
			"Notification policies require a team_id. Global scope is only supported for alert policies.",
		)
	}

	// Validate team_id format if present and known
	if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		if err := validatePolicyPathID(data.TeamID.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("team_id"), "Invalid Ordering Scope", err.Error())
		}
	}

	_, _, diagnostics := orderingSequence(ctx, data.Policies)
	resp.Diagnostics.Append(diagnostics...)
}

// orderingSequence validates every known field even when another field is
// unknown (for example new policy IDs). A dense, unique 1..N target range
// establishes the managed prefix without relying on configuration list order.
// The configured list itself must be preserved unchanged in Terraform state.
func orderingSequence(ctx context.Context, policies types.List) ([]string, bool, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if policies.IsUnknown() {
		return nil, false, diagnostics
	}
	if policies.IsNull() {
		diagnostics.AddAttributeError(path.Root("policies"), "Missing Policy Ordering", "policies must be a list of objects with policy_id and order; use [] to control no positions.")
		return nil, false, diagnostics
	}
	count := len(policies.Elements())
	ids := make([]string, count)
	seenIDs := make(map[string]int, count)
	seenOrders := make(map[int64]int, count)
	known := true
	for index, value := range policies.Elements() {
		entryPath := path.Root("policies").AtListIndex(index)
		if value.IsUnknown() {
			known = false
			continue
		}
		if value.IsNull() {
			diagnostics.AddAttributeError(entryPath, "Null Policy Ordering Entry", "Each entry must specify policy_id and order.")
			continue
		}
		object, ok := value.(types.Object)
		if !ok {
			diagnostics.AddAttributeError(entryPath, "Invalid Policy Ordering Entry", "Each entry must be an object with policy_id and order.")
			continue
		}
		var entry dataModels.PolicyOrderingEntry
		entryDiagnostics := object.As(ctx, &entry, basetypes.ObjectAsOptions{})
		diagnostics.Append(entryDiagnostics...)
		if entryDiagnostics.HasError() {
			continue
		}
		idValid, orderValid := false, false
		switch {
		case entry.PolicyID.IsUnknown():
			known = false
		case entry.PolicyID.IsNull():
			diagnostics.AddAttributeError(entryPath.AtName("policy_id"), "Missing Policy ID", "policy_id must not be null.")
		default:
			id := entry.PolicyID.ValueString()
			if err := validatePolicyPathID(id); err != nil {
				diagnostics.AddAttributeError(entryPath.AtName("policy_id"), "Invalid Policy ID", err.Error())
			} else if previous, exists := seenIDs[id]; exists {
				diagnostics.AddAttributeError(entryPath.AtName("policy_id"), "Duplicate Policy ID", fmt.Sprintf("policy_id %q is already specified in policies[%d]. Each ID must be unique.", id, previous))
			} else {
				seenIDs[id] = index
				idValid = true
			}
		}
		switch {
		case entry.Order.IsUnknown():
			known = false
		case entry.Order.IsNull():
			diagnostics.AddAttributeError(entryPath.AtName("order"), "Missing Policy Order", "order must not be null.")
		default:
			order := entry.Order.ValueInt64()
			if order < 1 || order > int64(count) {
				diagnostics.AddAttributeError(entryPath.AtName("order"), "Invalid Policy Order", fmt.Sprintf("order must be between 1 and %d for %d policies. Orders must cover 1..N without gaps.", count, count))
			} else if previous, exists := seenOrders[order]; exists {
				diagnostics.AddAttributeError(entryPath.AtName("order"), "Duplicate Policy Order", fmt.Sprintf("order %d is already specified in policies[%d]. Each position must be unique.", order, previous))
			} else {
				seenOrders[order] = index
				orderValid = true
			}
		}
		if idValid && orderValid {
			ids[entry.Order.ValueInt64()-1] = entry.PolicyID.ValueString()
		}
	}
	if diagnostics.HasError() || !known {
		return nil, false, diagnostics
	}
	return ids, true, diagnostics
}

func orderingIDsValue(ids []string) types.List {
	values := make([]attr.Value, len(ids))
	for i, id := range ids {
		values[i] = types.StringValue(id)
	}
	return types.ListValueMust(types.StringType, values)
}

// Used only when adopting a complete ordered list (import/state upgrade).
func orderingEntriesValue(ids []string) types.List {
	entries := make([]attr.Value, len(ids))
	for index, id := range ids {
		entries[index] = types.ObjectValueMust(dataModels.PolicyOrderingEntryType.AttrTypes, map[string]attr.Value{
			"policy_id": types.StringValue(id),
			"order":     types.Int64Value(int64(index + 1)),
		})
	}
	return types.ListValueMust(dataModels.PolicyOrderingEntryType, entries)
}
