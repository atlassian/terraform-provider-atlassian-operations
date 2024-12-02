terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

# Get Atlassian User by email address
data "atlassian-operations_user" "example" {
  email_address = "email@example.com"
}
