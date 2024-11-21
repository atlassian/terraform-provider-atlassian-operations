package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the client is properly configured.
	// Use environment variables to configure the client.
	providerConfig = `
provider "atlassian-ops" {
	cloud_id = ""
	domain_name = ""
	username = ""
	password = ""
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
		"atlassian-ops": providerserver.NewProtocol6WithError(New("test")()),
	}
)
