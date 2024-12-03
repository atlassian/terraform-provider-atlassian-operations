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
  type    = "AmazonSecurityHub"
  team_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  type_specific_properties = jsonencode({
    suppressNotifications : false
    securityHubIamRoleArn : "arn:aws:iam::xxxxxxxxxxxx:role/jsmSecurityHubRole"
    region : "XX_XXXX_X"
  })
}
