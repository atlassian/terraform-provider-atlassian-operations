terraform {
  required_providers {
    jsm-ops = {
      source = "registry.terraform.io/atlassian/jsm-ops"
    }
  }
}

resource "jsm-ops_email_integration" "example" {
  name    = "email integration"
  description = "This is an email integration"
  enabled = true
  team_id = "002af28e-bfff-4aeb-80fb-78f0debfd5df"
  type_specific_properties = {
    email_username: "iaral",
    suppress_notifications: false
  }
}
