package provider

import (
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleRotationResource_TimeOfDay(t *testing.T) {
	rotationName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	scheduleName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = atlassian-operations_schedule.example.id
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  end_date = "2023-11-11T05:00:00Z"
  type       = "weekly"
  length     = 2
  participants = [
	{
	  id = data.atlassian-operations_user.test1.account_id
	  type = "user"
	}
  ]
  time_restriction = {
	type = "time-of-day"
	restriction = {
	  start_hour = 9
	  end_hour = 17
	  start_min = 0
	  end_min = 0
	}
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "length", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.#", "1"),
					resource.TestCheckResourceAttrPair("atlassian-operations_schedule_rotation.example", "participants.0.id", "data.atlassian-operations_user.test1", "account_id"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.0.type", "user"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = atlassian-operations_schedule.example.id
  name       = "` + rotationUpdateName + `"
  start_date = "2023-11-10T05:00:00Z"
  end_date = "2023-11-11T05:00:00Z"
  type       = "daily"
  length     = 1
  time_restriction = {
	type = "time-of-day"
	restriction = {
	  start_hour = 10
	  end_hour = 17
	  start_min = 30
	  end_min = 30
	}
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "type", "daily"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.start_hour", "10"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.start_min", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restriction.end_min", "30"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccScheduleRotationResource_WeekdayAndTimeOfDay(t *testing.T) {
	rotationName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	scheduleName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = atlassian-operations_schedule.example.id
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
  time_restriction = {
	type = "weekday-and-time-of-day"
	restrictions = [
	  {
		start_day = "monday"
		end_day = "friday"
	    start_hour = 9
	    end_hour = 17
	    start_min = 0	
	    end_min = 0
	  }
    ]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = atlassian-operations_schedule.example.id
  name       = "` + rotationUpdateName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
  participants = [
	{
	  id = data.atlassian-operations_user.test1.account_id
	  type = "user"
	}
  ]
  time_restriction = {
	type = "weekday-and-time-of-day"
	restrictions = [
	  {
		start_day = "monday"
		end_day = "friday"
	    start_hour = 9
	    end_hour = 17
	    start_min = 0	
	    end_min = 0
	  },
	  {
		start_day = "tuesday"
		end_day = "thursday"
	    start_hour = 10
	    end_hour = 19
	    start_min = 30	
	    end_min = 30
	  }
    ]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.#", "1"),
					resource.TestCheckResourceAttrPair("atlassian-operations_schedule_rotation.example", "participants.0.id", "data.atlassian-operations_user.test1", "account_id"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.0.type", "user"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.#", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.1.start_day", "tuesday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.1.end_day", "thursday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.1.start_hour", "10"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.1.end_hour", "19"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.1.start_min", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.1.end_min", "30"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccScheduleRotationResource_NoRestriction(t *testing.T) {
	rotationName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	scheduleName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = atlassian-operations_schedule.example.id
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = atlassian-operations_schedule.example.id
  name       = "` + rotationUpdateName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
  time_restriction = {
	type = "weekday-and-time-of-day"
	restrictions = [
	  {
		start_day = "monday"
		end_day = "friday"
	    start_hour = 9
	    end_hour = 17
	    start_min = 0	
	    end_min = 0
	  }
    ]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func makeRestriction(startDay, endDay dto.Weekday, startHour, endHour, startMin, endMin int32) dto.WeekdayTimeRestrictionSettings {
	return dto.WeekdayTimeRestrictionSettings{
		StartDay:  startDay,
		EndDay:    endDay,
		StartHour: startHour,
		EndHour:   endHour,
		StartMin:  startMin,
		EndMin:    endMin,
	}
}

func TestNormalizeRestrictionsOrder_SameOrder(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Wednesday, dto.Thursday, 10, 18, 0, 0),
	}
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Wednesday, dto.Thursday, 10, 18, 0, 0),
	}

	result := normalizeRestrictionsOrder(planned, received)

	if len(*result) != 2 {
		t.Fatalf("expected 2 restrictions, got %d", len(*result))
	}
	if (*result)[0].StartDay != dto.Monday {
		t.Errorf("expected restrictions[0].start_day=monday, got %s", (*result)[0].StartDay)
	}
	if (*result)[1].StartDay != dto.Wednesday {
		t.Errorf("expected restrictions[1].start_day=wednesday, got %s", (*result)[1].StartDay)
	}
}

func TestNormalizeRestrictionsOrder_ReorderedByServer(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Wednesday, dto.Thursday, 10, 18, 0, 0),
		makeRestriction(dto.Friday, dto.Saturday, 19, 9, 0, 0),
	}
	// Server returns them in a different order
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Friday, dto.Saturday, 19, 9, 0, 0),
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Wednesday, dto.Thursday, 10, 18, 0, 0),
	}

	result := normalizeRestrictionsOrder(planned, received)

	if len(*result) != 3 {
		t.Fatalf("expected 3 restrictions, got %d", len(*result))
	}
	if (*result)[0].StartDay != dto.Monday {
		t.Errorf("expected restrictions[0].start_day=monday, got %s", (*result)[0].StartDay)
	}
	if (*result)[1].StartDay != dto.Wednesday {
		t.Errorf("expected restrictions[1].start_day=wednesday, got %s", (*result)[1].StartDay)
	}
	if (*result)[2].StartDay != dto.Friday {
		t.Errorf("expected restrictions[2].start_day=friday, got %s", (*result)[2].StartDay)
	}
}

func TestNormalizeRestrictionsOrder_NilPlanned(t *testing.T) {
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
	}

	result := normalizeRestrictionsOrder(nil, received)

	if result != received {
		t.Error("expected received to be returned as-is when planned is nil")
	}
}

