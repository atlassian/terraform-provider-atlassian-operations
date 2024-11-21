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
	password = "<YOUR_TOKEN_HERE>"
}

data "atlassian-ops_user" "example" {
	email_address = "iozkaya@atlassian.com"
}

output "example" {
	value = "data.atlassian-ops_user.example"
}
