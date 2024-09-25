terraform {
  required_providers {
    jsmops = {
      source = "registry.terraform.io/atlassian/jsm-ops-terraform-provider"
    }
  }
}

provider "jsmops" {
	cloud_id = "3a015c30-bac7-4abc-97a1-50c1feea188a"
	domain_name="iozkaya-us.jira-dev.com"
	username = "iozkaya@atlassian.com"
	password = "<YOUR_TOKEN_HERE>"
}

data "jsmops_user" "example" {
	email_address = "iozkaya@atlassian.com"
}

output "example" {
	value = "data.jsmops_user.example"
}
