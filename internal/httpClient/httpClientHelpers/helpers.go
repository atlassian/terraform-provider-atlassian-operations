package httpClientHelpers

import (
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
)

func GenerateJsmOpsClientRequest(providerModel dto.JsmopsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	if providerModel.GetIsStaging() {
		req.SetUrl(fmt.Sprintf("https://api.stg.atlassian.com/jsm/ops/api/%s", providerModel.GetCloudId()))
	} else {
		req.SetUrl(fmt.Sprintf("https://api.atlassian.com/jsm/ops/api/%s", providerModel.GetCloudId()))
	}

	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	return req
}

func GenerateTeamsClientRequest(providerModel dto.JsmopsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	req.SetUrl(fmt.Sprintf("https://%s/gateway/api/public/teams/v1/org/", providerModel.GetDomainName()))
	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	return req
}

func GenerateUserClientRequest(providerModel dto.JsmopsProviderModel) *httpClient.Request {
	req := httpClient.NewRequest()
	req.SetUrl(fmt.Sprintf("https://%s/rest/api/3/user/", providerModel.GetDomainName()))
	req.SetRetryCount(providerModel.GetApiRetryCount())
	req.SetRetryWaitTime(providerModel.GetApiRetryWait())
	req.SetRetryMaxWaitTime(providerModel.GetApiRetryWaitMax())
	req.SetBasicAuth(providerModel.GetEmailAddress(), providerModel.GetToken())
	return req
}
