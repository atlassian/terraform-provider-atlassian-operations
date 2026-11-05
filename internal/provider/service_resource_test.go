package provider

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceResource(t *testing.T) {
	serviceName := uuid.NewString()
	teamName := uuid.NewString()
	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

resource "atlassian-operations_service" "example" {
  name        = "` + serviceName + `"
  description = "Test JSM Service Description"
  tier        = 3
  type        = "BUSINESS_SERVICES"
  owner = atlassian-operations_team.example.id
  change_approvers = {
    groups = [
      "af4a5948-3c4d-4703-834e-404b02c8b313"
    ]
  }
  responders = {
    users = [
      data.atlassian-operations_user.test1.account_id
    ]
    teams = [
      atlassian-operations_team.example.id
    ]
  }
  stakeholders = {
    users = [
      data.atlassian-operations_user.test1.account_id
    ]
  }
  projects = {
    ids = [
      "10001"
    ]
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "name", serviceName),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "description", "Test JSM Service Description"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "tier", "3"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "type", "BUSINESS_SERVICES"),
					resource.TestCheckResourceAttrPair("atlassian-operations_service.example", "owner", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "change_approvers.groups.0", "af4a5948-3c4d-4703-834e-404b02c8b313"),
					resource.TestCheckResourceAttrPair("atlassian-operations_service.example", "responders.users.0", "data.atlassian-operations_user.test1", "account_id"),
					resource.TestCheckResourceAttrPair("atlassian-operations_service.example", "responders.teams.0", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttrPair("atlassian-operations_service.example", "stakeholders.users.0", "data.atlassian-operations_user.test1", "account_id"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "projects.ids.0", "10001"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_service.example",
				ImportState:       true,
				ImportStateVerify: true,
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

resource "atlassian-operations_service" "example" {
  name        = "` + serviceName + ` - Updated"
  description = "Updated Test JSM Service Description"
  tier        = 2
  type        = "BUSINESS_SERVICES"
  change_approvers = {
    groups = [
      "af4a5948-3c4d-4703-834e-404b02c8b313"
    ]
  }
  responders = {
    users = [
      data.atlassian-operations_user.test1.account_id
    ]
    teams = []
  }
  stakeholders = {
    users = [
      data.atlassian-operations_user.test1.account_id
    ]
  }
  projects = {
    ids = [
      "10001"
    ]
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "name", serviceName+" - Updated"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "description", "Updated Test JSM Service Description"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "tier", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "type", "BUSINESS_SERVICES"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "change_approvers.groups.0", "af4a5948-3c4d-4703-834e-404b02c8b313"),
					resource.TestCheckResourceAttrPair("atlassian-operations_service.example", "responders.users.0", "data.atlassian-operations_user.test1", "account_id"),
					resource.TestCheckResourceAttrPair("atlassian-operations_service.example", "stakeholders.users.0", "data.atlassian-operations_user.test1", "account_id"),
					resource.TestCheckResourceAttr("atlassian-operations_service.example", "projects.ids.0", "10001"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
