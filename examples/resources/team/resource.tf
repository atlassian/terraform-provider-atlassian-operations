terraform {
  required_providers {
    atlassian-ops = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

provider "atlassian-ops" {
  cloud_id    = "3a015c30-bac7-4abc-97a1-50c1feea188a"
  domain_name = "iozkaya-us.jira-dev.com"
  username    = "iozkaya@atlassian.com"
  password    = "<YOUR_TOKEN_HERE>"
}

resource "atlassian-ops_team" "example" {
  organization_id = "0j238a02-kja5-1jka-75j3-82a3dccj366j"
  description = "This is a team created by Terraform"
  display_name = "Terraform Team"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"
    }
  ]
}
