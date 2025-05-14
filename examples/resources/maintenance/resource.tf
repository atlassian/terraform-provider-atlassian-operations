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
  start_date  = "2023-06-15T10:00:00Z"
  end_date    = "2023-06-15T14:00:00Z"
  
  # Optional team ID if it's a team-specific maintenance window
  # team_id     = "your-team-id"
  
  # Rules define what entities are affected during the maintenance
  rules {
    state = "disabled"
    entity {
      id   = "integration-1234"
      type = "integration"
    }
  }
  
  rules {
    state = "disabled"
    entity {
      id   = "policy-5678"
      type = "policy"
    }
  }
} 