package provider

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func orderingTestSchema() schema.Schema {
	return schema.Schema{Attributes: schemaAttributes.PolicyOrderingResourceAttributes}
}
func orderingTestPlan(t *testing.T, ids ...string) tfsdk.Plan {
	t.Helper()
	if ids == nil {
		ids = []string{}
	}
	return orderingTestPlanEntries(t, orderingEntriesValue(ids))
}

func orderingTestPlanEntries(t *testing.T, entries types.List) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: orderingTestSchema()}
	d := plan.Set(context.Background(), &dataModels.PolicyOrderingModel{PolicyType: types.StringValue("alert"), ID: types.StringUnknown(), TeamID: types.StringNull(), Policies: entries, ObservedOrder: types.ListUnknown(types.StringType)})
	if d.HasError() {
		t.Fatal(d)
	}
	return plan
}
func orderingTestDesired(t *testing.T, data dataModels.PolicyOrderingModel) types.List {
	t.Helper()
	_, _, ids, err := policyOrderingInputs(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	return orderingIDsValue(ids)
}

func decodeOrderingState(t *testing.T, state tfsdk.State) dataModels.PolicyOrderingModel {
	t.Helper()
	var data dataModels.PolicyOrderingModel
	if d := state.Get(context.Background(), &data); d.HasError() {
		t.Fatal(d)
	}
	return data
}

func TestPolicyOrderingResourceLifecycle(t *testing.T) {
	ctx := context.Background()
	client := &memoryOrderingClient{ids: []string{"x", "b", "y", "a"}}
	r := &PolicyOrderingResource{client: client}
	create := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
	r.Create(ctx, resource.CreateRequest{Plan: orderingTestPlan(t, "a", "b")}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	if !reflect.DeepEqual(client.ids, []string{"a", "b", "x", "y"}) {
		t.Fatal(client.ids)
	}
	state := decodeOrderingState(t, create.State)
	if !orderingTestDesired(t, state).Equal(state.ObservedOrder) || state.ID.ValueString() != "global/alert" {
		t.Fatalf("state=%+v", state)
	}

	// Drift and a missing requested policy must not shrink desired intent.
	client.set("b", "x", "y")
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	if read.Diagnostics.HasError() {
		t.Fatal(read.Diagnostics)
	}
	observed := decodeOrderingState(t, read.State)
	if !observed.Policies.Equal(state.Policies) || observed.ObservedOrder.Equal(orderingTestDesired(t, state)) {
		t.Fatalf("bad drift state %+v", observed)
	}

	// Replacement/new membership is an ID change, not ownership of deletion.
	client.set("b", "x", "y", "new-a", "old-a")
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{Plan: orderingTestPlan(t, "new-a", "b")}, &update)
	if update.Diagnostics.HasError() {
		t.Fatal(update.Diagnostics)
	}
	if !reflect.DeepEqual(client.ids, []string{"new-a", "b", "x", "y", "old-a"}) {
		t.Fatal(client.ids)
	}
	moves := client.moves
	deleted := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &deleted)
	if !deleted.State.Raw.IsNull() || client.moves != moves || len(client.ids) != 5 {
		t.Fatal("delete modified remote policies")
	}
}

func TestPolicyOrderingValidatesBeforeMoving(t *testing.T) {
	for _, ids := range [][]string{{"a", "missing"}, {"a", "a"}, {""}} {
		client := &memoryOrderingClient{ids: []string{"a"}}
		r := &PolicyOrderingResource{client: client}
		resp := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
		r.Create(context.Background(), resource.CreateRequest{Plan: orderingTestPlan(t, ids...)}, &resp)
		if !resp.Diagnostics.HasError() || client.moves != 0 {
			t.Fatalf("ids=%v diagnostics=%v moves=%d", ids, resp.Diagnostics, client.moves)
		}
	}
}

func TestPolicyOrderingPartialFailureAndRetry(t *testing.T) {
	client := &memoryOrderingClient{ids: []string{"c", "b", "a"}, failMove: 2}
	r := &PolicyOrderingResource{client: client}
	plan := orderingTestPlan(t, "a", "b", "c")
	resp := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("missing move error")
	}
	state := decodeOrderingState(t, resp.State)
	if state.ID.ValueString() != "global/alert" || state.ObservedOrder.Equal(orderingTestDesired(t, state)) {
		t.Fatalf("false success state: %+v", state)
	}
	client.failMove = 0
	update := resource.UpdateResponse{State: resp.State}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan}, &update)
	if update.Diagnostics.HasError() {
		t.Fatal(update.Diagnostics)
	}
	if !reflect.DeepEqual(client.ids, []string{"a", "b", "c"}) {
		t.Fatal(client.ids)
	}
}

