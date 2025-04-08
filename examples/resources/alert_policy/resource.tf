terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}
# This example demonstrates how to create an alert policy in Atlassian Operations.
resource "atlassian-operations_alert_policy" "test" {
  name        = "alertPolicyName"
  description = "Test alert policy description"
  team_id     = "team-id"
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
      id   = "team-id"
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
}