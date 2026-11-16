---
page_title: "atlassian-operations_policy_ordering Resource - atlassian-operations"
subcategory: ""
description: |-
  Coordinates alert or notification policy ordering after policy creation.
---

# atlassian-operations_policy_ordering (Resource)

Orders alert policies in a team or globally, or notification policies in a team. Individual policy resources manage content and existence; this resource manages the ordered prefix of their list. It creates no backend object.

**This resource is opt-in.** Upgrading the provider does not enable it, change existing resource behavior, or require migration. Existing configurations can continue using per-policy `order` without adding an ordering resource.

## Configure order in one place

When adopting this resource, **omit top-level `order` from every referenced alert or notification policy** and set positions only in this resource's `policies[*].order`. Do not set both, even to matching values: both resources can move policies and cause inconsistent-state errors or repeated diffs. This also applies to values supplied by variables or module defaults. The provider does not enforce this rule across resource blocks.

Use one ordering resource per scope and policy type. Filter-condition `order` is separate and should remain unchanged. An omitted top-level order can still appear as a computed value in Terraform state; that is not an explicit configuration assignment and does not require state removal.

ID references make Terraform wait for the referenced operations before this resource runs. It moves the policies sequentially and verifies the result. Input list order does not matter; numeric `order` determines the sequence.

## Example Usage

Given `var.team_id` and a `local.policies` map whose entries contain unique desired positions `1..N`, keep the policy and ordering settings separate:

```terraform
resource "atlassian-operations_alert_policy" "demo" {
  for_each = local.policies
  team_id  = var.team_id
  type     = "alert"
  name     = "${each.key} alert policy"
  enabled  = true
  message  = "{{message}}"
  # No top-level order here, including through a module default.
}

resource "atlassian-operations_policy_ordering" "alerts" {
  policy_type = "alert"
  team_id     = var.team_id
  policies = [
    for key, policy in local.policies : {
      policy_id = atlassian-operations_alert_policy.demo[key].id
      order     = policy.order
    }
  ]
}
```

For notifications, set `policy_type = "notification"` and reference notification policy IDs. For global alerts, omit `team_id` from both the policies and ordering resource. See the [complete alert and notification example](https://bitbucket.org/jira-service-management/terraform-provider-atlassian-operations/src/main/examples/resources/policy_ordering/resource.tf).

## Schema

### Required

- `policy_type` (String): `alert` or `notification`. Changing type replaces the logical controller without deleting policies.
- `policies` (List of Object): Desired positions. Each entry requires:
  - `policy_id` (String): Unique ID belonging to the selected scope and policy type.
  - `order` (Number): Unique integer position. N entries must cover `1..N`, without gaps.

### Optional

- `team_id` (String): Required for notifications; omit only for global alerts. Changing scope replaces the logical controller without deleting policies.

### Read-only

- `id` (String): `team/<team_id>/<policy_type>` or `global/alert`; not a remote policy ID.
- `observed_order` (List of String): Actual first N IDs. Terraform detects differences from the desired sequence and reconciles on apply.

## Compatibility and operational limits

Existing alert and notification resources retain their schema, create/update flow, per-policy `order`, state, and import behavior. To opt in for policies you already manage:

1. Keep their Terraform addresses and IDs. Remove their explicit top-level `order` assignments, including any supplied by a module.
2. Add one ordering resource referencing those IDs, with the desired positions `1..N`. Do not temporarily configure both ordering approaches.
3. Review a full plan before applying. The new ordering resource does not itself require policies to be recreated, but existing resource rules can still require replacement. Do not remove individual policies from state as part of adoption.

If you do not opt in, leave your existing configuration as-is. Removing the ordering resource later does not automatically configure per-policy ordering or restore a previous sequence.

**Limit on subsequent content changes:** the existing resources may move a policy using its planned/computed position even when `order` is not explicitly configured. ID references sequence scheduled operations, but do not force the ordering resource to update when its inputs and refreshed sequence are unchanged. An upstream policy can therefore change the order after the controller's refresh; run a subsequent full plan/apply to detect and reconcile that drift. An upstream create/update error can also prevent the controller from running. This addition does not claim to fix those existing-resource behaviors.

Listed policies occupy the first N positions. Unlisted policies remain afterward, preserving relative order during reconciliation. An empty list controls no positions. Destroying this resource leaves policies and their current order intact.

Moves are **not transactional**: an error can leave a partial sequence. Resolve the error and rerun a full plan/apply. Apply is bounded to five minutes, including the provider-configured HTTP retries. After moves, verification performs at most four list reads, one second apart when the sequence differs; it does not keep issuing moves. Incomplete or inconsistent list snapshots fail rather than report a successful order. Targeted applies that omit this resource, UI changes, separate Terraform states, and other API writers are not coordinated by it.

## Import

Import adopts the complete current sequence. Align `policies` with that sequence or explicitly choose the desired prefix on the next apply.

```shell
terraform import atlassian-operations_policy_ordering.alerts 'team/<team_id>/alert'
terraform import atlassian-operations_policy_ordering.notifications 'team/<team_id>/notification'
terraform import atlassian-operations_policy_ordering.global_alerts 'global/alert'
```
