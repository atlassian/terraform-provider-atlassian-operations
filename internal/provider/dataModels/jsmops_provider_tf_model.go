package dataModels

import "github.com/hashicorp/terraform-plugin-framework/types"

type JsmopsProviderTfModel struct {
	CloudId         types.String `tfsdk:"cloud_id"`
	DomainName      types.String `tfsdk:"domain_name"`
	EmailAddress    types.String `tfsdk:"email_address"`
	Token           types.String `tfsdk:"token"`
	ApiRetryCount   types.Int32  `tfsdk:"api_retry_count"`
	ApiRetryWait    types.Int32  `tfsdk:"api_retry_wait"`
	ApiRetryWaitMax types.Int32  `tfsdk:"api_retry_wait_max"`
}
