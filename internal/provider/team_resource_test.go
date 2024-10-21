package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamResource(t *testing.T) {
	teamName := uuid.NewString()
	teamUpdateName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "jsm-ops_team" "example" {
  display_name = "` + teamName + `"
  description = "team description"
  organization_id = "0j238a02-kja5-1jka-75j3-82a3dccj366j"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_team.example", "display_name", teamName),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "description", "team description"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "organization_id", "0j238a02-kja5-1jka-75j3-82a3dccj366j"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "team_type", "MEMBER_INVITE"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.add_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.remove_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.update_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.delete_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.0.account_id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
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
resource "jsm-ops_team" "example" {
  display_name = "` + teamUpdateName + `"
  description = "team description_edited"
  organization_id = "0j238a02-kja5-1jka-75j3-82a3dccj366j"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"
    },
	{
       account_id = "712020:ce8310ee-7509-41b5-baa5-0c4f74dae467"
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_team.example", "display_name", teamUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "description", "team description_edited"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "organization_id", "0j238a02-kja5-1jka-75j3-82a3dccj366j"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "team_type", "MEMBER_INVITE"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.add_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.remove_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.update_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.delete_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.#", "2"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.0.account_id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.1.account_id", "712020:ce8310ee-7509-41b5-baa5-0c4f74dae467"),
				),
			},
			{
				Config: providerConfig + `
resource "jsm-ops_team" "example" {
  display_name = "` + teamUpdateName + `"
  description = "team description_edited"
  organization_id = "0j238a02-kja5-1jka-75j3-82a3dccj366j"
  team_type = "MEMBER_INVITE"
  member = [
    {
       account_id = "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_team.example", "display_name", teamUpdateName),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "description", "team description_edited"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "organization_id", "0j238a02-kja5-1jka-75j3-82a3dccj366j"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "team_type", "MEMBER_INVITE"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.add_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.remove_members", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.update_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "user_permissions.delete_team", "true"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.#", "1"),
					resource.TestCheckResourceAttr("jsm-ops_team.example", "member.0.account_id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
