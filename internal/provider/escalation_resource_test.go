package provider

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"testing"
)

func TestAccEscalationResource_Full(t *testing.T) {
	escalationName := uuid.NewString()
	escalationUpdateName := uuid.NewString()

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
			// Create and Read testing
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

resource "jsm-ops_escalation" "example" {
  name    = "` + escalationName + `"
  team_id = jsm-ops_team.example.id
  description = "escalation description"
  rules = [{
	condition = "if-not-acked"
	notify_type = "default"
    delay = 5
    recipient = {
    	id = data.jsm-ops_user.test1.account_id
		type = "user"
    }
  },
  {
	condition = "if-not-closed"
	notify_type = "all"
	delay = 1
	recipient = {
		id = jsm-ops_team.example.id
		type = "team"
    }
  }]
  enabled = true
  repeat = {
  	wait_interval = 5
    count = 10
    reset_recipient_states = true
    close_alert_after_all = true
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "name", escalationName),
					resource.TestCheckResourceAttrPair("jsm-ops_escalation.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "description", "escalation description"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("jsm-ops_escalation.example", "rules.*", map[string]string{
						"condition":   "if-not-acked",
						"notify_type": "default",
						"delay":       "5",
					}),
					resource.TestCheckResourceAttrSet("jsm-ops_escalation.example", "rules.0.recipient.id"),
					resource.TestCheckResourceAttrSet("jsm-ops_escalation.example", "rules.0.recipient.type"),
					resource.TestCheckTypeSetElemNestedAttrs("jsm-ops_escalation.example", "rules.*", map[string]string{
						"condition":   "if-not-closed",
						"notify_type": "all",
						"delay":       "1",
					}),
					resource.TestCheckResourceAttrSet("jsm-ops_escalation.example", "rules.1.recipient.id"),
					resource.TestCheckResourceAttrSet("jsm-ops_escalation.example", "rules.1.recipient.type"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "enabled", "true"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.wait_interval", "5"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.count", "10"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.reset_recipient_states", "true"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.close_alert_after_all", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_escalation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["jsm-ops_escalation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["jsm-ops_escalation.example"].Primary.Attributes["team_id"],
						nil
				},
			},
			// Update and Read testing
			{
				ExpectNonEmptyPlan: true,
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

resource "jsm-ops_escalation" "example" {
  name    = "` + escalationUpdateName + `"
  team_id = jsm-ops_team.example.id
  rules = [{
	condition = "if-not-closed"
	notify_type = "default"
    delay = 1
    recipient = {
    	id = data.jsm-ops_user.test1.account_id
		type = "user"
    }
  }]
  enabled = false
  repeat = {}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "name", escalationUpdateName),
					resource.TestCheckResourceAttrPair("jsm-ops_escalation.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "description", ""),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("jsm-ops_escalation.example", "rules.*", map[string]string{
						"condition":   "if-not-closed",
						"notify_type": "default",
						"delay":       "1",
					}),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_escalation.example", "rules.0.recipient.id", "data.jsm-ops_user.test1", "account_id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.0.recipient.type", "user"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "enabled", "false"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.wait_interval", "0"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.count", "1"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.reset_recipient_states", "false"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "repeat.close_alert_after_all", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccEscalationResource_Minimal(t *testing.T) {
	escalationName := uuid.NewString()

	escalationUpdateName := uuid.NewString()

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
			// Create and Read testing
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

resource "jsm-ops_escalation" "example" {
  name    = "` + escalationName + `"
  team_id = jsm-ops_team.example.id
  rules = [{
	condition = "if-not-acked"
	notify_type = "default"
    delay = 5
    recipient = {
    	id = data.jsm-ops_user.test1.account_id
		type = "user"
    }
  }]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "name", escalationName),
					resource.TestCheckResourceAttrPair("jsm-ops_escalation.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "description", ""),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("jsm-ops_escalation.example", "rules.*", map[string]string{
						"condition":   "if-not-acked",
						"notify_type": "default",
						"delay":       "5",
					}),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_escalation.example", "rules.0.recipient.id", "data.jsm-ops_user.test1", "account_id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.0.recipient.type", "user"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "enabled", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jsm-ops_escalation.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["jsm-ops_escalation.example"].Primary.ID +
							"," +
							state.RootModule().Resources["jsm-ops_escalation.example"].Primary.Attributes["team_id"],
						nil
				},
			},
			// Update and Read testing
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

resource "jsm-ops_escalation" "example" {
  name    = "` + escalationUpdateName + `"
  team_id = jsm-ops_team.example.id
  rules = [{
	condition = "if-not-closed"
	notify_type = "random"
    delay = 1
    recipient = {
    	id = jsm-ops_team.example.id
		type = "team"
    }
  }]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "name", escalationUpdateName),
					resource.TestCheckResourceAttrPair("jsm-ops_escalation.example", "team_id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "description", ""),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("jsm-ops_escalation.example", "rules.*", map[string]string{
						"condition":   "if-not-closed",
						"notify_type": "random",
						"delay":       "1",
					}),
					resource.TestCheckTypeSetElemAttrPair("jsm-ops_escalation.example", "rules.0.recipient.id", "jsm-ops_team.example", "id"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "rules.0.recipient.type", "team"),
					resource.TestCheckResourceAttr("jsm-ops_escalation.example", "enabled", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
