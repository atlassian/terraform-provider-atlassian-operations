package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the clientConfiguration is properly configured.
	// Use environment variables to configure the clientConfiguration.
	providerConfig = `
provider "atlassian-operations" {
	api_retry_count = 5
	api_retry_wait = 15
	api_retry_wait_max = 100
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"atlassian-operations": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func testAccPreCheck(t *testing.T) {
	if os.Getenv("ATLASSIAN_ACCTEST_ORGANIZATION_ID") == "" {
		t.Fatal("ATLASSIAN_ACCTEST_ORGANIZATION_ID must be set for acceptance tests")
	}
	if os.Getenv("ATLASSIAN_ACCTEST_EMAIL_PRIMARY") == "" {
		t.Fatal("ATLASSIAN_ACCTEST_EMAIL_PRIMARY must be set for acceptance tests")
	}
}
