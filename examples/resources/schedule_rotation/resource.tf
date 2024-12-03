terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_schedule_rotation" "example" {
  schedule_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  name        = "tf"
  start_date  = "2023-11-10T05:00:00Z"
  type        = "weekly"
}
