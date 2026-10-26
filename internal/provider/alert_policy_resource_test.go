package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccAlertPolicyResource(t *testing.T) {
	alertPolicyName := uuid.NewString()
	alertPolicyUpdateName := uuid.NewString()
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

resource "atlassian-operations_alert_policy" "test" {
  name        = "` + alertPolicyName + `"
  description = "Test alert policy description"
  team_id     = atlassian-operations_team.example.id
  type        = "alert"
  enabled     = true
  message     = "Test alert message"
  order       = 1

  filter = {
    type = "match-any-condition"
    conditions = [
      {
        field          = "message"
        not            = false
        operation      = "contains"
        expected_value = "error"
        order         = 0
      }
    ]
  }

  time_restriction = {
    enabled = true
    time_restrictions = [
      {
        start_hour   = 9
        start_minute = 0
        end_hour     = 17
        end_minute   = 0
      }
    ]
  }

  responders = [
    {
      type = "team"
      id   = atlassian-operations_team.example.id
    }
  ]

  actions = ["action1"]
  tags    = ["acceptance", "test"]
  details = {
    priority = "P1"
    category = "error"
  }

  continue               = false
  update_priority        = true
  priority_value         = "P1"
  keep_original_responders = true
  keep_original_details    = true
  keep_original_actions    = true
  keep_original_tags       = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "name", alertPolicyName),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "description", "Test alert policy description"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "type", "alert"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Test alert message"),
					resource.TestCheckResourceAttrPair("atlassian-operations_alert_policy.test", "team_id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.type", "match-any-condition"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.field", "message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.operation", "contains"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.expected_value", "error"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.time_restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "responders.0.type", "team"),
					resource.TestCheckResourceAttrPair("atlassian-operations_alert_policy.test", "responders.0.id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "actions.0", "action1"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "tags.1", "test"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "tags.0", "acceptance"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "details.priority", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "details.category", "error"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "continue", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_responders", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_details", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_actions", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_tags", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_alert_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_alert_policy.test"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_alert_policy.test"].Primary.Attributes["team_id"],
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
  display_name = "` + teamName + `"
  description = "Updated team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_alert_policy" "test" {
  name        = "` + alertPolicyUpdateName + `"
  description = "Updated alert policy description"
  team_id     = atlassian-operations_team.example.id
  type        = "alert"
  enabled     = false
  message     = "Updated alert message"

  filter = {
    type = "match-all-conditions"
    conditions = [
      {
        field          = "alias"
        not            = true
        operation      = "equals"
        expected_value = "resolved"
        order         = 0
      }
    ]
  }

  time_restriction = {
    enabled = true
    time_restrictions = [
      {
        start_hour   = 0
        start_minute = 0
        end_hour     = 23
        end_minute   = 59
      }
    ]
  }

  responders = [
    {
      type = "team"
      id   = atlassian-operations_team.example.id
    }
  ]

  actions = ["action2"]

  continue               = true
  update_priority        = true
  priority_value         = "P2"
  keep_original_responders = true
  keep_original_details    = false
  keep_original_actions    = true
  keep_original_tags       = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "name", alertPolicyUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "description", "Updated alert policy description"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "type", "alert"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "enabled", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Updated alert message"),
					resource.TestCheckResourceAttrPair("atlassian-operations_alert_policy.test", "team_id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.type", "match-all-conditions"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.field", "alias"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.not", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.operation", "equals"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.expected_value", "resolved"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.time_restrictions.0.start_hour", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.time_restrictions.0.end_hour", "23"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "responders.0.type", "team"),
					resource.TestCheckResourceAttrPair("atlassian-operations_alert_policy.test", "responders.0.id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "actions.0", "action2"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "continue", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_responders", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_details", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_actions", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_tags", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAlertPolicyResource_Global(t *testing.T) {
	alertPolicyName := uuid.NewString()
	alertPolicyUpdateName := uuid.NewString()
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

resource "atlassian-operations_alert_policy" "test" {
  name        = "` + alertPolicyName + `"
  description = "Test alert policy description"
  type        = "alert"
  enabled     = true
  message     = "Test alert message"

  filter = {
    type = "match-any-condition"
    conditions = [
      {
        field          = "message"
        not            = false
        operation      = "contains"
        expected_value = "error"
        order         = 0
      }
    ]
  }

  time_restriction = {
    enabled = true
    time_restrictions = [
      {
        start_hour   = 9
        start_minute = 0
        end_hour     = 17
        end_minute   = 0
      }
    ]
  }

  responders = [
    {
      type = "team"
      id   = atlassian-operations_team.example.id
    }
  ]

  actions = ["action1"]
  tags    = ["acceptance", "test"]
  details = {
    priority = "P1"
    category = "error"
  }

  continue               = false
  update_priority        = true
  priority_value         = "P1"
  keep_original_responders = true
  keep_original_details    = true
  keep_original_actions    = true
  keep_original_tags       = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "name", alertPolicyName),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "description", "Test alert policy description"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "type", "alert"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Test alert message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.type", "match-any-condition"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.field", "message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.operation", "contains"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.expected_value", "error"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.time_restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "responders.0.type", "team"),
					resource.TestCheckResourceAttrPair("atlassian-operations_alert_policy.test", "responders.0.id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "actions.0", "action1"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "tags.1", "test"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "tags.0", "acceptance"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "details.priority", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "details.category", "error"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "continue", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_responders", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_details", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_actions", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_tags", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_alert_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_alert_policy.test"].Primary.ID,
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
  display_name = "` + teamName + `"
  description = "Updated team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_alert_policy" "test" {
  name        = "` + alertPolicyUpdateName + `"
  description = "Updated alert policy description"
  type        = "alert"
  enabled     = false
  message     = "Updated alert message"

  filter = {
    type = "match-all-conditions"
    conditions = [
      {
        field          = "alias"
        not            = true
        operation      = "equals"
        expected_value = "resolved"
        order         = 0
      }
    ]
  }

  time_restriction = {
    enabled = true
    time_restrictions = [
      {
        start_hour   = 0
        start_minute = 0
        end_hour     = 23
        end_minute   = 59
      }
    ]
  }

  responders = [
    {
      type = "team"
      id   = atlassian-operations_team.example.id
    }
  ]

  actions = ["action2"]

  continue               = true
  update_priority        = true
  priority_value         = "P2"
  keep_original_responders = true
  keep_original_details    = false
  keep_original_actions    = true
  keep_original_tags       = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "name", alertPolicyUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "description", "Updated alert policy description"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "type", "alert"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "enabled", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Updated alert message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.type", "match-all-conditions"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.field", "alias"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.not", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.operation", "equals"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "filter.conditions.0.expected_value", "resolved"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.time_restrictions.0.start_hour", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "time_restriction.time_restrictions.0.end_hour", "23"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "responders.0.type", "team"),
					resource.TestCheckResourceAttrPair("atlassian-operations_alert_policy.test", "responders.0.id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "actions.0", "action2"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "continue", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_responders", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_details", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_actions", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "keep_original_tags", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testAccAlertPolicyPriorityConfig renders a minimal team-scoped alert policy used by the
// priority_value regression tests. priorityAttrs is injected verbatim so each step can vary
// update_priority / priority_value (or omit priority_value entirely).
func testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName, message, priorityAttrs string) string {
	return providerConfig + `
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

resource "atlassian-operations_alert_policy" "test" {
  name        = "` + alertPolicyName + `"
  team_id     = atlassian-operations_team.example.id
  type        = "alert"
  enabled     = true
  message     = "` + message + `"

  filter = {
    type = "match-any-condition"
    conditions = [
      {
        field          = "message"
        not            = false
        operation      = "contains"
        expected_value = "error"
        order          = 0
      }
    ]
  }

  ` + priorityAttrs + `
}`
}

// TestAccAlertPolicyResource_UpdatePriorityFalseWithoutValue reproduces JSDCLOUD-19677: with
// update_priority = false and no priority_value, the API omits priorityValue, and priority_value
// must stay null in every plan instead of becoming (known after apply). The second step changes an
// unrelated attribute because that is what triggers the framework to re-plan unset computed values.
func TestAccAlertPolicyResource_UpdatePriorityFalseWithoutValue(t *testing.T) {
	alertPolicyName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Test alert message", `update_priority = false`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("atlassian-operations_alert_policy.test", tfjsonpath.New("priority_value"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "false"),
					resource.TestCheckNoResourceAttr("atlassian-operations_alert_policy.test", "priority_value"),
				),
			},
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Updated alert message", `update_priority = false`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("atlassian-operations_alert_policy.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("atlassian-operations_alert_policy.test", tfjsonpath.New("priority_value"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Updated alert message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "false"),
					resource.TestCheckNoResourceAttr("atlassian-operations_alert_policy.test", "priority_value"),
				),
			},
		},
	})
}

