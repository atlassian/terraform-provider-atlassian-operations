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

data "jsmops_team" "example" {
	organization_id = "0j238a02-kja5-1jka-75j3-82a3dccj366j"
	team_id = "ef72bc0a-6f28-43d3-87e3-783ae3ed0ffa"
}

output "example" {
	value = "data.jsmops_team.example"
}
