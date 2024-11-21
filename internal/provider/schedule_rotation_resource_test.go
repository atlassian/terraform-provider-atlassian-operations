package provider

import (
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
data "atlassian-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-ops_team.example.id
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = atlassian-ops_schedule.example.id
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  end_date = "2023-11-11T05:00:00Z"
  type       = "weekly"
  length     = 2
  participants = [
	{
	  id = data.atlassian-ops_user.test1.account_id
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
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "length", "2"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.#", "1"),
					resource.TestCheckResourceAttrPair("atlassian-ops_schedule_rotation.example", "participants.0.id", "data.atlassian-ops_user.test1", "account_id"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.0.type", "user"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-ops_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-ops_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-ops_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-ops_team.example.id
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = atlassian-ops_schedule.example.id
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
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "end_date", "2023-11-11T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "type", "daily"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.start_hour", "10"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.start_min", "30"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restriction.end_min", "30"),
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
data "atlassian-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-ops_team.example.id
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = atlassian-ops_schedule.example.id
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
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.#", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-ops_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-ops_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-ops_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-ops_team.example.id
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = atlassian-ops_schedule.example.id
  name       = "` + rotationUpdateName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
  participants = [
	{
	  id = data.atlassian-ops_user.test1.account_id
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
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.#", "1"),
					resource.TestCheckResourceAttrPair("atlassian-ops_schedule_rotation.example", "participants.0.id", "data.atlassian-ops_user.test1", "account_id"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.0.type", "user"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.#", "2"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.1.start_day", "tuesday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.1.end_day", "thursday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.1.start_hour", "10"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.1.end_hour", "19"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.1.start_min", "30"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.1.end_min", "30"),
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
data "atlassian-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-ops_team.example.id
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = atlassian-ops_schedule.example.id
  name       = "` + rotationName + `"
  start_date = "2023-11-10T05:00:00Z"
  type       = "weekly"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "name", rotationName),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-ops_schedule_rotation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-ops_schedule_rotation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-ops_schedule_rotation.example"].Primary.Attributes["schedule_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-ops_team.example.id
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = atlassian-ops_schedule.example.id
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
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "name", rotationUpdateName),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "start_date", "2023-11-10T05:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "type", "weekly"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "length", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "participants.#", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.#", "1"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-ops_schedule_rotation.example", "time_restriction.restrictions.0.end_min", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
