package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMaintenanceResource(t *testing.T) {
	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	teamName := uuid.NewString()
	apiIntegrationName := uuid.NewString()
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
				Config: providerConfig + testAccMaintenanceResourceConfig(emailPrimary, teamName, organizationId, apiIntegrationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "description", "Test Maintenance Window"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "start_date", "2029-06-15T10:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "end_date", "2029-06-15T14:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "rules.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "rules.0.state", "disabled"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "rules.0.entity.type", "integration"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_maintenance.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccMaintenanceResourceUpdatedConfig(emailPrimary, teamName, organizationId, apiIntegrationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "description", "Updated Test Maintenance Window"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "start_date", "2029-06-16T10:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "end_date", "2029-06-16T16:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.test", "rules.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccMaintenanceResourceWithTeam(t *testing.T) {
	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	teamName := uuid.NewString()
	apiIntegrationName := uuid.NewString()
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
			// Create and Read testing for team-specific maintenance window
			{
				Config: providerConfig + testAccMaintenanceResourceWithTeamConfig(emailPrimary, teamName, organizationId, apiIntegrationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "description", "Team-specific Maintenance Window"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "start_date", "2029-07-15T10:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "end_date", "2029-07-15T14:00:00Z"),
					resource.TestCheckResourceAttrSet("atlassian-operations_maintenance.team_test", "team_id"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "rules.#", "1"),
				),
			},
			// ImportState testing for team-specific maintenance window
			{
				ResourceName:      "atlassian-operations_maintenance.team_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_maintenance.team_test"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_maintenance.team_test"].Primary.Attributes["team_id"],
						nil
				},
			},
			// Update and Read testing for team-specific maintenance window
			{
				Config: providerConfig + testAccMaintenanceResourceWithTeamUpdatedConfig(emailPrimary, teamName, organizationId, apiIntegrationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "description", "Updated Team-specific Maintenance Window"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "start_date", "2029-07-16T10:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "end_date", "2029-07-16T16:00:00Z"),
					resource.TestCheckResourceAttr("atlassian-operations_maintenance.team_test", "rules.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccMaintenanceResourceConfig(apiPrimary string, teamName string, organizationId string, apiIntegrationName string) string {
	return `
data "atlassian-operations_user" "test1" {
	email_address = "` + apiPrimary + `"
  	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  type = "API"
  enabled = true
}

resource "atlassian-operations_maintenance" "test" {
  description = "Test Maintenance Window"
  start_date  = "2029-06-15T10:00:00Z"
  end_date    = "2029-06-15T14:00:00Z"
  
  rules = [
	{
    	state = "disabled"
    	entity = {
      		id   = atlassian-operations_api_integration.example.id
      		type = "integration"
    	}
  	}
	]
}
`
}

func testAccMaintenanceResourceUpdatedConfig(apiPrimary string, teamName string, organizationId string, apiIntegrationName string) string {
	return `
data "atlassian-operations_user" "test1" {
	email_address = "` + apiPrimary + `"
  	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  type = "API"
  enabled = true
}

resource "atlassian-operations_maintenance" "test" {
  description = "Updated Test Maintenance Window"
  start_date  = "2029-06-16T10:00:00Z"
  end_date    = "2029-06-16T16:00:00Z"
  
  rules = [ {
    state = "disabled"
    entity = {
      id   = atlassian-operations_api_integration.example.id
      type = "integration"
    }
  }
  ]
}
`
}

func testAccMaintenanceResourceWithTeamConfig(apiPrimary string, teamName string, organizationId string, apiIntegrationName string) string {
	return `
data "atlassian-operations_user" "test1" {
	email_address = "` + apiPrimary + `"
  	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  type = "API"
  enabled = true
}
resource "atlassian-operations_maintenance" "team_test" {
  description = "Team-specific Maintenance Window"
  start_date  = "2029-07-15T10:00:00Z"
  end_date    = "2029-07-15T14:00:00Z"
  team_id     = atlassian-operations_team.example.id
  
  rules = [
	{
		state = "disabled"
		entity = {
		  id   = atlassian-operations_api_integration.example.id
		  type = "integration"
		}
	  }
	]
}
`
}

func testAccMaintenanceResourceWithTeamUpdatedConfig(apiPrimary string, teamName string, organizationId string, apiIntegrationName string) string {
	return `
data "atlassian-operations_user" "test1" {
	email_address = "` + apiPrimary + `"
  	organization_id = "` + organizationId + `"
}

resource "atlassian-operations_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  type = "API"
  enabled = true
}
resource "atlassian-operations_maintenance" "team_test" {
  description = "Updated Team-specific Maintenance Window"
  start_date  = "2029-07-16T10:00:00Z"
  end_date    = "2029-07-16T16:00:00Z"
  team_id     = atlassian-operations_team.example.id
  
  rules= [
	{
    state = "disabled"
    entity = {
      id   = atlassian-operations_api_integration.example.id
      type = "integration"
    }
  }
	]
}
`
}
