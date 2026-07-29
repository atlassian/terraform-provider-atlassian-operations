package provider

import (
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
)

func TestRotationDtoToModel_TimeRestrictionOrderIsIgnored(t *testing.T) {
	monday := dto.WeekdayTimeRestrictionSettings{
		StartDay: dto.Monday, EndDay: dto.Monday,
		StartHour: 9, EndHour: 17,
	}
	tuesday := dto.WeekdayTimeRestrictionSettings{
		StartDay: dto.Tuesday, EndDay: dto.Tuesday,
		StartHour: 9, EndHour: 17,
	}

	firstOrder := []dto.WeekdayTimeRestrictionSettings{monday, tuesday}
	reversedOrder := []dto.WeekdayTimeRestrictionSettings{tuesday, monday}

	first := RotationDtoToModel("schedule-id", dto.Rotation{
		TimeRestriction: &dto.TimeRestriction{
			Type:                        dto.WeekdayAndTimeOfDay,
			WeekAndTimeOfDayRestriction: &firstOrder,
		},
	})
	reversed := RotationDtoToModel("schedule-id", dto.Rotation{
		TimeRestriction: &dto.TimeRestriction{
			Type:                        dto.WeekdayAndTimeOfDay,
			WeekAndTimeOfDayRestriction: &reversedOrder,
		},
	})

	if !first.TimeRestriction.Equal(reversed.TimeRestriction) {
		t.Fatalf("expected restriction order to be ignored\nfirst: %s\nreversed: %s", first.TimeRestriction, reversed.TimeRestriction)
	}
}
