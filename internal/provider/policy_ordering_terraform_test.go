package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	testresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

// These tests run the real Terraform CLI and framework protocol against an
// in-memory policy service. They never configure credentials or call JSM.
type orderingTestProvider struct{ client *memoryOrderingClient }

func (p *orderingTestProvider) Metadata(_ context.Context, _ fwprovider.MetadataRequest, resp *fwprovider.MetadataResponse) {
	resp.TypeName = "ordering"
}
func (p *orderingTestProvider) Schema(_ context.Context, _ fwprovider.SchemaRequest, resp *fwprovider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}
func (p *orderingTestProvider) Configure(context.Context, fwprovider.ConfigureRequest, *fwprovider.ConfigureResponse) {
}
func (p *orderingTestProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}
func (p *orderingTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &PolicyOrderingResource{client: p.client} },
		func() resource.Resource { return &orderingTestPolicy{client: p.client} },
	}
}

type orderingTestPolicy struct{ client *memoryOrderingClient }
type orderingTestPolicyModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *orderingTestPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}
func (r *orderingTestPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
	}}
}
func (r *orderingTestPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data orderingTestPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = data.Name
	r.client.mu.Lock()
	r.client.ids = append(r.client.ids, data.ID.ValueString())
	r.client.mu.Unlock()
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *orderingTestPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data orderingTestPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client.mu.Lock()
	defer r.client.mu.Unlock()
	for _, id := range r.client.ids {
		if id == data.ID.ValueString() {
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *orderingTestPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.State.Raw = req.Plan.Raw
}
func (r *orderingTestPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data orderingTestPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client.mu.Lock()
	defer r.client.mu.Unlock()
	for i, id := range r.client.ids {
		if id == data.ID.ValueString() {
			r.client.ids = append(r.client.ids[:i], r.client.ids[i+1:]...)
			break
		}
	}
}

func orderingTerraformWorkingDir(t *testing.T) string {
	t.Helper()
	configured := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if configured == "" {
		configured = os.Getenv("TF_CLI_PATH")
	}
	command := configured
	if command == "" {
		command = "terraform"
	}
	binary, err := exec.LookPath(command)
	if err != nil {
		if configured != "" || os.Getenv("CI") != "" || os.Getenv("BITBUCKET_BUILD_NUMBER") != "" {
			t.Fatalf("Terraform CLI required for protocol tests: %v", err)
		}
		t.Skip("Terraform CLI is required for local protocol tests")
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", binary)
	dir, err := os.MkdirTemp(".", "tmp_rovo_ordering_")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})
	absolute, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return absolute
}

func orderingTerraformConfig(sequence, names string) string {
	return fmt.Sprintf(`
resource "ordering_policy" "policies" {
  for_each = %s
  name = each.value
  lifecycle { create_before_destroy = true }
}
resource "ordering_policy_ordering" "test" {
  # Deliberately reverse the entry list: numeric order, not list order, wins.
  policies = reverse([for index, key in %s : {
    policy_id = ordering_policy.policies[key].id
    order     = index + 1
  }])
}
`, names, sequence)
}

type orderingRemoteCheck struct {
	client *memoryOrderingClient
	want   []string
}

func (c orderingRemoteCheck) CheckState(ctx context.Context, _ statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	positions, err := c.client.List(ctx, "")
	if err != nil {
		resp.Error = err
		return
	}
	got := orderingPrefix(positions, len(positions))
	if !reflect.DeepEqual(got, c.want) {
		resp.Error = fmt.Errorf("remote=%v want=%v", got, c.want)
	}
}

func TestPolicyOrderingTerraformRejectsDuplicatePositions(t *testing.T) {
	client := &memoryOrderingClient{}
	config := `
resource "ordering_policy" "policies" {
  for_each = toset(["a", "b"])
  name = each.key
}
resource "ordering_policy_ordering" "test" {
 policy_type = "alert"
  policies = [for p in ordering_policy.policies : { policy_id = p.id, order = 1 }]
}`
	testresource.UnitTest(t, testresource.TestCase{
		WorkingDir:               orderingTerraformWorkingDir(t),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(&orderingTestProvider{client: client})},
		Steps:                    []testresource.TestStep{{Config: config, ExpectError: regexp.MustCompile("Duplicate Policy Order")}},
	})
	if len(client.ids) != 0 || client.moves != 0 {
		t.Fatalf("invalid positions caused side effects: ids=%v moves=%d", client.ids, client.moves)
	}
}

