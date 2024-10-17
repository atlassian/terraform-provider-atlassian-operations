package provider

import (
	"github.com/google/uuid"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleResource_Full(t *testing.T) {
	scheduleName := uuid.NewString()
	rotationName := uuid.NewString()

	scheduleUpdateName := uuid.NewString()
	rotationUpdateName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
  description = "schedule description"
  timezone = "Europe/Istanbul"
  enabled = true
  rotations = [
    {
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
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "name", scheduleName),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "description", "schedule description"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "timezone", "Europe/Istanbul"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "team_id", "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.name", rotationName),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.length", "2"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.participants.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.participants.0.id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.participants.0.type", "user"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.time_restriction.restriction.start_hour", "9"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.time_restriction.restriction.start_min", "0"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.time_restriction.restriction.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_schedule.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule" "example" {
  name    = "` + scheduleUpdateName + `"
  team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
  rotations = [
    {
      name       = "` + rotationUpdateName + `"
      start_date = "2023-11-10T06:00:00Z"
      type       = "weekly"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "name", scheduleUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "description", ""),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "team_id", "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.name", rotationUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.start_date", "2023-11-10T06:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.type", "weekly"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.length", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.participants.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccScheduleResource_Minimal(t *testing.T) {
	scheduleName := uuid.NewString()

	scheduleUpdateName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
  rotations = [
    {
      start_date = "2023-11-10T06:00:00Z"
      type       = "weekly"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "name", scheduleName),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "team_id", "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.start_date", "2023-11-10T06:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.type", "weekly"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_schedule.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_schedule" "example" {
  name    = "` + scheduleUpdateName + `"
  team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
  description = "schedule description"
  enabled = false
  rotations = [
    {
      start_date = "2023-11-10T06:00:00Z"
      type       = "weekly"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "name", scheduleUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "description", "schedule description"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "enabled", "false"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "team_id", "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.start_date", "2023-11-10T06:00:00Z"),
					resource.TestCheckResourceAttr("jsm-ops_schedule.example", "rotations.0.type", "weekly"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
