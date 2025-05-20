package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserContactResource(t *testing.T) {

	organizationId := os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID")
	emailPrimary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY")
	emailSecondary := os.Getenv("ATLASSIAN_ACCTEST_EMAIL_SECONDARY")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if organizationId == "" {
				t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
			}
			if emailPrimary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
			if emailSecondary == "" {
				t.Fatal("ATLASSIAN_ACCTEST_EMAIL_SECONDARY must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "atlassian-operations_user_contact" "example" {
  method  = "email"
  to      = "kagan@opsgenie.com"
  enabled = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_user_contact.example", "method", "email"),
					resource.TestCheckResourceAttr("atlassian-operations_user_contact.example", "to", "kagan@opsgenie.com"),
					resource.TestCheckResourceAttr("atlassian-operations_user_contact.example", "enabled", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_user_contact.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "atlassian-operations_user_contact" "example" {
  method  = "email"
  to      = "kagan+updated@opsgenie.com"
  enabled = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_user_contact.example", "method", "email"),
					resource.TestCheckResourceAttr("atlassian-operations_user_contact.example", "to", "kagan+updated@opsgenie.com"),
					resource.TestCheckResourceAttr("atlassian-operations_user_contact.example", "enabled", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
