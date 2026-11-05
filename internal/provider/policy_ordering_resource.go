package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &PolicyOrderingResource{}
	_ resource.ResourceWithConfigure      = &PolicyOrderingResource{}
	_ resource.ResourceWithModifyPlan     = &PolicyOrderingResource{}
	_ resource.ResourceWithImportState    = &PolicyOrderingResource{}
	_ resource.ResourceWithValidateConfig = &PolicyOrderingResource{}
)

type PolicyOrderingResource struct {
	client        policyOrderingClient
	clientFactory func(policyType) policyOrderingClient
}

func NewPolicyOrderingResource() resource.Resource {
	return &PolicyOrderingResource{}
}

func (r *PolicyOrderingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_ordering"
}

func (r *PolicyOrderingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ordered prefix of policies after referenced policies are created. Alert policies support team/global scopes; notification policies require a team. Use one ordering owner per scope and policy type, without explicit per-policy order. Moves are sequential, verified, and non-atomic. Unlisted policies are retained. Destroy leaves policies and their order intact. Reconciles when scheduled; unchanged existing policy-resource behavior and competing writers may require a subsequent full plan/apply.",
		Attributes:  schemaAttributes.PolicyOrderingResourceAttributes,
	}
}

func (r *PolicyOrderingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	configuration, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected dto.AtlassianOpsProviderModel, got %T.", req.ProviderData))
		return
	}
	r.clientFactory = func(pt policyType) policyOrderingClient { return newPolicyOrderingClient(configuration, pt) }
}

// A computed observation alone does not schedule an update. Mark it unknown
// when refreshed remote order differs from the desired sequence so apply must
// compute it by reconciliation. Never reorder or change configured policies.
func (r *PolicyOrderingResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var planned, state dataModels.PolicyOrderingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planned)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired, known, diagnostics := orderingSequence(ctx, planned.Policies)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() || !known {
		return
	}
	if !orderingIDsValue(desired).Equal(state.ObservedOrder) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("observed_order"), types.ListUnknown(types.StringType))...)
	}
}