func TestPolicyOrderingEmptyAndImport(t *testing.T) {
	client := &memoryOrderingClient{ids: []string{"b", "a"}}
	r := &PolicyOrderingResource{client: client}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
	r.Create(context.Background(), resource.CreateRequest{Plan: orderingTestPlan(t)}, &resp)
	if resp.Diagnostics.HasError() || client.moves != 0 {
		t.Fatalf("empty: %v", resp.Diagnostics)
	}
	for _, id := range []string{"global/alert", "team/t/alert"} {
		imp := resource.ImportStateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &imp)
		if imp.Diagnostics.HasError() {
			t.Fatal(imp.Diagnostics)
		}
		state := decodeOrderingState(t, imp.State)
		if !orderingTestDesired(t, state).Equal(state.ObservedOrder) || len(state.Policies.Elements()) != 2 {
			t.Fatalf("import=%+v", state)
		}
	}
}

func TestPolicyOrderingModifyPlan(t *testing.T) {
	ctx := context.Background()
	for _, drift := range []bool{false, true} {
		plan := orderingTestPlan(t, "a", "b")
		state := tfsdk.State{Schema: orderingTestSchema(), Raw: plan.Raw}
		data := decodeOrderingState(t, state)
		data.ID = types.StringValue("global/alert")
		data.ObservedOrder = orderingTestDesired(t, data)
		if drift {
			data.ObservedOrder = orderingPrefixValue([]policyPosition{{ID: "b"}, {ID: "a"}}, 2)
		}
		if d := state.Set(ctx, &data); d.HasError() {
			t.Fatal(d)
		}
		plan.Raw = state.Raw
		resp := resource.ModifyPlanResponse{Plan: plan}
		(&PolicyOrderingResource{}).ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, State: state}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatal(resp.Diagnostics)
		}
		var out dataModels.PolicyOrderingModel
		if d := resp.Plan.Get(ctx, &out); d.HasError() {
			t.Fatal(d)
		}
		if out.ObservedOrder.IsUnknown() != drift {
			t.Fatalf("drift=%v unknown=%v", drift, out.ObservedOrder.IsUnknown())
		}
		if !out.Policies.Equal(data.Policies) {
			t.Fatal("configured intent modified")
		}
	}
	plan := orderingTestPlan(t, "a")
	plan.Raw = tftypes.NewValue(plan.Raw.Type(), nil)
	resp := resource.ModifyPlanResponse{Plan: plan}
	(&PolicyOrderingResource{}).ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan}, &resp)
	if !resp.Plan.Raw.IsNull() {
		t.Fatal("destroy plan modified")
	}
}

func genericOrderingPlan(t *testing.T, pt types.String, team types.String, ids ...string) tfsdk.Plan {
	t.Helper()
	r := NewPolicyOrderingResource()
	var sr resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &sr)
	plan := tfsdk.Plan{Schema: sr.Schema}
	data := dataModels.PolicyOrderingModel{ID: types.StringUnknown(), PolicyType: pt, TeamID: team, Policies: orderingEntriesValue(ids), ObservedOrder: types.ListUnknown(types.StringType)}
	if d := plan.Set(context.Background(), &data); d.HasError() {
		t.Fatal(d)
	}
	return plan
}

