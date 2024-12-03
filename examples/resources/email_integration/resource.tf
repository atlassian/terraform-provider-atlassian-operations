terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_email_integration" "example" {
  name    = "email integration"
  enabled = true
  team_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  type_specific_properties = {
    email_username : "xxxxx",
    suppress_notifications : false
  }
}
