package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHeartbeatResource(t *testing.T) {
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
				Config: providerConfig + testAccHeartbeatResourceConfig(teamName, emailPrimary, organizationId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "name", "test-heartbeat"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "description", "Test heartbeat"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "interval", "5"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "interval_unit", "minutes"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "alert_message", "Service heartbeat missed"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "alert_tags.#", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "alert_priority", "P2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_heartbeat.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The import state doesn't include certain computed fields
				ImportStateVerifyIgnore:              []string{"status"},
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_heartbeat.test"].Primary.Attributes["name"] +
							"," +
							state.RootModule().Resources["atlassian-operations_heartbeat.test"].Primary.Attributes["team_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccHeartbeatResourceUpdatedConfig(teamName, emailPrimary, organizationId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "name", "test-heartbeat"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "description", "Updated test heartbeat"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "interval", "10"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "interval_unit", "minutes"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "alert_message", "Critical service heartbeat missed"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "alert_tags.#", "3"),
					resource.TestCheckResourceAttr("atlassian-operations_heartbeat.test", "alert_priority", "P1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccHeartbeatResourceConfig(teamName string, emailPrimary string, organizationId string) string {
	return `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
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

resource "atlassian-operations_heartbeat" "test" {
  name          = "test-heartbeat"
  description   = "Test heartbeat"
  interval      = 5
  interval_unit = "minutes"
  enabled       = true
  team_id       = atlassian-operations_team.example.id
  alert_message = "Service heartbeat missed"
  alert_tags    = ["critical", "service"]
  alert_priority = "P2"
}
`
}

func testAccHeartbeatResourceUpdatedConfig(teamName string, emailPrimary string, organizationId string) string {
	return `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
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

resource "atlassian-operations_heartbeat" "test" {
  name          = "test-heartbeat"
  description   = "Updated test heartbeat"
  interval      = 10
  interval_unit = "minutes"
  enabled       = true
  team_id       = atlassian-operations_team.example.id
  alert_message = "Critical service heartbeat missed"
  alert_tags    = ["critical", "service", "high-priority"]
  alert_priority = "P1"
}
`
}
