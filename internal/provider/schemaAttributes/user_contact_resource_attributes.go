package schemaAttributes

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var UserContactResourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"method": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf("email", "sms", "voice", "mobile"),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
				if request.StateValue.ValueString() != request.ConfigValue.ValueString() {
					response.RequiresReplace = true
					return
				}
			},
				"Force replacement since method value updated",
				"Force replacement since method value updated"),
		},
		Description: "The method of contact for the user. Valid values are 'email', 'sms', 'voice', or 'mobile'.",
	},
	"to": schema.StringAttribute{
		Required:    true,
		Description: "The contact information for the user, such as an email address or phone number.",
	},
	"enabled": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether this contact method is enabled for the user.",
	},
}
