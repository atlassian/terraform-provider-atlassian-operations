# A service account's scoped API token
provider "atlassian-operations" {
  cloud_id        = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  domain_name     = "xxxx.atlassian.net"
  email_address   = "my-bot@serviceaccount.atlassian.com"
  token           = "<SERVICE_ACCOUNT_SCOPED_API_TOKEN>"
  is_scoped_token = true
}
