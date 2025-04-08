package provider

import (
	"github.com/google/uuid"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomRoleResource(t *testing.T) {
	// Generate unique names for the resources
	roleName := uuid.NewString()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "atlassian-operations_custom_role" "test" {
  name = "` + roleName + `"
  granted_rights = [
    "alert-acknowledge",
    "alert-action",
    "alert-add-note",
    "alert-close",
    "alert-create"
  ]
  disallowed_rights = [
    "alert-delete"
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "name", roleName),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.#", "5"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.0", "alert-acknowledge"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.1", "alert-action"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.2", "alert-add-note"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.3", "alert-close"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.4", "alert-create"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "disallowed_rights.#", "1"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "disallowed_rights.0", "alert-delete"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_custom_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "atlassian-operations_custom_role" "test" {
  name = "Updated ` + roleName + `"
  granted_rights = [
    "alert-acknowledge",
    "alert-action",
    "alert-add-note",
    "alert-close",
    "alert-create",
    "alert-delete"
  ]
  disallowed_rights = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "name", "Updated "+roleName),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.#", "6"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.0", "alert-acknowledge"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.1", "alert-action"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.2", "alert-add-note"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.3", "alert-close"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.4", "alert-create"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "granted_rights.5", "alert-delete"),
					resource.TestCheckResourceAttr("atlassian-operations_custom_role.test", "disallowed_rights.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