func TestNormalizeRestrictionsOrder_NilReceived(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
	}

	result := normalizeRestrictionsOrder(planned, nil)

	if result != nil {
		t.Error("expected nil to be returned when received is nil")
	}
}

func TestNormalizeRestrictionsOrder_DifferentLengths(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
	}
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Wednesday, dto.Thursday, 10, 18, 0, 0),
	}

	result := normalizeRestrictionsOrder(planned, received)

	if result != received {
		t.Error("expected received to be returned as-is when lengths differ")
	}
}

func TestNormalizeRestrictionsOrder_DifferentContent(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
	}
	// Server returns a genuinely different restriction
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Friday, dto.Saturday, 10, 18, 0, 0),
	}

	result := normalizeRestrictionsOrder(planned, received)

	if result != received {
		t.Error("expected received to be returned as-is when content differs")
	}
}

func TestNormalizeRestrictionsOrder_EmptyLists(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{}
	received := &[]dto.WeekdayTimeRestrictionSettings{}

	result := normalizeRestrictionsOrder(planned, received)

	if len(*result) != 0 {
		t.Errorf("expected empty result, got %d items", len(*result))
	}
}

func TestNormalizeRestrictionsOrder_PreservesAllFields(t *testing.T) {
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Friday, 9, 17, 30, 45),
	}
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Friday, 9, 17, 30, 45),
	}

	result := normalizeRestrictionsOrder(planned, received)

	r := (*result)[0]
	if r.StartDay != dto.Monday || r.EndDay != dto.Friday ||
		r.StartHour != 9 || r.EndHour != 17 ||
		r.StartMin != 30 || r.EndMin != 45 {
		t.Errorf("fields not preserved correctly: %+v", r)
	}
}

func TestNormalizeRestrictionsOrder_DuplicateEntries(t *testing.T) {
	// Each duplicate in planned should match a distinct entry in received
	planned := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
	}
	received := &[]dto.WeekdayTimeRestrictionSettings{
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
		makeRestriction(dto.Monday, dto.Tuesday, 9, 17, 0, 0),
	}

	result := normalizeRestrictionsOrder(planned, received)

	if len(*result) != 2 {
		t.Fatalf("expected 2 restrictions, got %d", len(*result))
	}
}
