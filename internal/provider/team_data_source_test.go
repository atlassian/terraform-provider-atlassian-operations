package provider

import (
	"github.com/google/uuid"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamDataSource(t *testing.T) {
	teamName := uuid.NewString()

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
				Config: providerConfig +
					`
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

						data "jsm-ops_team" "test" {
							organization_id = "` + organizationId + `"
							id = jsm-ops_team.example.id
						}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "organization_id", "jsm-ops_team.example", "organization_id"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "description", "jsm-ops_team.example", "description"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "display_name", "jsm-ops_team.example", "display_name"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "team_type", "jsm-ops_team.example", "team_type"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "user_permissions.update_team", "jsm-ops_team.example", "user_permissions.update_team"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "user_permissions.delete_team", "jsm-ops_team.example", "user_permissions.delete_team"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "user_permissions.add_members", "jsm-ops_team.example", "user_permissions.add_members"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "user_permissions.remove_members", "jsm-ops_team.example", "user_permissions.remove_members"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "member.#", "jsm-ops_team.example", "member.#"),
					resource.TestCheckResourceAttrPair("data.jsm-ops_team.test", "member.0.account_id", "jsm-ops_team.example", "member.0.account_id"),
				),
			},
		},
	})
}
