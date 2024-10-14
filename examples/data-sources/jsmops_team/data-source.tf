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

data "jsm-ops_team" "example" {
	organization_id = "0j238a02-kja5-1jka-75j3-82a3dccj366j"
	id = "6848b028-db3d-4d1e-9a3c-d3513354ce61"
}

output "example" {
	value = "data.jsm-ops_team.example"
}
