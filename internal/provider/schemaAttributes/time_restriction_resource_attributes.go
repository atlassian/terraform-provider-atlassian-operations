package schemaAttributes

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var weekdayValidator = []validator.String{
	stringvalidator.OneOf([]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}...),
}

var hourValidator = []validator.Int32{
	int32validator.AtLeast(0),
	int32validator.AtMost(23),
}

var minuteValidator = []validator.Int32{
	int32validator.OneOf([]int32{0, 30}...),
}

var TimeRestrictionResourceAttributes = map[string]schema.Attribute{
	"type": schema.StringAttribute{
		Description: "The type of the time restriction",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{"time-of-day", "weekday-and-time-of-day"}...),
		},
	},
	"restriction": schema.SingleNestedAttribute{
		Computed:   true,
		Optional:   true,
		Attributes: TimeOfDayTimeRestrictionResourceAttributes,
	},
	"restrictions": schema.ListNestedAttribute{
		Description: "The restrictions of the time restriction",
		Computed:    true,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: WeekdayTimeRestrictionResourceAttributes,
		},
	},
}

var TimeOfDayTimeRestrictionResourceAttributes = map[string]schema.Attribute{
	"start_hour": schema.Int32Attribute{
		Description: "The start hour of the restriction",
		Required:    true,
		Validators:  hourValidator,
	},
	"end_hour": schema.Int32Attribute{
		Description: "The end hour of the restriction",
		Required:    true,
		Validators:  hourValidator,
	},
	"start_min": schema.Int32Attribute{
		Description: "The start minute of the restriction",
		Required:    true,
		Validators:  minuteValidator,
	},
	"end_min": schema.Int32Attribute{
		Description: "The end minute of the restriction",
		Required:    true,
		Validators:  minuteValidator,
	},
}

var WeekdayTimeRestrictionResourceAttributes = map[string]schema.Attribute{
	"start_day": schema.StringAttribute{
		Description: "The start day of the restriction",
		Required:    true,
		Validators:  weekdayValidator,
	},
	"end_day": schema.StringAttribute{
		Description: "The end day of the restriction",
		Required:    true,
		Validators:  weekdayValidator,
	},
	"start_hour": schema.Int32Attribute{
		Description: "The start hour of the restriction",
		Required:    true,
		Validators:  hourValidator,
	},
	"end_hour": schema.Int32Attribute{
		Description: "The end hour of the restriction",
		Required:    true,
		Validators:  hourValidator,
	},
	"start_min": schema.Int32Attribute{
		Description: "The start minute of the restriction",
		Required:    true,
		Validators:  minuteValidator,
	},
	"end_min": schema.Int32Attribute{
		Description: "The end minute of the restriction",
		Required:    true,
		Validators:  minuteValidator,
	},
}
