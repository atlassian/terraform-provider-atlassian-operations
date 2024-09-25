terraform {
  required_providers {
    jsm-ops = {
      source = "registry.terraform.io/atlassian/jsm-ops"
    }
  }
}

provider "jsm-ops" {
	cloud_id = "3a015c30-bac7-4abc-97a1-50c1feea188a"
	domain_name="iozkaya-us.jira-dev.com"
	username = "iozkaya@atlassian.com"
	password = "<YOUR_TOKEN_HERE>"
}

data "jsm-ops_user" "example" {
	email_address = "iozkaya@atlassian.com"
}

output "example" {
	value = "data.jsm-ops_user.example"
}
