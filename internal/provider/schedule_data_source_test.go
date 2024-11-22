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

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "atlassian-operations_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "atlassian-operations_team" "example" {
  organization_id = "` + organizationId + `"
  description = "This is a team created by Terraform"
  display_name = "` + teamName + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
      account_id = data.atlassian-operations_user.test1.account_id
    }
  ]
}

resource "atlassian-operations_schedule" "example" {
  name    = "` + scheduleName + `"
  team_id = atlassian-operations_team.example.id
}

data "atlassian-operations_schedule" "test" {
	depends_on = ["atlassian-operations_schedule.example"]
	name = "` + scheduleName + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttrPair("data.atlassian-operations_schedule.test", "id", "atlassian-operations_schedule.example", "id"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_schedule.test", "name", "atlassian-operations_schedule.example", "name"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_schedule.test", "description", "atlassian-operations_schedule.example", "description"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_schedule.test", "timezone", "atlassian-operations_schedule.example", "timezone"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_schedule.test", "enabled", "atlassian-operations_schedule.example", "enabled"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_schedule.test", "team_id", "atlassian-operations_team.example", "id"),
				),
			},
		},
	})
}
