package provider

import (
	"context"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMergeJsonExact(t *testing.T) {
	cases := []struct {
		name     string
		prior    jsontypes.Exact
		apiValue map[string]interface{}
		want     jsontypes.Exact
	}{
		{
			name:     "null prior and empty response stays null",
			prior:    jsontypes.NewExactNull(),
			apiValue: nil,
			want:     jsontypes.NewExactNull(),
		},
		{
			name:     "null prior takes the response",
			prior:    jsontypes.NewExactNull(),
			apiValue: map[string]interface{}{"a": "1"},
			want:     jsontypes.NewExactValue(`{"a":"1"}`),
		},
		{
			name:     "unknown prior takes the response",
			prior:    jsontypes.NewExactUnknown(),
			apiValue: map[string]interface{}{"a": "1"},
			want:     jsontypes.NewExactValue(`{"a":"1"}`),
		},
		{
			name:     "known prior survives an empty response",
			prior:    jsontypes.NewExactValue(`{"a":"1"}`),
			apiValue: nil,
			want:     jsontypes.NewExactValue(`{"a":"1"}`),
		},
		{
			name:     "prior kept when the response only adds keys",
			prior:    jsontypes.NewExactValue(`{"a":"1"}`),
			apiValue: map[string]interface{}{"a": "1", "b": ""},
			want:     jsontypes.NewExactValue(`{"a":"1"}`),
		},
		{
			name:     "prior kept when the response elides keys",
			prior:    jsontypes.NewExactValue(`{"a":"1","b":"2"}`),
			apiValue: map[string]interface{}{"a": "1"},
			want:     jsontypes.NewExactValue(`{"a":"1","b":"2"}`),
		},
		{
			name:     "response wins on a conflicting shared key",
			prior:    jsontypes.NewExactValue(`{"a":"1"}`),
			apiValue: map[string]interface{}{"a": "2"},
			want:     jsontypes.NewExactValue(`{"a":"2"}`),
		},
		{
			name:     "nested shared values compare deeply: equal keeps prior",
			prior:    jsontypes.NewExactValue(`{"responders":[{"id":"x","type":"team"}],"details":{}}`),
			apiValue: map[string]interface{}{"responders": []interface{}{map[string]interface{}{"id": "x", "type": "team"}}},
			want:     jsontypes.NewExactValue(`{"responders":[{"id":"x","type":"team"}],"details":{}}`),
		},
		{
			name:     "nested shared values compare deeply: change wins",
			prior:    jsontypes.NewExactValue(`{"responders":[{"id":"x","type":"team"}]}`),
			apiValue: map[string]interface{}{"responders": []interface{}{map[string]interface{}{"id": "y", "type": "team"}}},
			want:     jsontypes.NewExactValue(`{"responders":[{"id":"y","type":"team"}]}`),
		},
		{
			// Older provider versions wrote "{}" and "null" strings to state;
			// they have no shared keys with any response, so they are kept and
			// plans stay converged for existing deployments.
			name:     "legacy empty-object prior is kept",
			prior:    jsontypes.NewExactValue(`{}`),
			apiValue: map[string]interface{}{"a": "1"},
			want:     jsontypes.NewExactValue(`{}`),
		},
		{
			name:     "legacy null-literal prior is kept",
			prior:    jsontypes.NewExactValue(`null`),
			apiValue: map[string]interface{}{"a": "1"},
			want:     jsontypes.NewExactValue(`null`),
		},
		{
			name:     "non-object prior falls back to the response",
			prior:    jsontypes.NewExactValue(`["not","an","object"]`),
			apiValue: map[string]interface{}{"a": "1"},
			want:     jsontypes.NewExactValue(`{"a":"1"}`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeJsonExact(tc.prior, tc.apiValue)
			if !got.Equal(tc.want) {
				t.Errorf("mergeJsonExact() = %s, want %s", got, tc.want)
			}
		})
	}
}

// advanced is server-managed: an unknown or null plan value must stay out of
// the request body, while a known value is sent explicitly.
func TestApiIntegrationModelToDto_AdvancedPointer(t *testing.T) {
	cases := []struct {
		name  string
		value types.Bool
		want  *bool
	}{
		{name: "unknown is omitted", value: types.BoolUnknown(), want: nil},
		{name: "null is omitted", value: types.BoolNull(), want: nil},
		{name: "false is sent", value: types.BoolValue(false), want: boolPtr(false)},
		{name: "true is sent", value: types.BoolValue(true), want: boolPtr(true)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := dataModels.ApiIntegrationModel{
				Name:                   types.StringValue("x"),
				Type:                   types.StringValue("API"),
				Advanced:               tc.value,
				MaintenanceSources:     types.ListNull(types.ObjectType{AttrTypes: dataModels.IntegrationMaintenanceSourcesResponseModelMap}),
				Directions:             types.ListNull(types.StringType),
				Domains:                types.ListNull(types.StringType),
				TypeSpecificProperties: jsontypes.NewExactNull(),
			}
			got := ApiIntegrationModelToDto(context.Background(), model).Advanced
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("expected advanced to be omitted, got %v", *got)
			case tc.want != nil && (got == nil || *got != *tc.want):
				t.Errorf("expected advanced %v, got %v", *tc.want, got)
			}
		})
	}
}