type genericOrderingTestProvider struct {
	*orderingTestProvider
	clients map[policyType]*memoryOrderingClient
}

func (p *genericOrderingTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &PolicyOrderingResource{clientFactory: func(pt policyType) policyOrderingClient { return p.clients[pt] }}
		},
		func() resource.Resource { return &orderingTestPolicy{client: p.client} },
	}
}

func TestPolicyOrderingTerraformBothTypes(t *testing.T) {
	for _, pt := range []policyType{policyTypeAlert, policyTypeNotification} {
		t.Run(string(pt), func(t *testing.T) {
			client := &memoryOrderingClient{}
			provider := &genericOrderingTestProvider{orderingTestProvider: &orderingTestProvider{client: client}, clients: map[policyType]*memoryOrderingClient{pt: client}}
			config := orderingTerraformConfig(`["a","b","c"]`, `toset(["c","a","b"])`)
			config = strings.ReplaceAll(config, `"test" {`, `"test" {`+fmt.Sprintf("\n policy_type = %q\n team_id = \"team\"\n", pt))
			updated := strings.ReplaceAll(config, `["a","b","c"]`, `["c","b","a"]`)
			expanded := strings.ReplaceAll(updated, `toset(["c","a","b"])`, `toset(["c","a","b","d"])`)
			expanded = strings.ReplaceAll(expanded, `["c","b","a"]`, `["d","c","b","a"]`)
			replaced := strings.ReplaceAll(expanded, `"a"`, `"new-a"`)
			shrunk := strings.ReplaceAll(replaced, `toset(["c","new-a","b","d"])`, `toset(["new-a","b"])`)
			shrunk = strings.ReplaceAll(shrunk, `["d","c","b","new-a"]`, `["new-a","b"]`)
			testresource.UnitTest(t, testresource.TestCase{
				WorkingDir:               orderingTerraformWorkingDir(t),
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(provider)},
				Steps: []testresource.TestStep{
					{Config: config, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"a", "b", "c"}}}},
					{Config: config, PlanOnly: true},
					{Config: config, PreConfig: func() { client.set("c", "b", "a") }, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"a", "b", "c"}}}},
					{Config: updated, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"c", "b", "a"}}}},
					{Config: updated, PlanOnly: true},
					{Config: expanded},
					{Config: replaced, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"d", "c", "b", "new-a"}}}},
					{Config: shrunk, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"new-a", "b"}}}},
					{Config: shrunk, PlanOnly: true},
					{Config: `provider "ordering" {}`},
				},
			})
		})
	}
}

func TestPolicyOrderingTerraformGlobalAlert(t *testing.T) {
	client := &memoryOrderingClient{ids: []string{"b", "a"}}
	provider := &genericOrderingTestProvider{orderingTestProvider: &orderingTestProvider{client: client}, clients: map[policyType]*memoryOrderingClient{policyTypeAlert: client}}
	config := `resource "ordering_policy_ordering" "test" {
  policy_type = "alert"
  policies = [{policy_id="a",order=1},{policy_id="b",order=2}]
}`
	testresource.UnitTest(t, testresource.TestCase{
		WorkingDir:               orderingTerraformWorkingDir(t),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(provider)},
		Steps: []testresource.TestStep{
			{Config: config, Check: testresource.TestCheckResourceAttr("ordering_policy_ordering.test", "id", "global/alert")},
			{Config: config, ResourceName: "ordering_policy_ordering.test", ImportState: true, ImportStateId: "global/alert", ImportStateVerify: true},
			{Config: config, PlanOnly: true},
		},
	})
}

