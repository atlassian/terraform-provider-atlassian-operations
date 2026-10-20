package provider

import (
	"context"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// A known planned field_mappings value must survive apply even when the API
// merges default keys into it (JSDCLOUD-19687).
func TestIntegrationActionDtoToModel_KnownPlanValueWinsOverNormalizedResponse(t *testing.T) {
	configuredFieldMappings := `{"alias":"{{alias}}"}`
	plan := dataModels.IntegrationActionModel{
		ID:            types.StringUnknown(),
		IntegrationID: types.StringValue("integration-id"),
		Type:          types.StringValue("close"),
		Name:          types.StringValue("Close Alert"),
		Domain:        types.StringValue("alert"),
		Direction:     types.StringValue("incoming"),
		FieldMappings: jsontypes.NewExactValue(configuredFieldMappings),
	}

	apiResponse := &dto.IntegrationActionDto{
		ID:        "action-id",
		Type:      "close",
		Name:      "Close Alert",
		Domain:    "alert",
		Direction: "incoming",
		FieldMappings: map[string]interface{}{
			"alias": "{{alias}}",
			"note":  "", // merged in by the API
			"user":  "", // merged in by the API
		},
	}

	result, diags := IntegrationActionDtoToModel(context.Background(), apiResponse, plan.IntegrationID, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !result.FieldMappings.Equal(jsontypes.NewExactValue(configuredFieldMappings)) {
		t.Errorf("expected field_mappings to keep the planned value %s, got %s", configuredFieldMappings, result.FieldMappings)
	}
}

// Read must preserve prior state for attributes the GET response omits
// (JSDCLOUD-19666, JSDCLOUD-19687).
func TestIntegrationActionDtoToModel_PreservesPriorStateWhenResponseOmitsFields(t *testing.T) {
	priorFilter := types.ObjectValueMust(
		dataModels.FilterModelMap,
		map[string]attr.Value{
			"conditions_empty":     types.BoolValue(false),
			"condition_match_type": types.StringValue("match-all"),
			"conditions":           types.ListValueMust(types.ObjectType{AttrTypes: dataModels.FilterConditionModelMap}, []attr.Value{}),
		},
	)
	priorActionMapping := types.ObjectValueMust(
		dataModels.ActionMappingModelMap,
		map[string]attr.Value{
			"type":      types.StringValue("acknowledge"),
			"parameter": jsontypes.NewExactValue(`{}`),
		},
	)
	priorState := dataModels.IntegrationActionModel{
		ID:                     types.StringValue("action-id"),
		IntegrationID:          types.StringValue("integration-id"),
		Type:                   types.StringValue("acknowledge"),
		Name:                   types.StringValue("Acknowledge Alert"),
		Domain:                 types.StringValue("alert"),
		Direction:              types.StringValue("incoming"),
		Filter:                 priorFilter,
		ActionMapping:          priorActionMapping,
		FieldMappings:          jsontypes.NewExactValue(`{"alias":"{{alias}}"}`),
		TypeSpecificProperties: jsontypes.NewExactValue(`{}`),
	}

	apiResponse := &dto.IntegrationActionDto{
		ID:        "action-id",
		Type:      "acknowledge",
		Name:      "Acknowledge Alert",
		Domain:    "alert",
		Direction: "incoming",
		// All other attributes omitted by the API.
	}

	result, diags := IntegrationActionDtoToModel(context.Background(), apiResponse, priorState.IntegrationID, &priorState)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !result.Filter.Equal(priorFilter) {
		t.Errorf("expected filter to be preserved as %s, got %s", priorFilter, result.Filter)
	}
	if !result.ActionMapping.Equal(priorActionMapping) {
		t.Errorf("expected action_mapping to be preserved as %s, got %s", priorActionMapping, result.ActionMapping)
	}
	if !result.FieldMappings.Equal(priorState.FieldMappings) {
		t.Errorf("expected field_mappings to be preserved as %s, got %s", priorState.FieldMappings, result.FieldMappings)
	}
	if !result.TypeSpecificProperties.Equal(priorState.TypeSpecificProperties) {
		t.Errorf("expected type_specific_properties to be preserved as %s, got %s", priorState.TypeSpecificProperties, result.TypeSpecificProperties)
	}
}

// With an unknown prior value, the API response is the source of truth.
func TestIntegrationActionDtoToModel_UsesResponseWhenPriorUnknown(t *testing.T) {
	plan := dataModels.IntegrationActionModel{
		ID:                     types.StringUnknown(),
		IntegrationID:          types.StringValue("integration-id"),
		Type:                   types.StringValue("close"),
		Name:                   types.StringValue("Close Alert"),
		Domain:                 types.StringValue("alert"),
		Direction:              types.StringValue("incoming"),
		Filter:                 types.ObjectUnknown(dataModels.FilterModelMap),
		ActionMapping:          types.ObjectUnknown(dataModels.ActionMappingModelMap),
		FieldMappings:          jsontypes.NewExactUnknown(),
		TypeSpecificProperties: jsontypes.NewExactUnknown(),
	}

	apiResponse := &dto.IntegrationActionDto{
		ID:        "action-id",
		Type:      "close",
		Name:      "Close Alert",
		Domain:    "alert",
		Direction: "incoming",
		Filter: &dto.FilterDto{
			ConditionsEmpty:    false,
			ConditionMatchType: "match-all",
			Conditions:         []dto.FilterConditionDto{},
		},
		FieldMappings: map[string]interface{}{"alias": "{{alias}}", "note": "", "user": ""},
	}

	result, diags := IntegrationActionDtoToModel(context.Background(), apiResponse, plan.IntegrationID, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if result.Filter.IsNull() || result.Filter.IsUnknown() {
		t.Errorf("expected filter from the response, got %s", result.Filter)
	}
	if result.FieldMappings.IsNull() || result.FieldMappings.IsUnknown() {
		t.Errorf("expected field_mappings from the response, got %s", result.FieldMappings)
	}
	// Omitted attributes must still resolve to known (possibly null) values.
	if result.ActionMapping.IsUnknown() {
		t.Errorf("expected action_mapping to resolve to a known value, got %s", result.ActionMapping)
	}
	if result.TypeSpecificProperties.IsUnknown() {
		t.Errorf("expected type_specific_properties to resolve to a known value, got %s", result.TypeSpecificProperties)
	}
}

// Values the API response does report win over prior state, so genuine
// changes made on the Atlassian side surface as drift.
func TestIntegrationActionDtoToModel_ApiValueWinsOnGenuineChange(t *testing.T) {
	priorState := dataModels.IntegrationActionModel{
		ID:            types.StringValue("action-id"),
		IntegrationID: types.StringValue("integration-id"),
		Type:          types.StringValue("close"),
		Name:          types.StringValue("Close Alert"),
		Domain:        types.StringValue("alert"),
		Direction:     types.StringValue("incoming"),
		Filter: types.ObjectValueMust(
			dataModels.FilterModelMap,
			map[string]attr.Value{
				"conditions_empty":     types.BoolValue(false),
				"condition_match_type": types.StringValue("match-all"),
				"conditions":           types.ListValueMust(types.ObjectType{AttrTypes: dataModels.FilterConditionModelMap}, []attr.Value{}),
			},
		),
		FieldMappings: jsontypes.NewExactValue(`{"alias":"{{alias}}"}`),
	}

	apiResponse := &dto.IntegrationActionDto{
		ID:        "action-id",
		Type:      "close",
		Name:      "Close Alert",
		Domain:    "alert",
		Direction: "incoming",
		Filter: &dto.FilterDto{
			ConditionsEmpty:    false,
			ConditionMatchType: "match-any-condition", // changed on the Atlassian side
			Conditions:         []dto.FilterConditionDto{},
		},
		FieldMappings: map[string]interface{}{
			"alias": "{{alert.alias}}", // changed on the Atlassian side
		},
	}

	result, diags := IntegrationActionDtoToModel(context.Background(), apiResponse, priorState.IntegrationID, &priorState)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !result.Filter.Attributes()["condition_match_type"].Equal(types.StringValue("match-any-condition")) {
		t.Errorf("expected filter from the response, got %s", result.Filter)
	}
	if !result.FieldMappings.Equal(jsontypes.NewExactValue(`{"alias":"{{alert.alias}}"}`)) {
		t.Errorf("expected field_mappings from the response, got %s", result.FieldMappings)
	}
}

// A plan object can be known while still carrying unknown computed attributes
// (e.g. filter.conditions left unset in config). Such a prior must not be
// copied into applied state — unknown values are invalid after apply.
func TestIntegrationActionDtoToModel_DoesNotApplyPriorWithUnknownNestedValues(t *testing.T) {
	planFilter := types.ObjectValueMust(
		dataModels.FilterModelMap,
		map[string]attr.Value{
			"conditions_empty":     types.BoolValue(false),
			"condition_match_type": types.StringValue("match-all"),
			"conditions":           types.ListUnknown(types.ObjectType{AttrTypes: dataModels.FilterConditionModelMap}),
		},
	)
	plan := dataModels.IntegrationActionModel{
		ID:            types.StringUnknown(),
		IntegrationID: types.StringValue("integration-id"),
		Type:          types.StringValue("close"),
		Name:          types.StringValue("Close Alert"),
		Domain:        types.StringValue("alert"),
		Direction:     types.StringValue("incoming"),
		Filter:        planFilter,
	}

	apiResponse := &dto.IntegrationActionDto{
		ID:        "action-id",
		Type:      "close",
		Name:      "Close Alert",
		Domain:    "alert",
		Direction: "incoming",
		Filter: &dto.FilterDto{
			ConditionsEmpty:    false,
			ConditionMatchType: "match-all",
			Conditions:         []dto.FilterConditionDto{},
		},
	}

	result, diags := IntegrationActionDtoToModel(context.Background(), apiResponse, plan.IntegrationID, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if containsUnknownValues(result.Filter) {
		t.Errorf("expected filter without unknown values, got %s", result.Filter)
	}
	attrs := result.Filter.Attributes()
	if !attrs["condition_match_type"].Equal(types.StringValue("match-all")) {
		t.Errorf("expected condition_match_type from the response, got %s", attrs["condition_match_type"])
	}

	// Same scenario when the response omits the filter entirely: known leaves
	// are kept and the unknown attribute resolves to null.
	apiResponse.Filter = nil
	result, diags = IntegrationActionDtoToModel(context.Background(), apiResponse, plan.IntegrationID, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if containsUnknownValues(result.Filter) {
		t.Errorf("expected filter without unknown values, got %s", result.Filter)
	}
	attrs = result.Filter.Attributes()
	if !attrs["conditions_empty"].Equal(types.BoolValue(false)) {
		t.Errorf("expected conditions_empty to keep the planned value, got %s", attrs["conditions_empty"])
	}
	if !attrs["conditions"].IsNull() {
		t.Errorf("expected unknown conditions to resolve to null, got %s", attrs["conditions"])
	}
}

// Same unknown-value handling for action_mapping: a plan object with an
// unknown computed parameter must resolve to null when the response omits the
// mapping, keeping the known type.
func TestIntegrationActionDtoToModel_ResolvesUnknownActionMappingParameter(t *testing.T) {
	plan := dataModels.IntegrationActionModel{
		ID:            types.StringUnknown(),
		IntegrationID: types.StringValue("integration-id"),
		Type:          types.StringValue("close"),
		Name:          types.StringValue("Close Alert"),
		Domain:        types.StringValue("alert"),
		Direction:     types.StringValue("incoming"),
		ActionMapping: types.ObjectValueMust(
			dataModels.ActionMappingModelMap,
			map[string]attr.Value{
				"type":      types.StringValue("acknowledge"),
				"parameter": jsontypes.NewExactUnknown(),
			},
		),
	}

	apiResponse := &dto.IntegrationActionDto{
		ID:        "action-id",
		Type:      "close",
		Name:      "Close Alert",
		Domain:    "alert",
		Direction: "incoming",
		// ActionMapping omitted by the API.
	}

	result, diags := IntegrationActionDtoToModel(context.Background(), apiResponse, plan.IntegrationID, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if containsUnknownValues(result.ActionMapping) {
		t.Errorf("expected action_mapping without unknown values, got %s", result.ActionMapping)
	}
	attrs := result.ActionMapping.Attributes()
	if !attrs["type"].Equal(types.StringValue("acknowledge")) {
		t.Errorf("expected action_mapping type to keep the planned value, got %s", attrs["type"])
	}
	if !attrs["parameter"].IsNull() {
		t.Errorf("expected unknown parameter to resolve to null, got %s", attrs["parameter"])
	}
}

func TestContainsUnknownValues(t *testing.T) {
	conditionType := types.ObjectType{AttrTypes: dataModels.FilterConditionModelMap}
	knownCondition := types.ObjectValueMust(dataModels.FilterConditionModelMap, map[string]attr.Value{
		"field":            types.StringValue("message"),
		"operation":        types.StringValue("matches"),
		"expected_value":   types.StringValue("x"),
		"key":              types.StringValue(""),
		"not":              types.BoolValue(false),
		"order":            types.Int64Value(0),
		"system_condition": types.BoolValue(false),
	})
	unknownCondition := types.ObjectValueMust(dataModels.FilterConditionModelMap, map[string]attr.Value{
		"field":            types.StringValue("message"),
		"operation":        types.StringValue("matches"),
		"expected_value":   types.StringUnknown(),
		"key":              types.StringValue(""),
		"not":              types.BoolValue(false),
		"order":            types.Int64Value(0),
		"system_condition": types.BoolValue(false),
	})

	cases := []struct {
		name  string
		value attr.Value
		want  bool
	}{
		{name: "known primitive", value: types.StringValue("x"), want: false},
		{name: "unknown primitive", value: types.StringUnknown(), want: true},
		{name: "null collection", value: types.ListNull(conditionType), want: false},
		{name: "known nested list", value: types.ListValueMust(conditionType, []attr.Value{knownCondition}), want: false},
		{name: "unknown deep in a list element", value: types.ListValueMust(conditionType, []attr.Value{knownCondition, unknownCondition}), want: true},
		{name: "unknown in a set", value: types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()}), want: true},
		{name: "unknown in a map", value: types.MapValueMust(types.StringType, map[string]attr.Value{"k": types.StringUnknown()}), want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := containsUnknownValues(tc.value); got != tc.want {
				t.Errorf("containsUnknownValues() = %v, want %v", got, tc.want)
			}
		})
	}
}