func TestPolicyOrderingInvalidScopeAndType(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pt          types.String
		team        types.String
		configError bool
	}{
		{"global notification", types.StringValue("notification"), types.StringNull(), true},
		{"empty team", types.StringValue("notification"), types.StringValue(""), true},
		{"bad team", types.StringValue("alert"), types.StringValue("../other"), true},
		{"unknown type", types.StringUnknown(), types.StringValue("team"), false},
		{"unknown team", types.StringValue("notification"), types.StringUnknown(), false},
		{"unsupported type", types.StringValue("routing"), types.StringValue("team"), true},
		{"missing type", types.StringNull(), types.StringValue("team"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &memoryOrderingClient{ids: []string{"a"}}
			r := &PolicyOrderingResource{client: client}
			plan := genericOrderingPlan(t, tc.pt, tc.team, "a")
			validation := resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Raw: plan.Raw, Schema: plan.Schema}}, &validation)
			if validation.Diagnostics.HasError() != tc.configError {
				t.Fatalf("config: %v", validation.Diagnostics)
			}
			response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
			if !response.Diagnostics.HasError() || client.moves != 0 {
				t.Fatalf("diagnostics=%v moves=%d", response.Diagnostics, client.moves)
			}
		})
	}
	for _, pt := range []policyType{policyTypeNotification, "bad"} {
		api := orderingTestAPI(t, func(http.ResponseWriter, *http.Request) { t.Error("invalid scope reached API") })
		api.policyType = pt
		if _, err := api.List(context.Background(), ""); err == nil {
			t.Fatal("invalid global scope accepted")
		}
		if err := api.Move(context.Background(), "", "a", 0); err == nil {
			t.Fatal("invalid global move accepted")
		}
	}
	if _, err := policyEndpoint("bad", "team"); err == nil {
		t.Fatal("invalid team type accepted")
	}
}

func TestPolicyOrderingImportAndMembership(t *testing.T) {
	for _, id := range []string{"global/alert", "team/team/alert", "team/team/notification"} {
		t.Run(id, func(t *testing.T) {
			team, pt, _ := parsePolicyOrderingIdentity(id)
			client := &memoryOrderingClient{ids: []string{"b", "a"}}
			r := &PolicyOrderingResource{clientFactory: func(selected policyType) policyOrderingClient {
				if selected != pt {
					t.Fatalf("type=%s", selected)
				}
				return client
			}}
			plan := genericOrderingPlan(t, types.StringValue(string(pt)), types.StringNull(), "a")
			response := resource.ImportStateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			var data dataModels.PolicyOrderingModel
			if d := response.State.Get(context.Background(), &data); d.HasError() {
				t.Fatal(d)
			}
			if data.ID.ValueString() != id || data.PolicyType.ValueString() != string(pt) || data.TeamID.ValueString() != team || team == "" && !data.TeamID.IsNull() {
				t.Fatalf("state=%+v", data)
			}
			badPlan := genericOrderingPlan(t, data.PolicyType, data.TeamID, "other-type-id")
			create := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: badPlan}, &create)
			if !create.Diagnostics.HasError() || client.moves != 0 {
				t.Fatal("missing/wrong-type policy was moved")
			}
		})
	}
	for _, id := range []string{"global/notification", "team//alert", "team/team/unknown", "team/../alert", "team/a/b/notification"} {
		if _, _, err := parsePolicyOrderingIdentity(id); err == nil {
			t.Errorf("invalid import accepted: %s", id)
		}
	}
}

func TestPolicyOrderingConfigureAndRegistration(t *testing.T) {
	ctx := context.Background()
	provider := New("test")()
	found := false
	for _, factory := range provider.Resources(ctx) {
		r := factory()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "atlassian-operations"}, &metadata)
		if metadata.TypeName == "atlassian-operations_policy_ordering" {
			found = true
		}
	}
	if !found {
		t.Fatal("ordering resource not registered")
	}
	r := NewPolicyOrderingResource().(*PolicyOrderingResource)
	bad := resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong"}, &bad)
	if !bad.Diagnostics.HasError() {
		t.Fatal("invalid provider data accepted")
	}
	for _, product := range []string{"jira-service-desk", "compass"} {
		configuration := dto.NewAtlassianOpsProviderModel(product, "cloud", "example.atlassian.net", "test@example.com", "test-token", "", "", "", 2, time.Second, 2*time.Second, false)
		resp := resource.ConfigureResponse{}
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: configuration}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatal(resp.Diagnostics)
		}
		api, ok := r.clientForType(policyTypeAlert).(*policyOrderingAPI)
		if !ok {
			t.Fatal("missing API client")
		}
		req := api.requestFactory()
		if req.GetInnerClient().RetryMax != 2 {
			t.Fatal("provider retry settings not preserved")
		}
		if !strings.Contains(req.GetInnerRequest().URL.Path, "cloud") {
			t.Fatal("cloud ID not routed")
		}
	}
}

