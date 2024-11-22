terraform {
  required_providers {
    atlassian-operations = {
      source = "registry.terraform.io/atlassian/atlassian-operations"
    }
  }
}

resource "atlassian-operations_api_integration" "example" {
  name    = "api integration"
  enabled = true
  type = "AmazonSecurityHub"
  team_id = "81577279-b6ed-4dae-8bf4-e3119fbdf046"
  type_specific_properties = jsonencode({
      suppressNotifications: false
      securityHubIamRoleArn: "arn:aws:iam::416306766477:role/jsmSecurityHubRole"
      region: "AP_SOUTH_1"
    })
}
