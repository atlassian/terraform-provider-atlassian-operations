package customPlanModifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = &stringNullIfBoolFieldFalseModifier{}

type stringNullIfBoolFieldFalseModifier struct {
	boolField path.Path
}

// PlanModifyString only acts when the attribute is not set in config and the framework has marked it unknown:
//   - if the bool field is planned as false, the attribute is planned as null, because the API does not return it then;
//   - otherwise it behaves like UseStateForUnknown: the prior state value is kept unless it is null.
func (m stringNullIfBoolFieldFalseModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || !req.PlanValue.IsUnknown() {
		return
	}

	var boolFieldValue types.Bool
	diags := req.Plan.GetAttribute(ctx, m.boolField, &boolFieldValue)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if !boolFieldValue.IsNull() && !boolFieldValue.IsUnknown() && !boolFieldValue.ValueBool() {
		resp.PlanValue = types.StringNull()
		return
	}

	if req.StateValue.IsNull() {
		return
	}

	resp.PlanValue = req.StateValue
}

func (m stringNullIfBoolFieldFalseModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Plans null when '%s' is false and the value is not configured; otherwise uses the prior state value for unknown plans.", m.boolField)
}

func (m stringNullIfBoolFieldFalseModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// StringNullIfBoolFieldFalse returns a plan modifier for an Optional+Computed string attribute that is only
// returned by the API while boolField is true. See PlanModifyString for the exact behaviour.
func StringNullIfBoolFieldFalse(boolField path.Path) planmodifier.String {
	return &stringNullIfBoolFieldFalseModifier{
		boolField: boolField,
	}
}
