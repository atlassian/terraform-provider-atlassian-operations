terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_notification_rule" "example" {
  name        = "Critical Incident Alert"
  action_type = "create-alert"
  enabled     = true

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

  steps = [
    {
      send_after = 15
      enabled    = true
      contact = {
        method = "mobile"
        to     = "olive - Android"
      }
    }
  ]

  repeat = {
    loop_after = 30
    enabled    = true
  }
}