terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}
# Example: Create an email contact
resource "atlassian-operations_user_contact" "email_contact" {
  method  = "email"
  to      = "kagan2@opsgenie.com"
  enabled = true
}

# Example: Create a voice contact (disabled)
resource "atlassian-operations_user_contact" "voice_contact" {
  method  = "voice"
  to      = "49-5360287176"  # Phone number
  enabled = false  # Contact is created but disabled
}