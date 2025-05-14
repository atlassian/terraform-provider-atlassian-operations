package provider

import (
	"github.com/google/uuid"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamDataSource(t *testing.T) {
	teamName := uuid.NewString()

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
				Config: providerConfig +
					`
						data "atlassian-operations_user" "test1" {
							email_address = "` + emailPrimary + `"
							organization_id = "` + organizationId + `"
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

						data "atlassian-operations_team" "test" {
							organization_id = "` + organizationId + `"
							id = atlassian-operations_team.example.id
						}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "id", "atlassian-operations_team.example", "id"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "organization_id", "atlassian-operations_team.example", "organization_id"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "description", "atlassian-operations_team.example", "description"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "display_name", "atlassian-operations_team.example", "display_name"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "team_type", "atlassian-operations_team.example", "team_type"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "user_permissions.update_team", "atlassian-operations_team.example", "user_permissions.update_team"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "user_permissions.delete_team", "atlassian-operations_team.example", "user_permissions.delete_team"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "user_permissions.add_members", "atlassian-operations_team.example", "user_permissions.add_members"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "user_permissions.remove_members", "atlassian-operations_team.example", "user_permissions.remove_members"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "member.#", "atlassian-operations_team.example", "member.#"),
					resource.TestCheckResourceAttrPair("data.atlassian-operations_team.test", "member.0.account_id", "atlassian-operations_team.example", "member.0.account_id"),
				),
			},
		},
	})
}
