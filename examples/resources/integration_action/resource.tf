terraform {
  required_providers {
    atlassian-operations = {
      source = "atlassian/atlassian-operations"
    }
  }
}

# Basic integration action example
resource "atlassian-operations_integration_action" "basic" {
  integration_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" # Replace with your integration ID
  name          = "Example Integration Action"
  type          = "create"
  domain        = "alert"
  direction     = "incoming"
  group_type    = "forwarding"
  enabled       = true

  filter = {
    conditions_empty = false
    condition_match_type = "match-all-conditions"
    conditions = [
      {
        field = "message"
        operation = "matches"
        expected_value = "critical alert"
        not = false
        order = 0
        system_condition = false
      }
    ]
  }

  type_specific_properties = jsonencode({
    appendAttachments: true
    keepActionsFromPayload: true
    keepExtraPropertiesFromPayload: true
    keepRespondersFromPayload: false
    keepTagsFromPayload: true
  })

  field_mappings = jsonencode({
    actions: []
    alias: ""
    description: "{{alert.description}}"
    details: {}
    entity: ""
    message: "{{alert.message}}"
    note: ""
    responders: [{
      id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" # Replace with your responder ID
      type: "team"
    }]
    priority: "{{alert.priority}}"
    source: ""
    tags: []
    user: ""
  })
}