package httpClientHelpers

import (
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
)

func GenerateJsmOpsClientRequest(providerModel dto.AtlassianOpsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()

	switch providerModel.GetProductType() {
	case "jira-service-desk":
		req.SetUrl(fmt.Sprintf("%s/jsm/ops/api/%s", getAtlassianApiDomain(providerModel.GetIsStaging()), providerModel.GetCloudId()))
	case "compass":
		req.SetUrl(fmt.Sprintf("%s/compass/cloud/%s/ops", getAtlassianApiDomain(providerModel.GetIsStaging()), providerModel.GetCloudId()))
	}

	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	return req
}

// GenerateTeamsClientRequest builds a request for the Teams REST API, the classic path. The REST API does not accept
// scoped tokens, so it always authenticates with token; teams_token is for the GraphQL client only.
func GenerateTeamsClientRequest(providerModel dto.AtlassianOpsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	req.SetUrl(fmt.Sprintf("https://%s/gateway/api/public/teams/v1/org/", providerModel.GetDomainName()))
	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	return req
}

func GenerateServiceClientRequest(providerModel dto.AtlassianOpsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	req.SetUrl(fmt.Sprintf("%s/jsm/api/%s", getAtlassianApiDomain(providerModel.GetIsStaging()), providerModel.GetCloudId()))
	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	return req
}

// GenerateAGGClientRequest builds a request for the Atlassian GraphQL Gateway, which manages teams for scoped API
// tokens; the Teams REST API rejects them. Like the other Atlassian APIs it lives on api.atlassian.com
// (api.stg.atlassian.com when ATLASSIAN_OPS_STAGING=1), not on the site domain. It authenticates with teams_token
// when set (see GetTeamsToken), otherwise with token.
func GenerateAGGClientRequest(providerModel dto.AtlassianOpsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	req.SetUrl(fmt.Sprintf("%s/gateway/api/graphql", getAtlassianApiDomain(providerModel.GetIsStaging())))
	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetTeamsToken())
	return req
}

func GenerateUserClientRequest(providerModel dto.AtlassianOpsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	switch providerModel.GetProductType() {
	case "jira-service-desk":
		if providerModel.IsScopedToken() {
			// Scoped API tokens are rejected by the user API on the site domain; it must be reached through the gateway.
			req.SetUrl(fmt.Sprintf("%s/ex/jira/%s/rest/api/3/user/", getAtlassianApiDomain(providerModel.GetIsStaging()), providerModel.GetCloudId()))
		} else {
			req.SetUrl(fmt.Sprintf("https://%s/rest/api/3/user/", providerModel.GetDomainName()))
		}
		req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	default:
		req.SetUrl(fmt.Sprintf("%s/admin/v2/orgs/", getAtlassianApiDomain(providerModel.GetIsStaging())))
		req.SetBearerAuth(providerModel.GetOrgAdminToken())
	}

	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	return req
}

func getAtlassianApiDomain(isStaging bool) string {
	if isStaging {
		return "https://api.stg.atlassian.com"
	}
	return "https://api.atlassian.com"
}
