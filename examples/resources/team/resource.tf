terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_team" "example" {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  description     = "This is a team created by Terraform"
  display_name    = "Terraform Team"
  team_type       = "MEMBER_INVITE"
  member = [
    {
      account_id = "XXXXXX:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    }
  ]
}
