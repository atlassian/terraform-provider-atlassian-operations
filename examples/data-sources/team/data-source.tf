terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

# Get Atlassian Operations Teams by organization ID and team ID
data "atlassian-operations_team" "example" {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  id              = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

