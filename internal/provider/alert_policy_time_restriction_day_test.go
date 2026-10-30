package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
)

var timeRestrictionPeriodAttrTypes = map[string]attr.Type{
	"start_day":    types.Int64Type,
	"start_hour":   types.Int64Type,
	"start_minute": types.Int64Type,
	"end_day":      types.Int64Type,
	"end_hour":     types.Int64Type,
	"end_minute":   types.Int64Type,
}

func buildAlertPolicyModelWithPeriod(period map[string]attr.Value) *dataModels.AlertPolicyModel {
	periodObj := types.ObjectValueMust(timeRestrictionPeriodAttrTypes, period)
	timeRestriction := types.ObjectValueMust(
		map[string]attr.Type{
			"enabled":           types.BoolType,
			"time_restrictions": types.ListType{ElemType: types.ObjectType{AttrTypes: timeRestrictionPeriodAttrTypes}},
		},
		map[string]attr.Value{
			"enabled": types.BoolValue(true),
			"time_restrictions": types.ListValueMust(
				types.ObjectType{AttrTypes: timeRestrictionPeriodAttrTypes},
				[]attr.Value{periodObj},
			),
		},
	)

	return &dataModels.AlertPolicyModel{
		ID:              types.StringValue("policy-id"),
		Type:            types.StringValue("alert"),
		Name:            types.StringValue("test-policy"),
		Enabled:         types.BoolValue(true),
		TimeRestriction: timeRestriction,
		Responders:      types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"type": types.StringType, "id": types.StringType}}),
		Actions:         types.ListNull(types.StringType),
		Tags:            types.SetNull(types.StringType),
		Details:         types.MapNull(types.StringType),
	}
}

// TestAlertPolicyTimeRestrictionDay_RoundTrip exercises the start_day/end_day
// fields added to alert_policy time restriction periods (mirrors ROH-751 for
// notification_policy): model -> DTO -> model must preserve the values.
func TestAlertPolicyTimeRestrictionDay_RoundTrip(t *testing.T) {
	model := buildAlertPolicyModelWithPeriod(map[string]attr.Value{
		"start_day":    types.Int64Value(1),
		"start_hour":   types.Int64Value(9),
		"start_minute": types.Int64Value(0),
		"end_day":      types.Int64Value(5),
		"end_hour":     types.Int64Value(17),
		"end_minute":   types.Int64Value(0),
	})

	policyDto, diags := AlertPolicyModelToDto(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("AlertPolicyModelToDto returned errors: %v", diags)
	}

	if policyDto.TimeRestriction == nil || len(policyDto.TimeRestriction.TimeRestrictions) != 1 {
		t.Fatalf("expected one time restriction period in DTO, got %+v", policyDto.TimeRestriction)
	}

	period := policyDto.TimeRestriction.TimeRestrictions[0]
	if period.StartDay == nil || *period.StartDay != 1 {
		t.Fatalf("expected StartDay=1, got %v", period.StartDay)
	}
	if period.EndDay == nil || *period.EndDay != 5 {
		t.Fatalf("expected EndDay=5, got %v", period.EndDay)
	}

	roundTripped, err := AlertPolicyDtoToModel(context.Background(), 0, policyDto)
	if err != nil {
		t.Fatalf("AlertPolicyDtoToModel returned error: %v", err)
	}

	var trModel dataModels.AlertTimeRestrictionModel
	diags = roundTripped.TimeRestriction.As(context.Background(), &trModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("failed to extract time restriction model: %v", diags)
	}

	var periods []dataModels.AlertTimeRestrictionPeriodModel
	diags = trModel.TimeRestrictions.ElementsAs(context.Background(), &periods, false)
	if diags.HasError() {
		t.Fatalf("failed to extract time restriction periods: %v", diags)
	}
	if len(periods) != 1 {
		t.Fatalf("expected one period after round trip, got %d", len(periods))
	}

	if periods[0].StartDay.ValueInt64() != 1 {
		t.Errorf("expected start_day=1 after round trip, got %d", periods[0].StartDay.ValueInt64())
	}
	if periods[0].EndDay.ValueInt64() != 5 {
		t.Errorf("expected end_day=5 after round trip, got %d", periods[0].EndDay.ValueInt64())
	}
}

