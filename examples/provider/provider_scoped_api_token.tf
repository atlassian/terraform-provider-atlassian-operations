# A user's scoped API token
provider "atlassian-operations" {
  cloud_id        = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  domain_name     = "xxxx.atlassian.net"
  email_address   = "email@example.com"
  token           = "<YOUR_SCOPED_API_TOKEN>"
  is_scoped_token = true
}
