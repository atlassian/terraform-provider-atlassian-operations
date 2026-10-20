package provider

import (
	"encoding/json"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func boolPtr(b bool) *bool {
	return &b
}

// Read must preserve prior state for fields the GET response omits (JSDCLOUD-19666).
func TestApiIntegrationDtoToModel_PreservesPriorStateWhenResponseOmitsFields(t *testing.T) {
	priorState := dataModels.ApiIntegrationModel{
		Id:       types.StringValue("integration-id"),
		Name:     types.StringValue("prometheus"),
		ApiKey:   types.StringValue("secret-key"),
		Type:     types.StringValue("Prometheus"),
		Enabled:  types.BoolValue(true),
		TeamId:   types.StringValue("team-id"),
		Advanced: types.BoolValue(true),
		MaintenanceSources: types.ListValueMust(
			types.ObjectType{AttrTypes: dataModels.IntegrationMaintenanceSourcesResponseModelMap},
			[]attr.Value{},
		),
		Directions:             types.ListValueMust(types.StringType, []attr.Value{types.StringValue("incoming")}),
		Domains:                types.ListValueMust(types.StringType, []attr.Value{types.StringValue("alert")}),
		TypeSpecificProperties: jsontypes.NewExactValue(`{"suppressNotifications":false}`),
		DeleteDefaultActions:   types.BoolValue(false),
	}

	apiResponse := dto.ApiIntegration{
		Id:      "integration-id",
		Name:    "prometheus",
		Type:    "Prometheus",
		Enabled: true,
		TeamId:  "team-id",
		// All other fields omitted by the API.
	}

	result := ApiIntegrationDtoToModel(apiResponse, priorState)

	if !result.Advanced.Equal(priorState.Advanced) {
		t.Errorf("expected advanced to be preserved as %s, got %s", priorState.Advanced, result.Advanced)
	}
	if !result.ApiKey.Equal(priorState.ApiKey) {
		t.Errorf("expected api_key to be preserved as %s, got %s", priorState.ApiKey, result.ApiKey)
	}
	if !result.MaintenanceSources.Equal(priorState.MaintenanceSources) {
		t.Errorf("expected maintenance_sources to be preserved as %s, got %s", priorState.MaintenanceSources, result.MaintenanceSources)
	}
	if !result.Directions.Equal(priorState.Directions) {
		t.Errorf("expected directions to be preserved as %s, got %s", priorState.Directions, result.Directions)
	}
	if !result.Domains.Equal(priorState.Domains) {
		t.Errorf("expected domains to be preserved as %s, got %s", priorState.Domains, result.Domains)
	}
	if !result.TypeSpecificProperties.Equal(priorState.TypeSpecificProperties) {
		t.Errorf("expected type_specific_properties to be preserved as %s, got %s", priorState.TypeSpecificProperties, result.TypeSpecificProperties)
	}
}

// A known planned value must survive apply when the API merely merges default
// keys into it (JSDCLOUD-19666).
func TestApiIntegrationDtoToModel_KnownPlanValueWinsOverDefaultedResponse(t *testing.T) {
	configured := `{"suppressNotifications":false,"writeAccessGranted":true}`
	plan := dataModels.ApiIntegrationModel{
		Name:                   types.StringValue("sns"),
		Type:                   types.StringValue("AmazonSns"),
		TypeSpecificProperties: jsontypes.NewExactValue(configured),
	}

	apiResponse := dto.ApiIntegration{
		Id:      "integration-id",
		Name:    "sns",
		Type:    "AmazonSns",
		Enabled: true,
		TeamId:  "team-id",
		TypeSpecificProperties: map[string]interface{}{
			"suppressNotifications": false,
			"writeAccessGranted":    true,
			"isDefaultIntegration":  false, // added by the API
		},
	}

	result := ApiIntegrationDtoToModel(apiResponse, plan)

	if !result.TypeSpecificProperties.Equal(jsontypes.NewExactValue(configured)) {
		t.Errorf("expected type_specific_properties to keep the planned value %s, got %s", configured, result.TypeSpecificProperties)
	}
}

// When the API reports a different value for a key the prior also has, the
// value genuinely changed on the Atlassian side: the API copy wins so the
// drift is reported instead of silently discarded.
func TestApiIntegrationDtoToModel_ApiValueWinsOnConflictingSharedKey(t *testing.T) {
	prior := dataModels.ApiIntegrationModel{
		Name:                   types.StringValue("sns"),
		Type:                   types.StringValue("AmazonSns"),
		TypeSpecificProperties: jsontypes.NewExactValue(`{"readAccessGranted":false,"writeAccessGranted":true}`),
	}

	apiResponse := dto.ApiIntegration{
		Id:      "integration-id",
		Name:    "sns",
		Type:    "AmazonSns",
		Enabled: true,
		TeamId:  "team-id",
		TypeSpecificProperties: map[string]interface{}{
			"readAccessGranted":  true, // changed on the Atlassian side
			"writeAccessGranted": true,
		},
	}

	result := ApiIntegrationDtoToModel(apiResponse, prior)

	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result.TypeSpecificProperties.ValueString()), &resultMap); err != nil {
		t.Fatalf("unexpected type_specific_properties %s: %v", result.TypeSpecificProperties, err)
	}
	if resultMap["readAccessGranted"] != true {
		t.Errorf("expected readAccessGranted from the API response, got %s", result.TypeSpecificProperties)
	}
}

// With an unknown prior value, the API response is the source of truth.
func TestApiIntegrationDtoToModel_UsesResponseWhenPriorUnknown(t *testing.T) {
	plan := dataModels.ApiIntegrationModel{
		Name:                   types.StringValue("prometheus"),
		Type:                   types.StringValue("Prometheus"),
		Advanced:               types.BoolUnknown(),
		TypeSpecificProperties: jsontypes.NewExactUnknown(),
		MaintenanceSources:     types.ListUnknown(types.ObjectType{AttrTypes: dataModels.IntegrationMaintenanceSourcesResponseModelMap}),
		Directions:             types.ListUnknown(types.StringType),
		Domains:                types.ListUnknown(types.StringType),
	}

	apiResponse := dto.ApiIntegration{
		Id:                     "integration-id",
		Name:                   "prometheus",
		Type:                   "Prometheus",
		Advanced:               boolPtr(true),
		Directions:             []string{"incoming"},
		Domains:                []string{"alert"},
		TypeSpecificProperties: map[string]interface{}{"suppressNotifications": false},
	}

	result := ApiIntegrationDtoToModel(apiResponse, plan)

	if !result.Advanced.Equal(types.BoolValue(true)) {
		t.Errorf("expected advanced to be true, got %s", result.Advanced)
	}
	if !result.Directions.Equal(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("incoming")})) {
		t.Errorf("expected directions from the response, got %s", result.Directions)
	}
	if !result.Domains.Equal(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("alert")})) {
		t.Errorf("expected domains from the response, got %s", result.Domains)
	}
	if result.TypeSpecificProperties.IsNull() || result.TypeSpecificProperties.IsUnknown() {
		t.Errorf("expected type_specific_properties from the response, got %s", result.TypeSpecificProperties)
	}
	// Omitted fields must still resolve to known values.
	if result.MaintenanceSources.IsNull() || result.MaintenanceSources.IsUnknown() {
		t.Errorf("expected maintenance_sources to resolve to a known value, got %s", result.MaintenanceSources)
	}
}
