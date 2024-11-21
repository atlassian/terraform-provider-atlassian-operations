terraform {
  required_providers {
    atlassian-ops = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

provider "atlassian-ops" {
  cloud_id = "3a015c30-bac7-4abc-97a1-50c1feea188a"
  domain_name="iozkaya-us.jira-dev.com"
  username = "iozkaya@atlassian.com"
  password = "<YOUR_TOKEN>"
}

resource "atlassian-ops_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
  name        = "tf"
  start_date  = "2023-11-10T05:00:00Z"
  type        = "weekly"
}
