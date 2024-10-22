package provider

import (
	"github.com/google/uuid"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleDataSource(t *testing.T) {
	teamName := uuid.NewString()
	scheduleName := uuid.NewString()

	organizationId := os.Getenv("JSM_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("JSM_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("JSM_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("JSM_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "jsm-ops_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}

resource "jsm-ops_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = jsm-ops_team.example.id
}

data "jsm-ops_schedule" "test" {
	depends_on = ["jsm-ops_schedule.example"]
	name = "` + scheduleName + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttrPair("data.jsm-ops_schedule.test", "id", "jsm-ops_schedule.example", "id"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_schedule.test", "name", "jsm-ops_schedule.example", "name"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_schedule.test", "description", "jsm-ops_schedule.example", "description"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_schedule.test", "timezone", "jsm-ops_schedule.example", "timezone"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_schedule.test", "enabled", "jsm-ops_schedule.example", "enabled"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_schedule.test", "team_id", "jsm-ops_team.example", "id"),
				),
			},
		},
	})
}
