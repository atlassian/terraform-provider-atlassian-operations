package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationPolicyResource(t *testing.T) {
	notificationPolicyName := uuid.NewString()
	notificationPolicyUpdateName := uuid.NewString()
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

resource "atlassian-operations_notification_policy" "test" {
  name        = "` + notificationPolicyName + `"
  type = "notification"
  description = "Test notification policy description"
  team_id     = atlassian-operations_team.example.id
  enabled     = true
  order       = 0.0

  filter = {
    type = "match-all-conditions"
    conditions = [
      {
        field          = "priority"
        not            = false
        operation      = "equals"
        expected_value = "P1"
        order          = 1
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

  auto_restart_action = {
    wait_duration   = 60
    max_repeat_count = 2
    duration_format = "minutes"
  }

  auto_close_action = {
    wait_duration   = 120
    duration_format = "minutes"
  }

  deduplication_action = {
    deduplication_action_type = "valueBased"
    frequency                = 2
    count_value_limit        = 5
    wait_duration            = 1
    duration_format          = "minutes"
  }

  suppress = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "name", notificationPolicyName),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "description", "Test notification policy description"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttrPair("atlassian-operations_notification_policy.test", "team_id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "order", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.type", "match-all-conditions"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.field", "priority"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.not", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.operation", "equals"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.expected_value", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.order", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.start_minute", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.end_minute", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_restart_action.wait_duration", "60"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_restart_action.max_repeat_count", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_restart_action.duration_format", "minutes"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_close_action.wait_duration", "120"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_close_action.duration_format", "minutes"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "deduplication_action.deduplication_action_type", "valueBased"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "deduplication_action.frequency", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "deduplication_action.count_value_limit", "5"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "deduplication_action.wait_duration", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "deduplication_action.duration_format", "minutes"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "suppress", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_notification_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["atlassian-operations_notification_policy.test"].Primary.ID +
							"," +
							state.RootModule().Resources["atlassian-operations_notification_policy.test"].Primary.Attributes["team_id"],
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

resource "atlassian-operations_notification_policy" "test" {
  name        = "` + notificationPolicyUpdateName + `"
  type = "notification"
  description = "Updated notification policy description"
  team_id     = atlassian-operations_team.example.id
  enabled     = false

  filter = {
    type = "match-any-condition"
    conditions = [
      {
        field          = "details"
        not            = false
        operation      = "contains-key"
        expected_value = "P1"
        order          = 1
      },
      {
        field          = "message"
        not            = false
        operation      = "equals"
        expected_value = "P1"
        order          = 2
      }
    ]
  }

  time_restriction = {
    enabled = true
    time_restrictions = [
      {
        start_hour = 0
        start_minute = 0
        end_hour = 23
        end_minute = 59
      }
    ]
  }

  auto_restart_action = {
    max_repeat_count = 20
    wait_duration = 3
	duration_format = "hours"
  }

  auto_close_action = {
    wait_duration = 7
	duration_format = "days"
  }

  delay_action = {
    delay_time      = {
		hours = 0
		minutes = 15
	}
    delay_option    = "nextTime"
    wait_duration   = 5
    duration_format = "minutes"
  }

  suppress = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "name", notificationPolicyUpdateName),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "description", "Updated notification policy description"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "enabled", "false"),
					resource.TestCheckResourceAttrPair("atlassian-operations_notification_policy.test", "team_id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.type", "match-any-condition"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.field", "details"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.not", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.operation", "contains-key"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.expected_value", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.0.order", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.1.field", "message"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.1.not", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.1.operation", "equals"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.1.expected_value", "P1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "filter.conditions.1.order", "2"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.start_hour", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.start_minute", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.end_hour", "23"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "time_restriction.time_restrictions.0.end_minute", "59"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_restart_action.max_repeat_count", "20"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_restart_action.wait_duration", "3"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_restart_action.duration_format", "hours"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_close_action.wait_duration", "7"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "auto_close_action.duration_format", "days"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "delay_action.delay_time.hours", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "delay_action.delay_time.minutes", "15"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "delay_action.delay_option", "nextTime"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "delay_action.wait_duration", "5"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "delay_action.duration_format", "minutes"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_policy.test", "suppress", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
