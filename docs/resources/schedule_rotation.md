---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "atlassian-operations_schedule_rotation Resource - atlassian-operations"
subcategory: ""
description: |-
  
---

# atlassian-operations_schedule_rotation (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `schedule_id` (String) The ID of the schedule
- `start_date` (String) The start date of the rotation
- `type` (String) The type of the rotation

### Optional

- `end_date` (String) The end date of the rotation
- `length` (Number) The length of the rotation
- `name` (String) The name of the rotation
- `participants` (Attributes List) The participants of the rotation (see [below for nested schema](#nestedatt--participants))
- `time_restriction` (Attributes) (see [below for nested schema](#nestedatt--time_restriction))

### Read-Only

- `id` (String) The ID of the rotation

<a id="nestedatt--participants"></a>
### Nested Schema for `participants`

Required:

- `type` (String) The type of the participant

Optional:

- `id` (String) The ID of the participant


<a id="nestedatt--time_restriction"></a>
### Nested Schema for `time_restriction`

Required:

- `type` (String) The type of the time restriction

Optional:

- `restriction` (Attributes) (see [below for nested schema](#nestedatt--time_restriction--restriction))
- `restrictions` (Attributes List) The restrictions of the time restriction (see [below for nested schema](#nestedatt--time_restriction--restrictions))

<a id="nestedatt--time_restriction--restriction"></a>
### Nested Schema for `time_restriction.restriction`

Required:

- `end_hour` (Number) The end hour of the restriction
- `end_min` (Number) The end minute of the restriction
- `start_hour` (Number) The start hour of the restriction
- `start_min` (Number) The start minute of the restriction


<a id="nestedatt--time_restriction--restrictions"></a>
### Nested Schema for `time_restriction.restrictions`

Required:

- `end_day` (String) The end day of the restriction
- `end_hour` (Number) The end hour of the restriction
- `end_min` (Number) The end minute of the restriction
- `start_day` (String) The start day of the restriction
- `start_hour` (Number) The start hour of the restriction
- `start_min` (Number) The start minute of the restriction
