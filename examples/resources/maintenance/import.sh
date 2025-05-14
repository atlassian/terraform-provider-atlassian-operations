#!/bin/bash

# Import a global maintenance window
terraform import atlassian-operations_maintenance.example maintenance_id

# Import a team-specific maintenance window
# terraform import atlassian-operations_maintenance.example maintenance_id,team_id 