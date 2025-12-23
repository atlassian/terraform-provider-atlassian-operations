# Copyright (c) HashiCorp, Inc.

# Basic team role with read-only permissions
resource "atlassian-operations_team_role" "read_only" {
  team_id = "12345678-1234-1234-1234-123456789012"
  name    = "Read Only"

  granted_rights = [
    "access-reports"
  ]
  disallowed_rights = [
    "edit-schedules",
    "delete-schedules",
    "edit-escalations",
    "delete-escalations"
  ]
}

# Team role with schedule management permissions
resource "atlassian-operations_team_role" "schedule_manager" {
  team_id = "12345678-1234-1234-1234-123456789012"
  name    = "Schedule Manager"

  granted_rights = [
    "access-reports",
    "edit-schedules",
    "edit-routing-rules"
  ]
  disallowed_rights = [
    "delete-schedules"
  ]
}

# Admin role with all permissions
resource "atlassian-operations_team_role" "admin" {
  team_id = "12345678-1234-1234-1234-123456789012"
  name    = "Team Admin"

  granted_rights = [
    "access-reports",
    "delete-escalations",
    "delete-heartbeats",
    "delete-integrations",
    "delete-maintenance",
    "delete-policies",
    "delete-routing-rules",
    "delete-schedules",
    "delete-team-roles",
    "edit-escalations",
    "edit-heartbeats",
    "edit-integrations",
    "edit-maintenance",
    "edit-policies",
    "edit-routing-rules",
    "edit-schedules",
    "edit-team-roles"
  ]
}