func TestPolicyOrderingUnknownInputs(t *testing.T) {
	known := orderingEntriesValue([]string{"a"})
	entryType := dataModels.PolicyOrderingEntryType
	for _, tc := range []struct {
		name string
		team types.String
		ids  types.List
	}{
		{"unknown team", types.StringUnknown(), known},
		{"empty team", types.StringValue(""), known},
		{"unknown list", types.StringNull(), types.ListUnknown(entryType)},
		{"null list", types.StringNull(), types.ListNull(entryType)},
		{"unknown entry", types.StringNull(), types.ListValueMust(entryType, []attr.Value{types.ObjectUnknown(entryType.AttrTypes)})},
		{"null entry", types.StringNull(), types.ListValueMust(entryType, []attr.Value{types.ObjectNull(entryType.AttrTypes)})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := policyOrderingInputs(context.Background(), dataModels.PolicyOrderingModel{PolicyType: types.StringValue("alert"), TeamID: tc.team, Policies: tc.ids}); err == nil {
				t.Fatal("invalid apply inputs accepted")
			}
		})
	}
}

type observingOrderingClient struct {
	*memoryOrderingClient
	observe func(context.Context, int) ([]policyPosition, error)
	reads   int
}

func (c *observingOrderingClient) List(ctx context.Context, _ string) ([]policyPosition, error) {
	c.reads++
	return c.observe(ctx, c.reads)
}

func TestPolicyOrderingVerificationAndCancellation(t *testing.T) {
	for _, scenario := range []string{"visibility lag", "never converges", "read failure", "cancel during verification", "initial read failure"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := &observingOrderingClient{memoryOrderingClient: &memoryOrderingClient{ids: []string{"b", "a"}}}
			stale := []policyPosition{{ID: "b", Order: 1}, {ID: "a", Order: 2}}
			client.observe = func(ctx context.Context, read int) ([]policyPosition, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if scenario == "initial read failure" || scenario == "read failure" && read > 1 {
					return nil, errors.New("read unavailable")
				}
				if scenario == "cancel during verification" && read == 2 {
					cancel()
				}
				if read <= 2 || scenario == "never converges" {
					return stale, nil
				}
				return client.memoryOrderingClient.List(ctx, "")
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
			r := &PolicyOrderingResource{client: client}
			r.Create(ctx, resource.CreateRequest{Plan: orderingTestPlan(t, "a", "b")}, &resp)
			if resp.Diagnostics.HasError() != (scenario != "visibility lag") {
				t.Fatalf("diagnostics=%v", resp.Diagnostics)
			}
			if scenario == "initial read failure" {
				if client.moves != 0 {
					t.Fatal("moved before initial read succeeded")
				}
				return
			}
			data := decodeOrderingState(t, resp.State)
			if data.ID.ValueString() != "global/alert" || client.moves != 2 {
				t.Fatalf("lost identity or repeated moves: state=%+v moves=%d", data, client.moves)
			}
			switch scenario {
			case "visibility lag":
				if client.reads != 3 || !data.ObservedOrder.Equal(orderingIDsValue([]string{"a", "b"})) {
					t.Fatalf("reads=%d observation=%v", client.reads, data.ObservedOrder)
				}
			case "never converges":
				if client.reads != 6 || !data.ObservedOrder.Equal(orderingIDsValue([]string{"b", "a"})) {
					t.Fatalf("unbounded polling or false success: reads=%d observation=%v", client.reads, data.ObservedOrder)
				}
			default:
				if !data.ObservedOrder.IsNull() {
					t.Fatal("unavailable read-back must not claim successful order")
				}
			}
		})
	}
}

func TestPolicyOrderingReadErrorPreservesState(t *testing.T) {
	ctx := context.Background()
	plan := orderingTestPlan(t, "a")
	state := tfsdk.State{Schema: orderingTestSchema(), Raw: plan.Raw}
	client := &memoryOrderingClient{ids: []string{"a"}}
	r := &PolicyOrderingResource{client: client}
	create := resource.CreateResponse{State: state}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	client.listErr = context.DeadlineExceeded
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	if !read.Diagnostics.HasError() || !read.State.Raw.Equal(create.State.Raw) {
		t.Fatal("read failure replaced valid state")
	}
	for _, id := range []string{"", "team//alert", "global"} {
		imp := resource.ImportStateResponse{State: tfsdk.State{Schema: orderingTestSchema()}}
		r.ImportState(ctx, resource.ImportStateRequest{ID: id}, &imp)
		if !imp.Diagnostics.HasError() {
			t.Fatalf("invalid import ID accepted: %q", id)
		}
	}
}
