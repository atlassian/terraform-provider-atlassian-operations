# A user's scoped API token
provider "atlassian-operations" {
  cloud_id      = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  domain_name   = "xxxx.atlassian.net"
  email_address = "email@example.com"
  token         = "<YOUR_SCOPED_API_TOKEN>"       # Jira + JSM Operations scopes
  teams_token   = "<YOUR_SCOPED_TEAMS_API_TOKEN>" # Teams scopes
  auth_type     = "scoped"
}
