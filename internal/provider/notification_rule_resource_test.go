package provider

import (
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationRuleCreateAlertResource(t *testing.T) {
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

resource "atlassian-operations_notification_rule" "example" {
  name        = "Critical Incident Alert"
  action_type = "create-alert"
  enabled     = true

  time_restriction = {
    type = "weekday-and-time-of-day"
    restrictions = [{
      start_day  = "monday"
      end_day    = "friday"
      start_hour = 9
      end_hour   = 17
      start_min  = 0
      end_min    = 0
    }]
  }

  order = 0 
  criteria = {
    type = "match-all"
  }


  steps = [
    {
      send_after = 15
      enabled    = true
      contact = {
        method = "email"
        to     = "` + emailPrimary + `"
      }
    }
  ]

  repeat = {
    loop_after = 60
    enabled    = true
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "name", "Critical Incident Alert"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "action_type", "create-alert"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "order", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.end_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "criteria.type", "match-all"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.send_after", "15"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.method", "email"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.to", emailPrimary),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "repeat.loop_after", "60"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "repeat.enabled", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_notification_rule.example",
				ImportState:       true,
				ImportStateVerify: true,
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

resource "atlassian-operations_notification_rule" "example" {
  name        = "Updated Critical Incident Alert"
  action_type = "create-alert"
  enabled     = false

  time_restriction = {
    type = "time-of-day"
    restriction = {
      start_hour = 8
      end_hour   = 20
      start_min  = 30
      end_min    = 30
    }
  }
  order = 0 
  criteria = {
    type = "match-all"
  }
  steps = [
    {
      send_after = 15
      enabled    = true
      contact = {
        method = "email"
        to     = "` + emailPrimary + `"
      }
    }
  ]

  repeat = {
    loop_after = 30
    enabled    = true
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "name", "Updated Critical Incident Alert"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "action_type", "create-alert"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "enabled", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "order", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.start_hour", "8"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.end_hour", "20"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.start_min", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.end_min", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "criteria.type", "match-all"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.send_after", "15"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.method", "email"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.to", emailPrimary),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "repeat.loop_after", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "repeat.enabled", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccNotificationRuleScheduleStartResource(t *testing.T) {
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

resource "atlassian-operations_notification_rule" "example" {
  name        = "Critical Incident Alert"
  action_type = "schedule-start"
  enabled     = true

  time_restriction = {
    type = "weekday-and-time-of-day"
    restrictions = [{
      start_day  = "monday"
      end_day    = "friday"
      start_hour = 9
      end_hour   = 17
      start_min  = 0
      end_min    = 0
    }]
  }

  criteria = {
    type = "match-all"
  }

  notification_time = [
   "15-minutes-ago",
   "1-hour-ago",
    "1-day-ago"
]

  steps = [
    {
      enabled    = true
      contact = {
        method = "email"
        to     = "` + emailPrimary + `"
      }
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "name", "Critical Incident Alert"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "action_type", "schedule-start"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.type", "weekday-and-time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.start_day", "monday"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.end_day", "friday"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.start_hour", "9"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.end_hour", "17"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.start_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restrictions.0.end_min", "0"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "criteria.type", "match-all"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.method", "email"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.to", emailPrimary),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_notification_rule.example",
				ImportState:       true,
				ImportStateVerify: true,
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

resource "atlassian-operations_notification_rule" "example" {
  name        = "Updated Critical Incident Alert"
  action_type = "schedule-start"
  enabled     = false

  time_restriction = {
    type = "time-of-day"
    restriction = {
      start_hour = 8
      end_hour   = 20
      start_min  = 30
      end_min    = 30
    }
  }

  criteria = {
    type = "match-all"
  }

  notification_time = [
   "15-minutes-ago"
]

  steps = [
    {
      enabled    = true
      contact = {
        method = "email"
        to     = "` + emailPrimary + `"
      }
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "name", "Updated Critical Incident Alert"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "action_type", "schedule-start"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "enabled", "false"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.type", "time-of-day"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.start_hour", "8"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.end_hour", "20"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.start_min", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "time_restriction.restriction.end_min", "30"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "criteria.type", "match-all"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.enabled", "true"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.method", "email"),
					resource.TestCheckResourceAttr("atlassian-operations_notification_rule.example", "steps.0.contact.to", emailPrimary),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
