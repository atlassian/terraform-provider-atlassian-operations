package provider

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIntegrationActionResource(t *testing.T) {
	teamName := uuid.NewString()
	apiIntegrationName := uuid.NewString()
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

resource "atlassian-operations_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  type = "API"
  enabled = true
}

resource "atlassian-operations_integration_action" "example" {
  integration_id = atlassian-operations_api_integration.example.id
  name          = "Example Integration Action"
  type          = "create"
  domain        = "alert"
  direction     = "incoming"
  group_type    = "forwarding"
  enabled       = true

  filter = {
    conditions_empty = false
    condition_match_type = "match-all-conditions"
    conditions = [
      {
        field = "message"
        operation = "matches"
        expected_value = "critical alert"
        not = false
        order = 0
        system_condition = false
      }
    ]
  }

  type_specific_properties = jsonencode({
	appendAttachments: true
	keepActionsFromPayload: true
	keepExtraPropertiesFromPayload: true
	keepRespondersFromPayload: false
	keepTagsFromPayload: true
  })

  field_mappings = jsonencode({
	actions: []
	alias: ""
	description: "{{alert.description}}"
	details: {}
	entity: ""
    message: "{{alert.message}}"
	note: ""
	responders: [{
		id: atlassian-operations_team.example.id
		type: "team"
	}]
	priority: "{{alert.priority}}"
	source: ""
	tags: []
	user: ""
  })
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "name", "Example Integration Action"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "type", "create"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "domain", "alert"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "direction", "incoming"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "group_type", "forwarding"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions_empty", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.condition_match_type", "match-all-conditions"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.field", "message"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.operation", "matches"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.expected_value", "critical alert"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.not", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.order", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.system_condition", "false"),
					resource.TestCheckResourceAttrWith("atlassian-operations_integration_action.example", "type_specific_properties", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["appendAttachments"] != true {
							return fmt.Errorf("appendAttachments expected to get 'true', got '%s'", dataObj["appendAttachments"])
						}
						if dataObj["keepActionsFromPayload"] != true {
							return fmt.Errorf("keepActionsFromPayload expected to get 'true', got '%s'", dataObj["keepActionsFromPayload"])
						}
						if dataObj["keepRespondersFromPayload"] != false {
							return fmt.Errorf("keepRespondersFromPayload expected to get 'false', got '%s'", dataObj["keepRespondersFromPayload"])
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("atlassian-operations_integration_action.example", "field_mappings", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["message"] != "{{alert.message}}" {
							return fmt.Errorf("expected to get '{{alert.message}}', got '%s'", dataObj["message"])
						}
						if dataObj["priority"] != "{{alert.priority}}" {
							return fmt.Errorf("expected to get '{{alert.priority}}', got '%s'", dataObj["priority"])
						}
						if dataObj["description"] != "{{alert.description}}" {
							return fmt.Errorf("expected to get '{{alert.description}}', got '%s'", dataObj["description"])
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "atlassian-operations_integration_action.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "group_type"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_integration_action.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_integration_action.example"].Primary.Attributes["integration_id"],
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

resource "atlassian-operations_api_integration" "example" {
  name    = "` + apiIntegrationName + `"
  team_id = atlassian-operations_team.example.id
  type = "API"
  enabled = true
}

resource "atlassian-operations_integration_action" "example" {
  integration_id = atlassian-operations_api_integration.example.id
  name          = "Updated Integration Action"
  type          = "create"
  domain        = "alert"
  direction     = "incoming"
  group_type    = "updating"
  enabled       = false

  filter = {
    conditions_empty = false
    condition_match_type = "match-any-condition"
    conditions = [
      {
        field = "priority"
        operation = "equals"
        expected_value = "P1"
        not = true
        order = 0
      }
    ]
  }

  type_specific_properties = jsonencode({
	appendAttachments: true
	keepActionsFromPayload: true
	keepExtraPropertiesFromPayload: true
	keepRespondersFromPayload: false
	keepTagsFromPayload: true
  })

  field_mappings = jsonencode({
	actions: []
	alias: "{{alert.alias}}"
	description: "{{alert.description}}"
	details: {}
	entity: ""
    message: "{{alert.message}}"
	note: ""
	responders: [{
		id: atlassian-operations_team.example.id
		type: "team"
	}]
	priority: "{{alert.priority}}"
	source: ""
	tags: []
	user: "" 
  })
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "name", "Updated Integration Action"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "type", "create"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "domain", "alert"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "direction", "incoming"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "group_type", "updating"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "enabled", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions_empty", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.condition_match_type", "match-any-condition"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.field", "priority"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.operation", "equals"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.expected_value", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.not", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_integration_action.example", "filter.conditions.0.order", "0"),
					resource.TestCheckResourceAttrWith("atlassian-operations_integration_action.example", "field_mappings", func(value string) error {
						var dataObj map[string]interface{}
						err := json.Unmarshal([]byte(value), &dataObj)
						if err != nil {
							return err
						}
						if dataObj["message"] != "{{alert.message}}" {
							return fmt.Errorf("expected to get '{{alert.message}}', got '%s'", dataObj["message"])
						}
						if dataObj["priority"] != "{{alert.priority}}" {
							return fmt.Errorf("expected to get '{{alert.priority}}', got '%s'", dataObj["priority"])
						}
						if dataObj["description"] != "{{alert.description}}" {
							return fmt.Errorf("expected to get '{{alert.description}}', got '%s'", dataObj["description"])
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
