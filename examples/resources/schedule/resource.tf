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
  password    = "<YOUR_TOKEN>"
}

resource "atlassian-ops_schedule" "example" {
  name    = "tf"
  team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
}
