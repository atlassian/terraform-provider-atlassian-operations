terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_schedule" "example" {
  name    = "tf"
  team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
}
