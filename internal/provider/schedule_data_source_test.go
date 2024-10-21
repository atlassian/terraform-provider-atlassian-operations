package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "jsm-ops_schedule" "test" {name = "Test_schedule"}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "id", "df47a95c-f9ae-4ca6-873b-375fcad3cd18"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "name", "Test_schedule"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "description", "Test Description"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "timezone", "Europe/Istanbul"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "team_id", "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.#", "1"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.id", "6174e943-e234-4e6a-8260-68c644553836"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.name", "Rotation 2"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.start_date", "2024-09-17T05:00:00Z"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.end_date", ""),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.type", "weekly"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.length", "1"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.participants.#", "2"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.participants.0.id", "712020:ce8310ee-7509-41b5-baa5-0c4f74dae467"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.participants.0.type", "user"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.time_restriction.restriction.start_hour", "8"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.time_restriction.restriction.start_min", "0"),
					resource.TestCheckResourceAttr("data.jsm-ops_schedule.test", "rotations.0.time_restriction.restriction.end_min", "0"),
				),
			},
		},
	})
}
