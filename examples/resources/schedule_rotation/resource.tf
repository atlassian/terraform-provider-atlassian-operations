terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = "df47a95c-f9ae-4ca6-873b-375fcad3cd18"
  name        = "tf"
  start_date  = "2023-11-10T05:00:00Z"
  type        = "weekly"
}
