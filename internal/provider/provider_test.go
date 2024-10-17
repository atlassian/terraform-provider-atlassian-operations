package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the client is properly configured.
	providerConfig = `
provider "jsm-ops" {
	cloud_id = "3a015c30-bac7-4abc-97a1-50c1feea188a"
	domain_name="iozkaya-us.jira-dev.com"
	username = "iozkaya@atlassian.com"
	password = "<YOUR_TOKEN_HERE>"
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"jsm-ops": providerserver.NewProtocol6WithError(New("test")()),
	}
)
