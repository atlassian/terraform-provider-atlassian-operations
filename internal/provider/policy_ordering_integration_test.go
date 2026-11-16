package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	testresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

// Test only the new controller's HTTP integration. Existing policy resources
// and their request paths are deliberately not replaced or modified.
type orderingHTTPTestProvider struct {
	*orderingTestProvider
	api policyOrderingClient
}

func (p *orderingHTTPTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &PolicyOrderingResource{client: p.api} },
	}
}

func TestPolicyOrderingTerraformHTTP(t *testing.T) {
	for _, tc := range []struct {
		pt   policyType
		team string
	}{
		{policyTypeAlert, ""}, {policyTypeAlert, "team"}, {policyTypeNotification, "team"},
	} {
		t.Run(string(tc.pt)+"/"+tc.team, func(t *testing.T) {
			client := &memoryOrderingClient{ids: []string{"outside", "b", "a"}}
			endpoint, err := policyEndpoint(tc.pt, tc.team)
			if err != nil {
				t.Fatal(err)
			}
			prefix := "/jsm/ops/api/cloud" + endpoint
			api := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == prefix:
					if r.URL.Query().Get("type") != string(tc.pt) {
						t.Errorf("wrong policy type: %s", r.URL)
					}
					positions, err := client.List(r.Context(), tc.team)
					if err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					values := make([]map[string]interface{}, 0, len(positions))
					for _, p := range positions {
						values = append(values, map[string]interface{}{"id": p.ID, "order": p.Order, "type": string(tc.pt)})
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"values": values})
				case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, prefix+"/") && strings.HasSuffix(r.URL.Path, "/change-order"):
					var body struct {
						Order int64 `json:"order"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix+"/"), "/change-order")
					if err := client.Move(r.Context(), tc.team, id, body.Order); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					_, _ = w.Write([]byte(`{}`))
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusBadRequest)
				}
			})
			api.policyType = tc.pt
			teamConfig := ""
			if tc.team != "" {
				teamConfig = fmt.Sprintf("team_id = %q", tc.team)
			}
			config := fmt.Sprintf(`resource "ordering_policy_ordering" "test" {
 policy_type = %q
 %s
 policies = [{policy_id="b",order=2},{policy_id="a",order=1}]
}`, tc.pt, teamConfig)
			empty := strings.ReplaceAll(config, `[{policy_id="b",order=2},{policy_id="a",order=1}]`, `[]`)
			provider := &orderingHTTPTestProvider{orderingTestProvider: &orderingTestProvider{client: client}, api: api}
			var movesBeforeEmpty int
			testresource.UnitTest(t, testresource.TestCase{
				WorkingDir:               orderingTerraformWorkingDir(t),
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(provider)},
				Steps: []testresource.TestStep{
					{Config: config, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"a", "b", "outside"}}}},
					{Config: config, PlanOnly: true},
					{Config: config, PreConfig: func() { client.set("b", "outside", "a") }, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"a", "b", "outside"}}}},
					{Config: config, ResourceName: "ordering_policy_ordering.test", ImportState: true, ImportStateId: policyOrderingIdentity(tc.team, tc.pt)},
					{Config: empty, PreConfig: func() { client.mu.Lock(); movesBeforeEmpty = client.moves; client.mu.Unlock() }},
					{Config: empty, PlanOnly: true},
					{Config: `provider "ordering" {}`},
				},
			})
			client.mu.Lock()
			defer client.mu.Unlock()
			if client.moves != movesBeforeEmpty || strings.Join(client.ids, ",") != "a,b,outside" {
				t.Fatal("empty list or destroy modified remote policies")
			}
		})
	}
}
