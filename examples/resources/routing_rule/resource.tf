terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_routing_rule" "example2" {
  team_id    = "3b7188be-91ff-40e8-8952-b0a83c7dfc58"
  name       = "Example Routing Rule"
  timezone   = "Europe/Berlin"

  criteria = {
    type = "match-all"
    conditions = [
      {
        field          = "message"
        operation      = "matches"
        expected_value = "my critical alert"
      }
    ]
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
    type = "none"
  }
}

resource "atlassian-operations_routing_rule" "example" {
  team_id    = "3b7188be-91ff-40e8-8952-b0a83c7dfc58"
  name       = "Example Routing Rule"
  timezone   = "Europe/Istanbul"

  criteria = {
      type = "match-all"
      conditions = [
        {
        field          = "message"
        operation      = "matches"
        expected_value = "my critical alert"
        }
      ]
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
    type = "none"
  }
}