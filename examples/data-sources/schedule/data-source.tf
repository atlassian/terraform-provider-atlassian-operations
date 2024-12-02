terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

# Get Atlassian Operations Schedule by name
data "atlassian-operations_schedule" "example" {
  name = "Test schedule"
}
