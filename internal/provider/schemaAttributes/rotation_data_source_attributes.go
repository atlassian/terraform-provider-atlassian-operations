package schemaAttributes

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

var RotationDataSourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the rotation",
		Computed:    true,
	},
	"name": schema.StringAttribute{
		Description: "The name of the rotation",
		Computed:    true,
	},
	"start_date": schema.StringAttribute{
		Description: "The start date of the rotation",
		Computed:    true,
	},
	"end_date": schema.StringAttribute{
		Description: "The end date of the rotation",
		Computed:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the rotation",
		Computed:    true,
	},
	"length": schema.Int32Attribute{
		Description: "The length of the rotation",
		Computed:    true,
	},
	"participants": schema.ListNestedAttribute{
		Description: "The participants of the rotation",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: ResponderInfoDataSourceAttributes,
		},
	},
	"time_restriction": schema.SingleNestedAttribute{
		Attributes: TimeRestrictionDataSourceAttributes,
		Computed:   true,
	},
}

var ResponderInfoDataSourceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Description: "The ID of the participant",
		Computed:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the participant",
		Computed:    true,
	},
}
