package customPlanModifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStringNullIfBoolFieldFalse(t *testing.T) {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"update_priority": schema.BoolAttribute{Optional: true, Computed: true},
			"priority_value":  schema.StringAttribute{Optional: true, Computed: true},
		},
	}
	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"update_priority": tftypes.Bool,
		"priority_value":  tftypes.String,
	}}
	planWith := func(updatePriority tftypes.Value) tfsdk.Plan {
		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
				"update_priority": updatePriority,
				"priority_value":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		}
	}
	boolFalse := tftypes.NewValue(tftypes.Bool, false)
	boolTrue := tftypes.NewValue(tftypes.Bool, true)
	boolUnknown := tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)

	testCases := map[string]struct {
		updatePriority tftypes.Value
		config         types.String
		plan           types.String
		state          types.String
		expected       types.String
	}{
		"configured value is kept": {
			updatePriority: boolTrue,
			config:         types.StringValue("P3"),
			plan:           types.StringValue("P3"),
			state:          types.StringValue("P2"),
			expected:       types.StringValue("P3"),
		},
		"known plan value is kept": {
			updatePriority: boolFalse,
			config:         types.StringNull(),
			plan:           types.StringNull(),
			state:          types.StringNull(),
			expected:       types.StringNull(),
		},
		"update_priority false with null state plans null": {
			updatePriority: boolFalse,
			config:         types.StringNull(),
			plan:           types.StringUnknown(),
			state:          types.StringNull(),
			expected:       types.StringNull(),
		},
		"switching update_priority to false plans null instead of the prior value": {
			updatePriority: boolFalse,
			config:         types.StringNull(),
			plan:           types.StringUnknown(),
			state:          types.StringValue("P2"),
			expected:       types.StringNull(),
		},
		"update_priority true keeps the prior state value": {
			updatePriority: boolTrue,
			config:         types.StringNull(),
			plan:           types.StringUnknown(),
			state:          types.StringValue("P2"),
			expected:       types.StringValue("P2"),
		},
		"update_priority true with null state stays unknown": {
			updatePriority: boolTrue,
			config:         types.StringNull(),
			plan:           types.StringUnknown(),
			state:          types.StringNull(),
			expected:       types.StringUnknown(),
		},
		"unknown update_priority keeps the prior state value": {
			updatePriority: boolUnknown,
			config:         types.StringNull(),
			plan:           types.StringUnknown(),
			state:          types.StringValue("P2"),
			expected:       types.StringValue("P2"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				Path:        path.Root("priority_value"),
				Plan:        planWith(testCase.updatePriority),
				ConfigValue: testCase.config,
				PlanValue:   testCase.plan,
				StateValue:  testCase.state,
			}
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}

			StringNullIfBoolFieldFalse(path.Root("update_priority")).PlanModifyString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if !resp.PlanValue.Equal(testCase.expected) {
				t.Errorf("expected %s, got %s", testCase.expected, resp.PlanValue)
			}
		})
	}
}
