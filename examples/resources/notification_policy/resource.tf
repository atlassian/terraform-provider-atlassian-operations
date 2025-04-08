# Configure the provider
terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

# Advanced notification policy with all features
resource "atlassian-operations_notification_policy" "advanced" {
  name        = "Advanced Notification Policy"
  description = "An advanced notification policy with all features"
  team_id     = "3b7188be-91ff-40e8-8952-b0a83c7dfc58"
  type = "notification"
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
}