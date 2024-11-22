terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

data "atlassian-operations_schedule" "example" {
	name = "Test schedule"
}

output "example" {
	value = "data.atlassian-operations_team.example"
}
