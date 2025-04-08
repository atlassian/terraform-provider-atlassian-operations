terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

# Example: Create a custom role with basic alert management permissions
resource "atlassian-operations_custom_role" "basic_role" {
  name = "Basic Alert Manager"
  granted_rights = [
    "alert-view",
    "alert-create",
    "alert-acknowledge",
    "alert-add-note",
    "alert-close"
  ]
  disallowed_rights = [
    "alert-delete",
    "alert-assign-ownership",
    "alert-escalate"
  ]
}

# Example: Create a custom role with advanced alert management permissions
resource "atlassian-operations_custom_role" "admin_role" {
  name = "Alert Administrator"
  granted_rights = [
    "alert-view",
    "alert-create",
    "alert-acknowledge",
    "alert-add-note",
    "alert-close",
    "alert-delete",
    "alert-assign-ownership",
    "alert-escalate",
    "alert-update-description",
    "alert-take-ownership",
    "alert-custom-action"
  ]
  disallowed_rights = []
}

# Example: Create a read-only custom role
resource "atlassian-operations_custom_role" "viewer_role" {
  name = "Alert Viewer"
  granted_rights = [
    "alert-view"
  ]
  disallowed_rights = [
    "alert-create",
    "alert-acknowledge",
    "alert-add-note",
    "alert-close",
    "alert-delete",
    "alert-assign-ownership",
    "alert-escalate",
    "alert-update-description",
    "alert-take-ownership",
    "alert-custom-action"
  ]
}