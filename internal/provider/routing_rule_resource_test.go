package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoutingRuleResource(t *testing.T) {
	scheduleName := uuid.NewString()
	escalationName := uuid.NewString()
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

resource "atlassian-operations_escalation" "example" {
  name    = "` + escalationName + `"
  team_id = atlassian-operations_team.example.id
  description = "escalation description"
  rules = [{
	condition = "if-not-acked"
	notify_type = "default"
    delay = 5
    recipient = {
    	id = data.atlassian-operations_user.test1.account_id
		type = "user"
    }
  },
  {
	condition = "if-not-closed"
	notify_type = "all"
	delay = 1
	recipient = {
		id = atlassian-operations_team.example.id
		type = "team"
    }
  }]
  enabled = true
  repeat = {
  	wait_interval = 5
    count = 10
    reset_recipient_states = true
    close_alert_after_all = true
  }
}

resource "atlassian-operations_routing_rule" "example" {
  team_id    = atlassian-operations_team.example.id
  name       = "Example Routing Rule"
  timezone   = "Europe/Istanbul"

  criteria = {
      type = "match-all"
    }

  time_restriction = {
    type = "time-of-day"
    restriction = {
      start_hour = 9
      end_hour = 17
      start_min = 0
      end_min = 0
    }
  }

  notify = {
    type = "escalation"
	id   = atlassian-operations_escalation.example.id
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "name", "Example Routing Rule"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "order", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "timezone", "Europe/Istanbul"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "criteria.type", "match-all"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restriction.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restriction.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restriction.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restriction.end_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "notify.type", "escalation"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_routing_rule.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_routing_rule.example"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_routing_rule.example"].Primary.Attributes["team_id"],
						nil
				},
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
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

resource "atlassian-operations_escalation" "example" {
  name    = "` + escalationName + `"
  team_id = atlassian-operations_team.example.id
  description = "escalation description"
  rules = [{
	condition = "if-not-acked"
	notify_type = "default"
    delay = 5
    recipient = {
    	id = data.atlassian-operations_user.test1.account_id
		type = "user"
    }
  },
  {
	condition = "if-not-closed"
	notify_type = "all"
	delay = 1
	recipient = {
		id = atlassian-operations_team.example.id
		type = "team"
    }
  }]
  enabled = true
  repeat = {
  	wait_interval = 5
    count = 10
    reset_recipient_states = true
    close_alert_after_all = true
  }
}

resource "atlassian-operations_routing_rule" "example" {
  team_id    = atlassian-operations_team.example.id
  name       = "Example Routing Rule"
  timezone   = "Europe/Istanbul"

  criteria = {
      type = "match-all-conditions"
      conditions = [
        {
        field          = "message"
        operation      = "matches"
        expected_value = "my critical alert"
        }
      ]
    }

  time_restriction = {
    type = "weekday-and-time-of-day"
    restrictions = [{
      start_day = "monday"
      end_day = "friday"
      start_hour = 9
      end_hour = 17
      start_min = 0
      end_min = 0
    }]
  }

  notify = {
    type = "none"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "name", "Example Routing Rule"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "criteria.type", "match-all-conditions"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "criteria.conditions.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "criteria.conditions.0.field", "message"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "criteria.conditions.0.operation", "matches"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "criteria.conditions.0.expected_value", "my critical alert"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "time_restriction.restrictions.0.end_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_routing_rule.example", "notify.type", "none"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
