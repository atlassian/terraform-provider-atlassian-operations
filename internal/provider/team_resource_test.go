package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamResource(t *testing.T) {
	teamName := uuid.NewString()
	teamUpdateName := uuid.NewString()

	organizationId := os.Getenv("JSM_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("JSM_ACCTEST_EMAIL_PRIMARY")
	emailSecondary := os.Getenv("JSM_ACCTEST_EMAIL_SECONDARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("JSM_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("JSM_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
			if emailSecondary == "" {
				t.Fatal("JSM_ACCTEST_EMAIL_SECONDARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "jsm-ops_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_team.example", "display_name", teamName),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "description", "team description"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "organization_id", organizationId),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "team_type", "MEMBER_INVITE"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.add_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.remove_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.update_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.delete_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_team.example", "member.*.account_id", "data.jsm-ops_user.test1", "account_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_team.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["jsm-ops_team.example"].Primary.ID +
							"," +
							state.RootModule().Resources["jsm-ops_team.example"].Primary.Attributes["organization_id"],
						nil
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

data "jsm-ops_user" "test2" {
	email_address = "` + emailSecondary + `"
}

resource "jsm-ops_team" "example" {
  display_name = "` + teamUpdateName + `"
  description = "team description_edited"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.jsm-ops_user.test1.account_id
    },
	{
       account_id = data.jsm-ops_user.test2.account_id
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_team.example", "display_name", teamUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "description", "team description_edited"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "organization_id", organizationId),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "team_type", "MEMBER_INVITE"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.add_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.remove_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.update_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.delete_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_team.example", "member.*.account_id", "data.jsm-ops_user.test1", "account_id"),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_team.example", "member.*.account_id", "data.jsm-ops_user.test2", "account_id"),
				),
			},
			{
				Config: providerConfig + `
data "jsm-ops_user" "test1" {
	email_address = "` + emailPrimary + `"
}

resource "jsm-ops_team" "example" {
  display_name = "` + teamUpdateName + `"
  description = "team description_edited"
  organization_id = "` + organizationId + `"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = data.jsm-ops_user.test1.account_id
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_team.example", "display_name", teamUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "description", "team description_edited"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "organization_id", organizationId),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "team_type", "MEMBER_INVITE"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.add_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.remove_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.update_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.delete_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_team.example", "member.*.account_id", "data.jsm-ops_user.test1", "account_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
