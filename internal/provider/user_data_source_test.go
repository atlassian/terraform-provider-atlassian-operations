package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig +
					`
						data "jsm-ops_user" "test" {
							email_address = "iozkaya@atlassian.com"
						}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data source
					// Verify all attributes are set
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "account_id", "712020:a933b550-3862-441c-ac99-e78ae6dacbcb"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "account_type", "atlassian"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "active", "true"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "application_roles.#", "1"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "application_roles.0.key", "jira-servicedesk"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "application_roles.0.name", "Jira Service Desk"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "avatar_urls.a_16x16", "https://secure.gravatar.com/avatar/bd16cb645843f29c2eea49dfdf0dfd9d?d=https%3A%2F%2Favatar-management--avatars.us-west-2.staging.public.atl-paas.net%2Fdefault-avatar-0.png"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "avatar_urls.a_24x24", "https://secure.gravatar.com/avatar/bd16cb645843f29c2eea49dfdf0dfd9d?d=https%3A%2F%2Favatar-management--avatars.us-west-2.staging.public.atl-paas.net%2Fdefault-avatar-0.png"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "avatar_urls.a_32x32", "https://secure.gravatar.com/avatar/bd16cb645843f29c2eea49dfdf0dfd9d?d=https%3A%2F%2Favatar-management--avatars.us-west-2.staging.public.atl-paas.net%2Fdefault-avatar-0.png"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "avatar_urls.a_48x48", "https://secure.gravatar.com/avatar/bd16cb645843f29c2eea49dfdf0dfd9d?d=https%3A%2F%2Favatar-management--avatars.us-west-2.staging.public.atl-paas.net%2Fdefault-avatar-0.png"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "display_name", "İbrahim Aral Özkaya"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "groups.#", "3"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "groups.0.name", "jira-servicemanagement-users-iozkaya-us"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "groups.0.group_id", "da699b7a-9d18-4148-84a2-134130203ae2"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "groups.0.self", "https://iozkaya-us.jira-dev.com/rest/api/3/group?groupId=da699b7a-9d18-4148-84a2-134130203ae2"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "locale", "en_US"),
					resource.TestCheckResourceAttr("data.jsm-ops_user.test", "timezone", "Europe/Istanbul"),
				),
			},
		},
	})
}
