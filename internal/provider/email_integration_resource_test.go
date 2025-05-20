package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEmailIntegrationResource(t *testing.T) {
	emailIntegrationName := uuid.NewString()
	emailIntegrationUpdateName := uuid.NewString()

	randomEmailUsername := uuid.NewString()
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


resource "atlassian-operations_email_integration" "example" {
  name    = "` + emailIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  enabled = true
  type_specific_properties = {
  	email_username = "` + randomEmailUsername + `"
    suppress_notifications = true
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "name", emailIntegrationName),
					resource.TestCheckResourceAttrPair("atlassian-operations_email_integration.example", "team_id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "type_specific_properties.email_username", randomEmailUsername),
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "type_specific_properties.suppress_notifications", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "atlassian-operations_email_integration.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"directions.#", "domains.#", "domains.0", "directions.0"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_email_integration.example"].Primary.ID,
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
resource "atlassian-operations_email_integration" "example" {
  name    = "` + emailIntegrationUpdateName + `"
  team_id = atlassian-operations_team.example.id
  enabled = false
  type_specific_properties = {
  	email_username = "` + randomEmailUsername + `"
    suppress_notifications = false
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "name", emailIntegrationUpdateName),
					resource.TestCheckResourceAttrPair("atlassian-operations_email_integration.example", "team_id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "enabled", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "type_specific_properties.email_username", randomEmailUsername),
					resource.TestCheckResourceAttr("atlassian-operations_email_integration.example", "type_specific_properties.suppress_notifications", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
