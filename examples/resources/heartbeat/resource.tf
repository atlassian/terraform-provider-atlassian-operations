terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

# This example demonstrates how to create a heartbeat in Atlassian Operations.
resource "atlassian-operations_heartbeat" "example" {
  name          = "api-health-check"
  description   = "Monitors the health of our API service"
  interval      = 5
  interval_unit = "minutes"
  enabled       = true
  team_id       = "team-123"
  
  # Alert configuration if heartbeat is missed
  alert_message  = "API service is not responding"
  alert_tags     = ["critical", "api", "infrastructure"]
  alert_priority = "P1"
} 