// TestAccAlertPolicyResource_UpdatePriorityTrueWithValue verifies an explicit priority_value is
// stored as configured, survives an unrelated update, and re-plans empty.
func TestAccAlertPolicyResource_UpdatePriorityTrueWithValue(t *testing.T) {
	alertPolicyName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	priorityAttrs := `update_priority = true
  priority_value  = "P2"`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Test alert message", priorityAttrs),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
				),
			},
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Updated alert message", priorityAttrs),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("atlassian-operations_alert_policy.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("atlassian-operations_alert_policy.test", tfjsonpath.New("priority_value"), knownvalue.StringExact("P2")),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Updated alert message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
				),
			},
		},
	})
}

// TestAccAlertPolicyResource_UpdatePriorityTrueToFalse verifies switching from
// update_priority = true / priority_value = "P2" to update_priority = false with no value plans
// priority_value as null (matching what the API returns) and settles to an empty plan.
func TestAccAlertPolicyResource_UpdatePriorityTrueToFalse(t *testing.T) {
	alertPolicyName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Test alert message", `update_priority = true
  priority_value  = "P2"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
				),
			},
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Test alert message", `update_priority = false`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("atlassian-operations_alert_policy.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("atlassian-operations_alert_policy.test", tfjsonpath.New("priority_value"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "false"),
					resource.TestCheckNoResourceAttr("atlassian-operations_alert_policy.test", "priority_value"),
				),
			},
		},
	})
}

// TestAccAlertPolicyResource_UpdatePriorityTrueValueOmittedFromConfig guards the existing behaviour for
// update_priority = true when priority_value is not in config but the server already holds one (for
// example a policy created outside Terraform). The API rejects creating update_priority = true without a
// priority value, so the policy is created with "P2" first and the value is then dropped from config.
// The message also changes so an update is planned: priority_value must keep the server value from state
// instead of becoming (known after apply) or null.
func TestAccAlertPolicyResource_UpdatePriorityTrueValueOmittedFromConfig(t *testing.T) {
	alertPolicyName := uuid.NewString()
	teamName := uuid.NewString()

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Test alert message", `update_priority = true
  priority_value  = "P2"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
				),
			},
			{
				Config: testAccAlertPolicyPriorityConfig(organizationId, emailPrimary, teamName, alertPolicyName,
					"Updated alert message", `update_priority = true`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("atlassian-operations_alert_policy.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("atlassian-operations_alert_policy.test", tfjsonpath.New("priority_value"), knownvalue.StringExact("P2")),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "message", "Updated alert message"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "update_priority", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_alert_policy.test", "priority_value", "P2"),
				),
			},
		},
	})
}
