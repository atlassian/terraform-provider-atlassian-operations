package provider

import (
	"github.com/google/uuid"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleResource_Full(t *testing.T) {
	scheduleName := uuid.NewString()
	scheduleUpdateName := uuid.NewString()

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
  description = "schedule description"
  timezone = "Europe/Istanbul"
  enabled = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "name", scheduleName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "description", "schedule description"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "timezone", "Europe/Istanbul"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "enabled", "true"),
					resource.TestCheckResourceAttrPair("atlassian-operations_schedule.example", "team_id", "atlassian-operations_team.example", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_schedule.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
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
  name    = "` + scheduleUpdateName + `"
  team_id = atlassian-operations_team.example.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "name", scheduleUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "description", ""),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "enabled", "true"),
					resource.TestCheckResourceAttrPair("atlassian-operations_schedule.example", "team_id", "atlassian-operations_team.example", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccScheduleResource_Minimal(t *testing.T) {
	scheduleName := uuid.NewString()

	scheduleUpdateName := uuid.NewString()

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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "name", scheduleName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "enabled", "true"),
					resource.TestCheckResourceAttrPair("atlassian-operations_schedule.example", "team_id", "atlassian-operations_team.example", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_schedule.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
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
  name    = "` + scheduleUpdateName + `"
  team_id = atlassian-operations_team.example.id
  description = "schedule description"
  enabled = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "name", scheduleUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "description", "schedule description"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("atlassian-operations_schedule.example", "enabled", "false"),
					resource.TestCheckResourceAttrPair("atlassian-operations_schedule.example", "team_id", "atlassian-operations_team.example", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
