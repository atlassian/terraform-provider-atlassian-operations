terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

# Basic integration action example
resource "atlassian-operations_integration_action" "basic" {
  integration_id = "your-integration-id"
  name          = "Basic Webhook Integration"
  type          = "create"
  domain        = "alert"
  direction     = "incoming"
  group_type    = "forwarding"
  enabled       = true

  filter = {
    conditions_empty = false
    condition_match_type = "match-all"
    conditions = [
      {
        field = "priority"
        operation = "equals"
        expected_value = "P1"
        key = "priority_level"
        not = false
        order = 1
        system_condition = false
      }
    ]
  }

  type_specific_properties = {
    "url" = "https://example.com/webhook"
    "method" = "POST"
  }

  field_mappings = {
    "message" = "{{alert.message}}"
    "status" = "{{alert.status}}"
  }
}

# Advanced integration action example with multiple conditions and custom action mapping
resource "atlassian-operations_integration_action" "advanced" {
  integration_id = "your-integration-id"
  name          = "Advanced Integration Action"
  type          = "update"
  domain        = "alert"
  direction     = "outgoing"
  group_type    = "updating"
  enabled       = true

  filter = {
    conditions_empty = false
    condition_match_type = "match-any-condition"
    conditions = [
      {
        field = "message"
        operation = "contains"
        expected_value = "critical"
        key = "severity"
        not = false
        order = 1
        system_condition = false
      },
      {
        field = "tags"
        operation = "contains-value"
        expected_value = "production"
        key = "environment"
        not = false
        order = 2
        system_condition = false
      }
    ]
  }

  field_mappings = {
    "description" = "{{alert.description}}"
    "priority" = "{{alert.priority}}"
    "source" = "{{alert.source}}"
    "tags" = "{{alert.tags}}"
  }

  action_mapping = {
    type = "custom"
    parameter = {
      "alert_type" = "incident"
      "team_id" = "{{team.id}}"
      "responders" = "{{alert.responders}}"
      "custom_field_1" = "custom_value_1"
      "custom_field_2" = "custom_value_2"
    }
  }
} 