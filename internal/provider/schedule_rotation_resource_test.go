package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleRotationResource_TimeOfDay(t *testing.T) {
	rotationName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  end_date = "2023-11-11T05:00:00Z"
  type       = "weekly"
  length     = 2
  participants = [
	{
	  id = "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"
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
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "length", "2"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.0.id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.0.type", "user"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.start_hour", "9"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.start_min", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["jsm-ops_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["jsm-ops_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
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
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "type", "daily"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.start_hour", "10"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.start_min", "30"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restriction.end_min", "30"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccScheduleRotationResource_WeekdayAndTimeOfDay(t *testing.T) {
	rotationName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
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
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["jsm-ops_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["jsm-ops_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
  name       = "` + rotationUpdateName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
  participants = [
	{
	  id = "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"
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
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.0.id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.0.type", "user"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.#", "2"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.1.start_day", "tuesday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.1.end_day", "thursday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.1.start_hour", "10"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.1.end_hour", "19"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.1.start_min", "30"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.1.end_min", "30"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccScheduleRotationResource_NoRestriction(t *testing.T) {
	rotationName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["jsm-ops_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["jsm-ops_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
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
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
