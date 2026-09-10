package provider

import (
	"context"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const testCloudId = "11111111-2222-3333-4444-555555555555"

func TestTeamIdToAri(t *testing.T) {
	tests := []struct {
		name     string
		teamId   string
		expected string
	}{
		{
			name:     "platform team id becomes an identity ARI",
			teamId:   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			expected: "ari:cloud:identity::team/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
		{
			name:     "opsgenie-native team id becomes an opsgenie ARI",
			teamId:   "og-99999999-8888-7777-6666-555555555555",
			expected: "ari:cloud:opsgenie:" + testCloudId + ":team/og-99999999-8888-7777-6666-555555555555",
		},
		{
			name:     "an ARI is left untouched",
			teamId:   "ari:cloud:identity::team/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			expected: "ari:cloud:identity::team/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
		{
			name:     "the empty owner used to clear an owner is preserved",
			teamId:   "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := teamIdToAri(test.teamId, testCloudId); got != test.expected {
				t.Errorf("teamIdToAri(%q) = %q, want %q", test.teamId, got, test.expected)
			}
		})
	}
}

func TestTeamAriToId(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "identity ARI reduces to the team id",
			value:    "ari:cloud:identity::team/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			expected: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
		{
			name:     "opsgenie ARI reduces to the team id",
			value:    "ari:cloud:opsgenie:" + testCloudId + ":team/og-99999999-8888-7777-6666-555555555555",
			expected: "og-99999999-8888-7777-6666-555555555555",
		},
		{
			name:     "a bare team id, as the API reports it today, is left untouched",
			value:    "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			expected: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
		{
			name:     "empty stays empty",
			value:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := teamAriToId(test.value); got != test.expected {
				t.Errorf("teamAriToId(%q) = %q, want %q", test.value, got, test.expected)
			}
		})
	}
}

// The API accepts ARIs on write but reports bare team ids on read. State must
// come back as the bare team id the practitioner configured either way,
// otherwise the apply fails with "Provider produced inconsistent result after
// apply".
func TestServiceTeamReferencesRoundTrip(t *testing.T) {
	ctx := context.Background()
	teamId := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	teamAri := "ari:cloud:identity::team/" + teamId

	model := &dataModels.ServiceModel{
		Name:        types.StringValue("example-service"),
		Description: types.StringValue("example service"),
		Tier:        types.Int32Value(2),
		Type:        types.StringValue("SOFTWARE_SERVICES"),
		Owner:       types.StringValue(teamId),
		Responders: types.ObjectValueMust(
			dataModels.RespondersModelMap,
			map[string]attr.Value{
				"users": types.ListValueMust(types.StringType, []attr.Value{}),
				"teams": types.ListValueMust(types.StringType, []attr.Value{types.StringValue(teamId)}),
			},
		),
	}

	serviceDto, diags := ServiceModelToDto(ctx, model, testCloudId)
	if diags.HasError() {
		t.Fatalf("ServiceModelToDto returned errors: %v", diags.Errors())
	}

	if serviceDto.Owner != teamAri {
		t.Errorf("owner sent to the API = %q, want the ARI %q", serviceDto.Owner, teamAri)
	}
	if serviceDto.Responders == nil || len(serviceDto.Responders.Teams) != 1 || serviceDto.Responders.Teams[0] != teamAri {
		t.Errorf("responder teams sent to the API = %v, want [%q]", serviceDto.Responders, teamAri)
	}

	// What the API reports back: the same service, with bare team ids.
	response := &dto.ServiceDto{
		ID:          "b:some-service-id",
		Name:        serviceDto.Name,
		Description: serviceDto.Description,
		Tier:        serviceDto.Tier,
		Type:        serviceDto.Type,
		Owner:       teamId,
		Responders:  &dto.RespondersDto{Teams: []string{teamId}},
	}

	roundTripped, diags := ServiceDtoToModel(ctx, response)
	if diags.HasError() {
		t.Fatalf("ServiceDtoToModel returned errors: %v", diags.Errors())
	}

	if roundTripped.Owner.ValueString() != teamId {
		t.Errorf("owner stored in state = %q, want the configured team id %q", roundTripped.Owner.ValueString(), teamId)
	}

	var responders dataModels.RespondersModel
	if diags := roundTripped.Responders.As(ctx, &responders, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("reading responders: %v", diags.Errors())
	}
	var teams []string
	if diags := responders.Teams.ElementsAs(ctx, &teams, false); diags.HasError() {
		t.Fatalf("reading responder teams: %v", diags.Errors())
	}
	if len(teams) != 1 || teams[0] != teamId {
		t.Errorf("responder teams stored in state = %v, want [%q]", teams, teamId)
	}
}
