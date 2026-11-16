variable "team_id" {
  type = string
}

locals {
  policies = {
    critical = { order = 1 }
    warning  = { order = 2 }
  }
  filter = {
    type = "match-any-condition"
    conditions = [{
      field          = "message"
      operation      = "contains"
      expected_value = "policy-ordering-example"
      not            = false
      order          = 0
    }]
  }
}

# Opt-in coordinated ordering: keep top-level order UNSET on both policy
# resources below. Set positions only in policy_ordering.policies[*].order.
# Do not configure both, even with matching values. Existing configs that use
# per-policy order can keep it without adding these ordering resources.
# The filter condition's nested order above is unrelated and stays unchanged.
# Existing resource update behavior is unchanged; review compatibility limits.
resource "atlassian-operations_alert_policy" "demo" {
  for_each = local.policies
  team_id  = var.team_id
  type     = "alert"
  name     = "${each.key} alert policy"
  enabled  = true
  message  = "{{message}}"
  filter   = local.filter
}

resource "atlassian-operations_notification_policy" "demo" {
  for_each = local.policies
  team_id  = var.team_id
  type     = "notification"
  name     = "${each.key} notification policy"
  enabled  = true
  filter   = local.filter
}

# For global alerts, omit team_id here AND on the alert policy resources.
resource "atlassian-operations_policy_ordering" "alerts" {
  policy_type = "alert"
  team_id     = var.team_id
  policies = [for key, policy in local.policies : {
    policy_id = atlassian-operations_alert_policy.demo[key].id
    order     = policy.order
  }]
}

resource "atlassian-operations_policy_ordering" "notifications" {
  policy_type = "notification"
  team_id     = var.team_id
  policies = [for key, policy in local.policies : {
    policy_id = atlassian-operations_notification_policy.demo[key].id
    order     = policy.order
  }]
}
