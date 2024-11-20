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

	organizationId := os.Getenv("JSM_ACCTEST_ORGANIZATION_ID")
	apiPrimary := os.Getenv("JSM_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("JSM_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if apiPrimary == "" {
				t.Fatal("JSM_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
}

resource "jsm-ops_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}

resource "jsm-ops_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = jsm-ops_team.example.id
  type = "API"
  enabled = true
  type_specific_properties = jsonencode({
    suppressNotifications: false
  })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "name", apiIntegrationName),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "type", "API"),
					resource.TestCheckResourceAttrPair("jsm-ops_api_integration.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "enabled", "true"),
					resource.TestCheckResourceAttrWith("jsm-ops_api_integration.example", "type_specific_properties", func(value string) error {
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
				ResourceName:            "jsm-ops_api_integration.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type_specific_properties", "directions", "domains"},
			},
			// Update and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
}
resource "jsm-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}
resource "jsm-ops_api_integration" "example" {
  name    = "` + apiIntegrationUpdateName + `"
  team_id = jsm-ops_team.example.id
  type = "API"
  enabled = false
  type_specific_properties = jsonencode({
    suppressNotifications: true
  })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "name", apiIntegrationUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "type", "API"),
					resource.TestCheckResourceAttrPair("jsm-ops_api_integration.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "enabled", "false"),
					resource.TestCheckResourceAttrWith("jsm-ops_api_integration.example", "type_specific_properties", func(value string) error {
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

	organizationId := os.Getenv("JSM_ACCTEST_ORGANIZATION_ID")
	apiPrimary := os.Getenv("JSM_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("JSM_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if apiPrimary == "" {
				t.Fatal("JSM_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
}

resource "jsm-ops_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}

resource "jsm-ops_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  enabled = true
  type = "AmazonSecurityHub"
  team_id = jsm-ops_team.example.id
  type_specific_properties = jsonencode({
      suppressNotifications: false
      region: "US_WEST_2"
    })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "name", apiIntegrationName),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "type", "AmazonSecurityHub"),
					resource.TestCheckResourceAttrPair("jsm-ops_api_integration.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "enabled", "true"),
					resource.TestCheckResourceAttrWith("jsm-ops_api_integration.example", "type_specific_properties", func(value string) error {
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
					resource.TestCheckResourceAttrWith("jsm-ops_api_integration.example", "type_specific_properties", func(value string) error {
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
				ResourceName:            "jsm-ops_api_integration.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type_specific_properties", "directions", "domains"},
			},
			// Update and Read testing
			{
				ExpectNonEmptyPlan: true,
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + apiPrimary + `"
}
resource "jsm-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}
resource "jsm-ops_api_integration" "example" {
  name    = "` + apiIntegrationUpdateName + `"
  enabled = false
  type = "AmazonSecurityHub"
  team_id = jsm-ops_team.example.id
  type_specific_properties = jsonencode({
      suppressNotifications: true
      region: "US_WEST_2"
    })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "name", apiIntegrationUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "type", "AmazonSecurityHub"),
					resource.TestCheckResourceAttrPair("jsm-ops_api_integration.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_api_integration.example", "enabled", "false"),
					resource.TestCheckResourceAttrWith("jsm-ops_api_integration.example", "type_specific_properties", func(value string) error {
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
					resource.TestCheckResourceAttrWith("jsm-ops_api_integration.example", "type_specific_properties", func(value string) error {
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
