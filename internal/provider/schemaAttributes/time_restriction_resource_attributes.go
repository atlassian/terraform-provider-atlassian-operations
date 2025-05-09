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
		Description: "The type of time restriction to apply. Must be either 'time-of-day' for daily recurring windows or 'weekday-and-time-of-day' for weekly schedules.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{"time-of-day", "weekday-and-time-of-day"}...),
		},
	},
	"restriction": schema.SingleNestedAttribute{
		Description: "Configuration for daily time windows. Used when type is 'time-of-day'. Specifies the same time window for every day.",
		Optional:    true,
		Attributes:  TimeOfDayTimeRestrictionResourceAttributes,
	},
	"restrictions": schema.ListNestedAttribute{
		Description: "List of weekly time windows. Used when type is 'weekday-and-time-of-day'. Allows different time windows for different days of the week.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: WeekdayTimeRestrictionResourceAttributes,
		},
	},
}

var TimeOfDayTimeRestrictionResourceAttributes = map[string]schema.Attribute{
	"start_hour": schema.Int32Attribute{
		Description: "The hour when the restriction begins (0-23, where 0 is midnight). Must be a valid 24-hour time.",
		Required:    true,
		Validators:  hourValidator,
	},
	"end_hour": schema.Int32Attribute{
		Description: "The hour when the restriction ends (0-23, where 0 is midnight). Must be a valid 24-hour time.",
		Required:    true,
		Validators:  hourValidator,
	},
	"start_min": schema.Int32Attribute{
		Description: "The minute when the restriction begins. Must be either 0 or 30 (half-hour increments only).",
		Required:    true,
		Validators:  minuteValidator,
	},
	"end_min": schema.Int32Attribute{
		Description: "The minute when the restriction ends. Must be either 0 or 30 (half-hour increments only).",
		Required:    true,
		Validators:  minuteValidator,
	},
}

var WeekdayTimeRestrictionResourceAttributes = map[string]schema.Attribute{
	"start_day": schema.StringAttribute{
		Description: "The day of the week when the restriction begins. Must be a lowercase day name (e.g., 'monday', 'tuesday').",
		Required:    true,
		Validators:  weekdayValidator,
	},
	"end_day": schema.StringAttribute{
		Description: "The day of the week when the restriction ends. Must be a lowercase day name (e.g., 'monday', 'tuesday').",
		Required:    true,
		Validators:  weekdayValidator,
	},
	"start_hour": schema.Int32Attribute{
		Description: "The hour when the restriction begins on the start day (0-23, where 0 is midnight). Must be a valid 24-hour time.",
		Required:    true,
		Validators:  hourValidator,
	},
	"end_hour": schema.Int32Attribute{
		Description: "The hour when the restriction ends on the end day (0-23, where 0 is midnight). Must be a valid 24-hour time.",
		Required:    true,
		Validators:  hourValidator,
	},
	"start_min": schema.Int32Attribute{
		Description: "The minute when the restriction begins on the start day. Must be either 0 or 30 (half-hour increments only).",
		Required:    true,
		Validators:  minuteValidator,
	},
	"end_min": schema.Int32Attribute{
		Description: "The minute when the restriction ends on the end day. Must be either 0 or 30 (half-hour increments only).",
		Required:    true,
		Validators:  minuteValidator,
	},
}
