# Import adopts the full current sequence. Use only the matching scope/type.
terraform import atlassian-operations_policy_ordering.alerts 'team/<team_id>/alert'
terraform import atlassian-operations_policy_ordering.notifications 'team/<team_id>/notification'
# For a separately configured global alert controller:
# terraform import atlassian-operations_policy_ordering.global_alerts 'global/alert'