func (r *PolicyOrderingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.apply(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *PolicyOrderingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.apply(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *PolicyOrderingResource) apply(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	// Bound the whole sequential operation, including HTTP retries. A timeout
	// is a partial failure, not permission to fabricate successful order.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	var data dataModels.PolicyOrderingModel
	diagnostics.Append(plan.Get(ctx, &data)...)
	if diagnostics.HasError() {
		return
	}
	teamID, pt, desired, err := policyOrderingInputs(ctx, data)
	if err != nil {
		diagnostics.AddError("Invalid Policy Ordering", err.Error())
		return
	}
	client := r.clientForType(pt)
	if client == nil {
		diagnostics.AddError("Unconfigured Policy Ordering", "The provider client is not configured.")
		return
	}
	positions, err := client.List(ctx, teamID)
	if err != nil {
		diagnostics.AddError("Unable to Read Policy Ordering", err.Error())
		return
	}
	present := make(map[string]bool, len(positions))
	for _, p := range positions {
		present[p.ID] = true
	}
	for _, id := range desired {
		if !present[id] {
			diagnostics.AddError("Policy Not Found", fmt.Sprintf("Policy %q is not in the requested policy type and scope. No policies were moved. Reference policy resource IDs and verify their team/global scope.", id))
			return
		}
	}
	data.ID = types.StringValue(policyOrderingIdentity(teamID, pt))
	data.ObservedOrder = orderingPrefixValue(positions, len(desired))
	// Persist the controller identity and truthful initial observation before
	// side effects. A failed create can then be recovered/tainted by Terraform.
	diagnostics.Append(state.Set(ctx, &data)...)
	if diagnostics.HasError() {
		return
	}
	if !slices.Equal(orderingPrefix(positions, len(desired)), desired) {
		for index, id := range desired {
			if err := client.Move(ctx, teamID, id, int64(index)); err != nil {
				r.recordFailure(ctx, client, state, diagnostics, &data, teamID, len(desired), fmt.Errorf("moving policy %q to position %d: %w", id, index+1, err))
				return
			}
		}
		// Read polling tolerates visibility lag, but does not repeatedly move
		// policies or fight another ordering owner indefinitely.
		for attempt := 0; attempt < 4; attempt++ {
			positions, err = client.List(ctx, teamID)
			if err != nil {
				r.recordFailure(ctx, client, state, diagnostics, &data, teamID, len(desired), err)
				return
			}
			data.ObservedOrder = orderingPrefixValue(positions, len(desired))
			if slices.Equal(orderingPrefix(positions, len(desired)), desired) {
				break
			}
			if attempt == 3 {
				r.recordFailure(ctx, client, state, diagnostics, &data, teamID, len(desired), fmt.Errorf("requested prefix %v, observed %v", desired, orderingPrefix(positions, len(desired))))
				return
			}
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				r.recordFailure(ctx, client, state, diagnostics, &data, teamID, len(desired), ctx.Err())
				return
			case <-timer.C:
			}
		}
	}
	diagnostics.Append(state.Set(ctx, &data)...)
}

func (r *PolicyOrderingResource) recordFailure(ctx context.Context, client policyOrderingClient, state *tfsdk.State, diagnostics *diag.Diagnostics, data *dataModels.PolicyOrderingModel, teamID string, count int, cause error) {
	// Avoid claiming that an earlier snapshot is the post-failure order if a
	// fresh observation cannot be obtained. A subsequent Read can recover it.
	data.ObservedOrder = types.ListNull(types.StringType)
	if positions, err := client.List(ctx, teamID); err == nil {
		data.ObservedOrder = orderingPrefixValue(positions, count)
	}
	diagnostics.Append(state.Set(ctx, data)...)
	diagnostics.AddError("Unable to Reconcile Policy Ordering", fmt.Sprintf("%s. Some moves may already have succeeded. Check for competing ordering owners and run a full plan/apply to reconcile again; no policies were deleted.", cause))
}

func (r *PolicyOrderingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.PolicyOrderingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	teamID, pt, desired, err := policyOrderingInputs(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Policy Ordering State", err.Error())
		return
	}
	client := r.clientForType(pt)
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured Policy Ordering", "The provider client is not configured.")
		return
	}
	positions, err := client.List(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Policy Ordering", err.Error())
		return
	}
	// Preserve desired IDs even when a policy is missing. Otherwise refresh
	// would silently relinquish ownership rather than expose drift.
	data.ObservedOrder = orderingPrefixValue(positions, len(desired))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyOrderingResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PolicyOrderingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	teamID, pt, err := parsePolicyOrderingIdentity(req.ID)
	if err != nil {
		message := "Use global/alert, team/<team_id>/alert, or team/<team_id>/notification. Import adopts the complete current sequence."
		resp.Diagnostics.AddError("Invalid Ordering Import ID", message)
		return
	}
	client := r.clientForType(pt)
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured Policy Ordering", "The provider client is not configured.")
		return
	}
	positions, err := client.List(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Import Policy Ordering", err.Error())
		return
	}
	data := dataModels.PolicyOrderingModel{
		ID: types.StringValue(req.ID), PolicyType: types.StringValue(string(pt)), TeamID: types.StringNull(),
		Policies:      orderingEntriesValue(orderingPrefix(positions, len(positions))),
		ObservedOrder: orderingPrefixValue(positions, len(positions)),
	}
	if teamID != "" {
		data.TeamID = types.StringValue(teamID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func parsePolicyOrderingIdentity(id string) (string, policyType, error) {
	if id == "global/alert" {
		return "", policyTypeAlert, nil
	}
	parts := strings.Split(id, "/")
	if len(parts) != 3 || parts[0] != "team" {
		return "", "", fmt.Errorf("invalid ordering identity")
	}
	pt := policyType(parts[2])
	if parts[1] == "" {
		return "", "", fmt.Errorf("team ID must not be empty")
	}
	if _, err := policyEndpoint(pt, parts[1]); err != nil {
		return "", "", err
	}
	return parts[1], pt, nil
}

func policyOrderingIdentity(teamID string, pt policyType) string {
	if teamID == "" {
		return "global/" + string(pt)
	}
	return "team/" + teamID + "/" + string(pt)
}

func policyOrderingInputs(ctx context.Context, data dataModels.PolicyOrderingModel) (string, policyType, []string, error) {
	if data.PolicyType.IsNull() || data.PolicyType.IsUnknown() {
		return "", "", nil, fmt.Errorf("policy_type must be known before ordering")
	}
	if data.TeamID.IsUnknown() {
		return "", "", nil, fmt.Errorf("team_id must be known before ordering; unknown is not global scope")
	}
	teamID, pt := data.TeamID.ValueString(), policyType(data.PolicyType.ValueString())
	if !data.TeamID.IsNull() {
		if err := validatePolicyPathID(teamID); err != nil {
			return "", "", nil, err
		}
	}
	if _, err := policyEndpoint(pt, teamID); err != nil {
		return "", "", nil, err
	}
	ids, known, diagnostics := orderingSequence(ctx, data.Policies)
	if diagnostics.HasError() {
		return "", "", nil, fmt.Errorf("%s", diagnostics.Errors()[0].Detail())
	}
	if !known {
		return "", "", nil, fmt.Errorf("every policy_id and order must be known before ordering")
	}
	return teamID, pt, ids, nil
}

func (r *PolicyOrderingResource) clientForType(pt policyType) policyOrderingClient {
	if r.clientFactory != nil {
		return r.clientFactory(pt)
	}
	return r.client
}

func orderingPrefix(positions []policyPosition, count int) []string {
	ids := make([]string, 0, min(count, len(positions)))
	for i := 0; i < count && i < len(positions); i++ {
		ids = append(ids, positions[i].ID)
	}
	return ids
}

func orderingPrefixValue(positions []policyPosition, count int) types.List {
	ids := orderingPrefix(positions, count)
	value, _ := types.ListValueFrom(context.Background(), types.StringType, ids)
	return value
}
