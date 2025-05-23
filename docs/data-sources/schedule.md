---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "atlassian-operations_schedule Data Source - atlassian-operations"
subcategory: ""
description: |-
  Schedule data source
---

# atlassian-operations_schedule (Data Source)

Schedule data source



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the schedule. This is used to look up the schedule and must be unique within your organization.

### Read-Only

- `description` (String) A detailed description of the schedule's purpose and coverage. This helps team members understand the schedule's role.
- `enabled` (Boolean) Indicates whether the schedule is currently active and can be used for rotations and assignments.
- `id` (String) The unique identifier of the schedule. This is automatically generated when the schedule is created.
- `team_id` (String) The unique identifier of the team that owns this schedule. Used for access control and organization.
- `timezone` (String) The timezone in IANA format (e.g., 'America/New_York') that this schedule operates in. All times in the schedule are interpreted in this timezone.
