package provider

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func orderingEntry(id types.String, order types.Int64) attr.Value {
	return types.ObjectValueMust(dataModels.PolicyOrderingEntryType.AttrTypes, map[string]attr.Value{"policy_id": id, "order": order})
}
func orderingEntryList(entries ...attr.Value) types.List {
	return types.ListValueMust(dataModels.PolicyOrderingEntryType, entries)
}

func TestPolicyOrderingExplicitInputValidation(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name      string
		entries   types.List
		want      []string
		known     bool
		errorText string
	}{
		{"shuffled", orderingEntryList(orderingEntry(types.StringValue("b"), types.Int64Value(2)), orderingEntry(types.StringValue("a"), types.Int64Value(1))), []string{"a", "b"}, true, ""},
		{"empty", orderingEntryList(), []string{}, true, ""},
		{"duplicate ID", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(1)), orderingEntry(types.StringValue("a"), types.Int64Value(2))), nil, false, "Duplicate Policy ID"},
		{"duplicate order", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(1)), orderingEntry(types.StringValue("b"), types.Int64Value(1))), nil, false, "Duplicate Policy Order"},
		{"gap", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(1)), orderingEntry(types.StringValue("b"), types.Int64Value(3))), nil, false, "Invalid Policy Order"},
		{"zero", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(0))), nil, false, "Invalid Policy Order"},
		{"negative", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(-1))), nil, false, "Invalid Policy Order"},
		{"null ID", orderingEntryList(orderingEntry(types.StringNull(), types.Int64Value(1))), nil, false, "Missing Policy ID"},
		{"null order", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Null())), nil, false, "Missing Policy Order"},
		{"empty ID", orderingEntryList(orderingEntry(types.StringValue(""), types.Int64Value(1))), nil, false, "Invalid Policy ID"},
		{"unknown ID allowed at plan", orderingEntryList(orderingEntry(types.StringUnknown(), types.Int64Value(1))), nil, false, ""},
		{"unknown order allowed at plan", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Unknown())), nil, false, ""},
		{"unknown ID invalid order", orderingEntryList(orderingEntry(types.StringUnknown(), types.Int64Value(2))), nil, false, "Invalid Policy Order"},
		{"unknown IDs duplicate order", orderingEntryList(orderingEntry(types.StringUnknown(), types.Int64Value(1)), orderingEntry(types.StringUnknown(), types.Int64Value(1))), nil, false, "Duplicate Policy Order"},
		{"unknown order duplicate ID", orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Unknown()), orderingEntry(types.StringValue("a"), types.Int64Value(1))), nil, false, "Duplicate Policy ID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, known, diagnostics := orderingSequence(ctx, tc.entries)
			if known != tc.known || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got=%v known=%v diagnostics=%v", got, known, diagnostics)
			}
			if diagnostics.HasError() != (tc.errorText != "") {
				t.Fatalf("diagnostics=%v", diagnostics)
			}
			if tc.errorText != "" && !strings.Contains(diagnostics.Errors()[0].Summary(), tc.errorText) {
				t.Fatalf("diagnostics=%v", diagnostics)
			}
			plan := orderingTestPlanEntries(t, tc.entries)
			response := resource.ValidateConfigResponse{}
			(&PolicyOrderingResource{}).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: plan.Schema, Raw: plan.Raw}}, &response)
			if response.Diagnostics.HasError() != diagnostics.HasError() {
				t.Fatalf("config validation disagrees: %v", response.Diagnostics)
			}
		})
	}
}

func TestPolicyOrderingExplicitCreatePreservesEntryOrder(t *testing.T) {
	entries := orderingEntryList(orderingEntry(types.StringValue("b"), types.Int64Value(2)), orderingEntry(types.StringValue("a"), types.Int64Value(1)))
	client := &memoryOrderingClient{ids: []string{"x", "b", "a"}}
	r := &PolicyOrderingResource{client: client}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
	r.Create(context.Background(), resource.CreateRequest{Plan: orderingTestPlanEntries(t, entries)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	if !reflect.DeepEqual(client.ids, []string{"a", "b", "x"}) {
		t.Fatal(client.ids)
	}
	data := decodeOrderingState(t, resp.State)
	if !data.Policies.Equal(entries) {
		t.Fatal("provider reordered configured entries")
	}
	if !data.ObservedOrder.Equal(orderingIDsValue([]string{"a", "b"})) {
		t.Fatal(data.ObservedOrder)
	}
}

func TestPolicyOrderingInvalidOrdersNeverMove(t *testing.T) {
	for _, entries := range []types.List{
		orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(0))),
		orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(2))),
		orderingEntryList(orderingEntry(types.StringValue("a"), types.Int64Value(1)), orderingEntry(types.StringValue("b"), types.Int64Value(1))),
	} {
		client := &memoryOrderingClient{ids: []string{"a", "b"}}
		resp := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
		(&PolicyOrderingResource{client: client}).Create(context.Background(), resource.CreateRequest{Plan: orderingTestPlanEntries(t, entries)}, &resp)
		if !resp.Diagnostics.HasError() || client.moves != 0 {
			t.Fatalf("diagnostics=%v moves=%d", resp.Diagnostics, client.moves)
		}
	}
}
