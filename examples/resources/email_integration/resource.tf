terraform {
  required_providers {
    atlassian-ops = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-ops_email_integration" "example" {
  name    = "email integration"
  enabled = true
  team_id = "002af28e-bfff-4aeb-80fb-78f0debfd5df"
  type_specific_properties = {
    email_username: "iaral",
    suppress_notifications: false
  }
}
