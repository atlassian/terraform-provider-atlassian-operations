package provider

import (
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestMaintenanceResourceSchema_TeamId(t *testing.T) {
	teamIdAttr, ok := schemaAttributes.MaintenanceResourceAttributes["team_id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected team_id to be schema.StringAttribute, got %T", schemaAttributes.MaintenanceResourceAttributes["team_id"])
	}
	if !teamIdAttr.Optional {
		t.Errorf("expected team_id to be Optional")
	}
	if !teamIdAttr.Computed {
		t.Errorf("expected team_id to be Computed")
	}
	if len(teamIdAttr.PlanModifiers) == 0 {
		t.Errorf("expected team_id to have PlanModifiers")
	}
}

func TestUserContactResourceSchema_Method(t *testing.T) {
	methodAttr, ok := schemaAttributes.UserContactResourceAttributes["method"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected method to be schema.StringAttribute, got %T", schemaAttributes.UserContactResourceAttributes["method"])
	}
	if !methodAttr.Required {
		t.Errorf("expected method to be Required")
	}
	if len(methodAttr.PlanModifiers) == 0 {
		t.Errorf("expected method to have PlanModifiers")
	}
}
