package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	emailPrimary := os.Getenv("JSM_ACCTEST_EMAIL_PRIMARY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			if emailPrimary == "" {
				t.Fatal("JSM_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
			}
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig +
					`
						data "jsm-ops_user" "test" {
							email_address = "` + emailPrimary + `"
						}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttrSet("data.jsm-ops_user.test", "account_id"),
					resource.TestCheckResourceAttrSet("data.jsm-ops_user.test", "account_type"),
					resource.TestCheckResourceAttrSet("data.jsm-ops_user.test", "active"),
					resource.TestCheckResourceAttrSet("data.jsm-ops_user.test", "display_name"),
					resource.TestCheckResourceAttrSet("data.jsm-ops_user.test", "locale"),
					resource.TestCheckResourceAttrSet("data.jsm-ops_user.test", "timezone"),
				),
			},
		},
	})
}
