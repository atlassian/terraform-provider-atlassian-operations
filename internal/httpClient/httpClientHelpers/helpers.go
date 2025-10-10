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

func GenerateUserClientRequest(providerModel dto.AtlassianOpsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	switch providerModel.GetProductType() {
	case "jira-service-desk":
		req.SetUrl(fmt.Sprintf("https://%s/rest/api/3/user/", providerModel.GetDomainName()))
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
