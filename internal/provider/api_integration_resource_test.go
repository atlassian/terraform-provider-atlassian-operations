package provider

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApiIntegrationResource_Api(t *testing.T) {
	apiIntegrationName := uuid.NewString()
	apiIntegrationUpdateName := uuid.NewString()

	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	apiPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if apiPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
}

resource "atlassian-ops_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-ops_team.example.id
  type = "API"
  enabled = true
  type_specific_properties = jsonencode({
    suppressNotifications: false
  })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "name", apiIntegrationName),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "type", "API"),
					resource.TestCheckResourceAttrPair("atlassian-ops_api_integration.example", "team_id", "atlassian-ops_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "enabled", "true"),
					resource.TestCheckResourceAttrWith("atlassian-ops_api_integration.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["suppressNotifications"] != false {
							return fmt.Errorf("expected to get 'false', got '%s'", dataObj["suppressNotifications"])
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "atlassian-ops_api_integration.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type_specific_properties", "directions", "domains"},
			},
			// Update and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
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
resource "atlassian-ops_api_integration" "example" {
  name    = "` + apiIntegrationUpdateName + `"
  team_id = atlassian-ops_team.example.id
  type = "API"
  enabled = false
  type_specific_properties = jsonencode({
    suppressNotifications: true
  })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "name", apiIntegrationUpdateName),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "type", "API"),
					resource.TestCheckResourceAttrPair("atlassian-ops_api_integration.example", "team_id", "atlassian-ops_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "enabled", "false"),
					resource.TestCheckResourceAttrWith("atlassian-ops_api_integration.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["suppressNotifications"] != true {
							return fmt.Errorf("'suppressNotifications' expected to get 'true', got '%s'", dataObj["suppressNotifications"])
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccApiIntegrationResource_SecurityHub(t *testing.T) {
	apiIntegrationName := uuid.NewString()
	apiIntegrationUpdateName := uuid.NewString()

	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	apiPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if apiPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
}

resource "atlassian-ops_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-ops_user.test1.account_id
    }
  ]
}

resource "atlassian-ops_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  enabled = true
  type = "AmazonSecurityHub"
  team_id = atlassian-ops_team.example.id
  type_specific_properties = jsonencode({
      suppressNotifications: false
      region: "US_WEST_2"
    })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "name", apiIntegrationName),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "type", "AmazonSecurityHub"),
					resource.TestCheckResourceAttrPair("atlassian-ops_api_integration.example", "team_id", "atlassian-ops_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "enabled", "true"),
					resource.TestCheckResourceAttrWith("atlassian-ops_api_integration.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["region"] != "US_WEST_2" {
							return fmt.Errorf("'region' expected to get 'US_WEST_2', got '%s'", dataObj["region"])
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("atlassian-ops_api_integration.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["suppressNotifications"] != false {
							return fmt.Errorf("'suppressNotifications' expected to get 'false', got '%s'", dataObj["suppressNotifications"])
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "atlassian-ops_api_integration.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type_specific_properties", "directions", "domains"},
			},
			// Update and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "atlassian-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
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
resource "atlassian-ops_api_integration" "example" {
  name    = "` + apiIntegrationUpdateName + `"
  enabled = false
  type = "AmazonSecurityHub"
  team_id = atlassian-ops_team.example.id
  type_specific_properties = jsonencode({
      suppressNotifications: true
      region: "US_WEST_2"
    })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "name", apiIntegrationUpdateName),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "type", "AmazonSecurityHub"),
					resource.TestCheckResourceAttrPair("atlassian-ops_api_integration.example", "team_id", "atlassian-ops_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-ops_api_integration.example", "enabled", "false"),
					resource.TestCheckResourceAttrWith("atlassian-ops_api_integration.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["region"] != "US_WEST_2" {
							return fmt.Errorf("'region' expected to get 'US_WEST_2', got '%s'", dataObj["region"])
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("atlassian-ops_api_integration.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["suppressNotifications"] != true {
							return fmt.Errorf("'suppressNotifications' expected to get 'true', got '%s'", dataObj["suppressNotifications"])
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