// TestAlertPolicyTimeRestrictionDay_OptionalWhenOmitted ensures start_day/end_day
// remain optional: a period without them must still convert cleanly and the DTO
// side must omit the day fields (nil pointer) rather than defaulting to 0.
func TestAlertPolicyTimeRestrictionDay_OptionalWhenOmitted(t *testing.T) {
	model := buildAlertPolicyModelWithPeriod(map[string]attr.Value{
		"start_day":    types.Int64Null(),
		"start_hour":   types.Int64Value(0),
		"start_minute": types.Int64Value(0),
		"end_day":      types.Int64Null(),
		"end_hour":     types.Int64Value(23),
		"end_minute":   types.Int64Value(59),
	})

	policyDto, diags := AlertPolicyModelToDto(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("AlertPolicyModelToDto returned errors: %v", diags)
	}

	period := policyDto.TimeRestriction.TimeRestrictions[0]
	if period.StartDay != nil {
		t.Errorf("expected nil StartDay when omitted, got %v", *period.StartDay)
	}
	if period.EndDay != nil {
		t.Errorf("expected nil EndDay when omitted, got %v", *period.EndDay)
	}

	roundTripped, err := AlertPolicyDtoToModel(context.Background(), 0, policyDto)
	if err != nil {
		t.Fatalf("AlertPolicyDtoToModel returned error: %v", err)
	}

	var trModel dataModels.AlertTimeRestrictionModel
	diags = roundTripped.TimeRestriction.As(context.Background(), &trModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("failed to extract time restriction model: %v", diags)
	}

	var periods []dataModels.AlertTimeRestrictionPeriodModel
	diags = trModel.TimeRestrictions.ElementsAs(context.Background(), &periods, false)
	if diags.HasError() {
		t.Fatalf("failed to extract time restriction periods: %v", diags)
	}

	if !periods[0].StartDay.IsNull() {
		t.Errorf("expected start_day to remain null after round trip, got %v", periods[0].StartDay)
	}
	if !periods[0].EndDay.IsNull() {
		t.Errorf("expected end_day to remain null after round trip, got %v", periods[0].EndDay)
	}
}

// TestAlertPolicyTimeRestrictionDay_ValidatorRejectsOutOfRange is an adversarial
// check on the 0-7 boundary declared for start_day/end_day: values outside that
// range must be rejected by the schema validators actually wired up, and the
// boundary values themselves (0 and 7) must be accepted.
func TestAlertPolicyTimeRestrictionDay_ValidatorRejectsOutOfRange(t *testing.T) {
	timeRestriction, ok := schemaAttributes.AlertPolicyResourceAttributes["time_restriction"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("time_restriction attribute has unexpected type %T", schemaAttributes.AlertPolicyResourceAttributes["time_restriction"])
	}
	timeRestrictions, ok := timeRestriction.Attributes["time_restrictions"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("time_restrictions attribute has unexpected type %T", timeRestriction.Attributes["time_restrictions"])
	}
	for _, attrName := range []string{"start_day", "end_day"} {
		t.Run(attrName, func(t *testing.T) {
			dayAttr, ok := timeRestrictions.NestedObject.Attributes[attrName].(schema.Int64Attribute)
			if !ok {
				t.Fatalf("%s attribute has unexpected type %T", attrName, timeRestrictions.NestedObject.Attributes[attrName])
			}

			runValidators := func(value int64) diag.Diagnostics {
				req := validator.Int64Request{
					Path:        path.Root(attrName),
					ConfigValue: types.Int64Value(value),
				}
				resp := &validator.Int64Response{}
				for _, v := range dayAttr.Validators {
					v.ValidateInt64(context.Background(), req, resp)
				}
				return resp.Diagnostics
			}

			for _, tc := range []struct {
				name    string
				value   int64
				wantErr bool
			}{
				{"below range", -1, true},
				{"above range", 8, true},
				{"lower boundary", 0, false},
				{"upper boundary", 7, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					diags := runValidators(tc.value)
					if diags.HasError() != tc.wantErr {
						t.Errorf("value=%d: expected HasError=%v, got diagnostics=%v", tc.value, tc.wantErr, diags)
					}
				})
			}
		})
	}
}
