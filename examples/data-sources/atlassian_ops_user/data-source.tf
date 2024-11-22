terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

data "atlassian-operations_user" "example" {
	email_address = "iozkaya@atlassian.com"
}

output "example" {
	value = "data.atlassian-operations_user.example"
}
