terraform {
  required_providers {
    atlassian-ops = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

data "atlassian-ops_team" "example" {
	organization_id = "8aab9c24-60d3-15bc-k703-8b29952kb34a"
	id = "002af28e-bfff-4aeb-80fb-78f0debfd5df"
}

output "example" {
	value = "data.atlassian-ops_team.example"
}