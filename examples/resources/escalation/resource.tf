terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_escalation" "example" {
  name        = "escalationName"
  team_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  description = "escalation description"
  rules = [{
    condition   = "if-not-closed"
    notify_type = "all"
    delay       = 1
    recipient = {
      id   = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      type = "team"
    }
  }]
  enabled = true
  repeat = {
    wait_interval          = 5
    count                  = 10
    reset_recipient_states = true
    close_alert_after_all  = true
  }
}
