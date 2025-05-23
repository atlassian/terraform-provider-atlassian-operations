package customValidators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = &listFieldNullIfOtherFieldValidator{}

type listFieldNullIfOtherFieldValidator struct {
	targetField  path.Expression
	fieldToCheck path.Expression
	checkValue   string
}

func (s listFieldNullIfOtherFieldValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	targetFieldExpressions := request.PathExpression.MergeExpressions(s.targetField)
	fieldToCheckExpressions := request.PathExpression.MergeExpressions(s.fieldToCheck)

	var targetFieldValue types.List
	var fieldToCheckValue types.String

	for _, expression := range targetFieldExpressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, matchedPath := range matchedPaths {
			var matchedPathValue attr.Value
			diags := request.Config.GetAttribute(ctx, matchedPath, &matchedPathValue)

			response.Diagnostics.Append(diags...)

			if diags.HasError() {
				continue
			}

			if matchedPathValue.IsNull() {
				continue
			}

			diags = tfsdk.ValueAs(ctx, matchedPathValue, &targetFieldValue)

			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}
		}
	}

	for _, expression := range fieldToCheckExpressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, matchedPath := range matchedPaths {
			var matchedPathValue attr.Value
			diags := request.Config.GetAttribute(ctx, matchedPath, &matchedPathValue)

			response.Diagnostics.Append(diags...)

			if diags.HasError() {
				continue
			}

			if matchedPathValue.IsNull() || matchedPathValue.IsUnknown() {
				continue
			}

			diags = tfsdk.ValueAs(ctx, matchedPathValue, &fieldToCheckValue)

			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}
		}
	}

	if (!targetFieldValue.IsNull()) &&
		(!fieldToCheckValue.IsNull() && !fieldToCheckValue.IsUnknown()) &&
		fieldToCheckValue.ValueString() == s.checkValue {
		response.Diagnostics.AddError("Invalid Attribute", fmt.Sprintf("The field '%s' must be null or unknown if the field '%s' is set to '%s'", s.targetField, s.fieldToCheck, s.checkValue))
	}
}

func (s listFieldNullIfOtherFieldValidator) Description(_ context.Context) string {
	return fmt.Sprintf("The field '%s' must be null if the field '%s' is set to '%s'", s.targetField, s.fieldToCheck, s.checkValue)
}

func (s listFieldNullIfOtherFieldValidator) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func ListFieldNullIfOtherField(targetField path.Expression, fieldToCheck path.Expression, checkValue string) validator.Object {
	return &listFieldNullIfOtherFieldValidator{
		targetField:  targetField,
		fieldToCheck: fieldToCheck,
		checkValue:   checkValue,
	}
}
