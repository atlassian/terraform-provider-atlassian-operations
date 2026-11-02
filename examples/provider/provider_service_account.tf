# A service account's scoped API token. auth_type defaults to "scoped" for a
# @serviceaccount.atlassian.com address, so it can stay out of the block.
provider "atlassian-operations" {
  cloud_id      = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  domain_name   = "xxxx.atlassian.net"
  email_address = "my-bot@serviceaccount.atlassian.com"
  token         = "<SERVICE_ACCOUNT_SCOPED_API_TOKEN>" # all scopes on one token; no teams_token needed
}