func TestPolicyOrderingTerraformTypeReplacement(t *testing.T) {
	alert := &memoryOrderingClient{ids: []string{"b", "a"}}
	notification := &memoryOrderingClient{ids: []string{"b", "a"}}
	provider := &genericOrderingTestProvider{orderingTestProvider: &orderingTestProvider{client: alert}, clients: map[policyType]*memoryOrderingClient{policyTypeAlert: alert, policyTypeNotification: notification}}
	config := `resource "ordering_policy_ordering" "test" {
 policy_type = "alert"
 team_id = "team"
 policies = [{policy_id="a",order=1},{policy_id="b",order=2}]
}`
	updated := strings.ReplaceAll(config, `"alert"`, `"notification"`)
	testresource.UnitTest(t, testresource.TestCase{
		WorkingDir:               orderingTerraformWorkingDir(t),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(provider)},
		Steps: []testresource.TestStep{
			{Config: config, Check: testresource.TestCheckResourceAttr("ordering_policy_ordering.test", "id", "team/team/alert")},
			{Config: config, ResourceName: "ordering_policy_ordering.test", ImportState: true, ImportStateId: "team/team/alert", ImportStateVerify: true},
			{Config: updated, Check: testresource.TestCheckResourceAttr("ordering_policy_ordering.test", "id", "team/team/notification")},
			{Config: updated, ResourceName: "ordering_policy_ordering.test", ImportState: true, ImportStateId: "team/team/notification", ImportStateVerify: true},
			{Config: updated, PlanOnly: true},
		},
	})
	if !reflect.DeepEqual(alert.ids, []string{"a", "b"}) || !reflect.DeepEqual(notification.ids, []string{"a", "b"}) {
		t.Fatal("wrong type routed or destroy altered policies")
	}
}

type scopedOrderingClient map[string]*memoryOrderingClient

func (c scopedOrderingClient) List(ctx context.Context, team string) ([]policyPosition, error) {
	return c[team].List(ctx, team)
}

func (c scopedOrderingClient) Move(ctx context.Context, team, id string, index int64) error {
	return c[team].Move(ctx, team, id, index)
}

func TestPolicyOrderingTerraformScopeReplacement(t *testing.T) {
	clients := scopedOrderingClient{
		"":       &memoryOrderingClient{ids: []string{"b", "a", "global-only"}},
		"first":  &memoryOrderingClient{ids: []string{"b", "a", "first-only"}},
		"second": &memoryOrderingClient{ids: []string{"b", "a", "second-only"}},
	}
	p := &orderingHTTPTestProvider{orderingTestProvider: &orderingTestProvider{}, api: clients}
	config := `resource "ordering_policy_ordering" "test" {
 policy_type = "alert"
 team_id = null
 policies = [{policy_id="a",order=1},{policy_id="b",order=2}]
}`
	first := strings.ReplaceAll(config, "team_id = null", `team_id = "first"`)
	second := strings.ReplaceAll(config, "team_id = null", `team_id = "second"`)
	testresource.UnitTest(t, testresource.TestCase{
		WorkingDir:               orderingTerraformWorkingDir(t),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(p)},
		Steps: []testresource.TestStep{
			{Config: config, Check: testresource.TestCheckResourceAttr("ordering_policy_ordering.test", "id", "global/alert")},
			{Config: first, Check: testresource.TestCheckResourceAttr("ordering_policy_ordering.test", "id", "team/first/alert")},
			{Config: second, Check: testresource.TestCheckResourceAttr("ordering_policy_ordering.test", "id", "team/second/alert")},
			{Config: second, PlanOnly: true},
			{Config: `provider "ordering" {}`},
		},
	})
	for scope, client := range clients {
		if client.moves != 2 || len(client.ids) != 3 || client.ids[0] != "a" || client.ids[1] != "b" {
			t.Fatalf("replacement or destroy altered old scope %q: moves=%d ids=%v", scope, client.moves, client.ids)
		}
	}
}

func TestPolicyOrderingTerraformNotificationPartialFailure(t *testing.T) {
	client := &memoryOrderingClient{ids: []string{"c", "b", "a"}, failMove: 2}
	provider := &genericOrderingTestProvider{orderingTestProvider: &orderingTestProvider{client: client}, clients: map[policyType]*memoryOrderingClient{policyTypeNotification: client}}
	config := `resource "ordering_policy_ordering" "test" {
 policy_type = "notification"
 team_id = "team"
 policies = [{policy_id="c",order=3},{policy_id="a",order=1},{policy_id="b",order=2}]
}`
	testresource.UnitTest(t, testresource.TestCase{
		WorkingDir:               orderingTerraformWorkingDir(t),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"ordering": providerserver.NewProtocol6WithError(provider)},
		Steps: []testresource.TestStep{
			{Config: config, ExpectError: regexp.MustCompile("Unable to Reconcile Policy Ordering")},
			{Config: config, PreConfig: func() { client.mu.Lock(); client.failMove = 0; client.mu.Unlock() }, ConfigStateChecks: []statecheck.StateCheck{orderingRemoteCheck{client: client, want: []string{"a", "b", "c"}}}},
			{Config: config, PlanOnly: true},
		},
	})
}
