terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_schedule" "example" {
  name    = "tf"
  team_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
