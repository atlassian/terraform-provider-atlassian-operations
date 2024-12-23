package schemaAttributes

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

var TimeRestrictionDataSourceAttributes = map[string]schema.Attribute{
	"type": schema.StringAttribute{
		Description: "The type of the time restriction",
		Computed:    true,
	},
	"restriction": schema.SingleNestedAttribute{
		Computed:   true,
		Optional:   true,
		Attributes: TimeOfDayTimeRestrictionDataSourceAttributes,
	},
	"restrictions": schema.ListNestedAttribute{
		Description: "The restrictions of the time restriction",
		Computed:    true,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: WeekdayTimeRestrictionDataSourceAttributes,
		},
	},
}

var TimeOfDayTimeRestrictionDataSourceAttributes = map[string]schema.Attribute{
	"start_hour": schema.Int32Attribute{
		Description: "The start hour of the restriction",
		Computed:    true,
	},
	"end_hour": schema.Int32Attribute{
		Description: "The end hour of the restriction",
		Computed:    true,
	},
	"start_min": schema.Int32Attribute{
		Description: "The start minute of the restriction",
		Computed:    true,
	},
	"end_min": schema.Int32Attribute{
		Description: "The end minute of the restriction",
		Computed:    true,
	},
}

var WeekdayTimeRestrictionDataSourceAttributes = map[string]schema.Attribute{
	"start_day": schema.StringAttribute{
		Description: "The start day of the restriction",
		Computed:    true,
	},
	"end_day": schema.StringAttribute{
		Description: "The end day of the restriction",
		Computed:    true,
	},
	"start_hour": schema.Int32Attribute{
		Description: "The start hour of the restriction",
		Computed:    true,
	},
	"end_hour": schema.Int32Attribute{
		Description: "The end hour of the restriction",
		Computed:    true,
	},
	"start_min": schema.Int32Attribute{
		Description: "The start minute of the restriction",
		Computed:    true,
	},
	"end_min": schema.Int32Attribute{
		Description: "The end minute of the restriction",
		Computed:    true,
	},
}
