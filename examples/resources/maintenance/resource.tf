terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

# This example demonstrates how to create a maintenance window in Atlassian Operations.
resource "atlassian-operations_maintenance" "example" {
  description = "Planned maintenance window for system upgrades"
  start_date  = "2029-06-16T10:00:00Z"
  end_date    = "2029-06-16T16:00:00Z"
  # Optional team ID if it's a team-specific maintenance window
  # team_id     = "your-team-id"

  rules = [ {
    state = "disabled"
    entity = {
      id   = "integration-1234" # Replace with your integration ID
      type = "integration"
    }
  }
  ]
} 