// Copyright (c) HashiCorp, Inc.

package provider

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamRoleResource(t *testing.T) {
	// Get team ID from environment variable
	teamId := os.Getenv("TEST_TEAM_ID")
	if teamId == "" {
		t.Skip("TEST_TEAM_ID environment variable must be set for acceptance tests")
	}

	// Generate unique names for the resources
	roleName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if teamId == "" {
				t.Fatal("TEST_TEAM_ID must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "atlassian-operations_team_role" "test" {
  team_id = "` + teamId + `"
  name    = "` + roleName + `"
  granted_rights = [
    "access-reports",
    "edit-schedules",
    "edit-escalations"
  ]
  disallowed_rights = [
    "delete-schedules",
    "delete-escalations"
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "team_id", teamId),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "name", roleName),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "granted_rights.#", "3"),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "disallowed_rights.#", "2"),
					resource.TestCheckResourceAttrSet("atlassian-operations_team_role.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "atlassian-operations_team_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing - change name and rights
			{
				Config: providerConfig + `
resource "atlassian-operations_team_role" "test" {
  team_id = "` + teamId + `"
  name    = "Updated ` + roleName + `"
  granted_rights = [
    "access-reports",
    "edit-schedules",
    "delete-schedules",
    "edit-escalations",
    "delete-escalations",
    "edit-policies"
  ]
  disallowed_rights = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "team_id", teamId),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "name", "Updated "+roleName),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "granted_rights.#", "6"),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test", "disallowed_rights.#", "0"),
					resource.TestCheckResourceAttrSet("atlassian-operations_team_role.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccTeamRoleResource_AllRights(t *testing.T) {
	// Get team ID from environment variable
	teamId := os.Getenv("TEST_TEAM_ID")
	if teamId == "" {
		t.Skip("TEST_TEAM_ID environment variable must be set for acceptance tests")
	}

	// Generate unique name for the resource
	roleName := fmt.Sprintf("admin-%s", uuid.NewString())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if teamId == "" {
				t.Fatal("TEST_TEAM_ID must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create role with all rights granted
			{
				Config: providerConfig + `
resource "atlassian-operations_team_role" "test_admin" {
  team_id = "` + teamId + `"
  name    = "` + roleName + `"
  granted_rights = [
    "access-reports",
    "delete-escalations",
    "delete-heartbeats",
    "delete-integrations",
    "delete-maintenance",
    "delete-policies",
    "delete-routing-rules",
    "delete-schedules",
    "delete-team-roles",
    "edit-escalations",
    "edit-heartbeats",
    "edit-integrations",
    "edit-maintenance",
    "edit-policies",
    "edit-routing-rules",
    "edit-schedules",
    "edit-team-roles"
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test_admin", "team_id", teamId),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test_admin", "name", roleName),
					resource.TestCheckResourceAttr("atlassian-operations_team_role.test_admin", "granted_rights.#", "17"),
					resource.TestCheckResourceAttrSet("atlassian-operations_team_role.test_admin", "id"),
			},
		},
	})
}

func TestAccTeamRoleResource_OverlappingRights(t *testing.T) {
	// Get team ID from environment variable
	teamId := os.Getenv("TEST_TEAM_ID")
	if teamId == "" {
		t.Skip("TEST_TEAM_ID environment variable must be set for acceptance tests")
	}

	// Generate unique name for the resource
	roleName := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if teamId == "" {
				t.Fatal("TEST_TEAM_ID must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test that overlapping rights are rejected
			{
				Config: providerConfig + `
resource "atlassian-operations_team_role" "test_invalid" {
  team_id = "` + teamId + `"
  name    = "` + roleName + `"
  granted_rights = [
    "access-reports",
    "edit-schedules"
  ]
  disallowed_rights = [
    "edit-schedules",
    "delete-schedules"
  ]
}`,
				ExpectError: regexp.MustCompile("the following rights appear in both granted_rights and disallowed_rights.*edit-schedules"),
			},
		},
	})
}
