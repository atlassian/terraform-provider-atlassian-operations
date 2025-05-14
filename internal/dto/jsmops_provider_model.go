package dto

import "time"

type AtlassianOpsProviderModel struct {
	productType     string
	cloudId         string
	domainName      string
	emailAddress    string
	token           string
	orgAdminToken   string
	apiRetryCount   int
	apiRetryWait    time.Duration
	apiRetryWaitMax time.Duration
	isStaging       bool
}

func NewAtlassianOpsProviderModel(
	productType string,
	cloudId string,
	domainName string,
	emailAddress string,
	token string,
	orgAdminToken string,
	apiRetryCount int,
	apiRetryWait time.Duration,
	apiRetryWaitMax time.Duration,
	isStaging bool,
) AtlassianOpsProviderModel {
	return AtlassianOpsProviderModel{
		productType:     productType,
		cloudId:         cloudId,
		domainName:      domainName,
		emailAddress:    emailAddress,
		token:           token,
		orgAdminToken:   orgAdminToken,
		apiRetryCount:   apiRetryCount,
		apiRetryWait:    apiRetryWait,
		apiRetryWaitMax: apiRetryWaitMax,
		isStaging:       isStaging,
	}
}

func (receiver AtlassianOpsProviderModel) GetProductType() string {
	return receiver.productType
}

func (receiver AtlassianOpsProviderModel) GetCloudId() string {
	return receiver.cloudId
}

func (receiver AtlassianOpsProviderModel) GetDomainName() string {
	return receiver.domainName
}

func (receiver AtlassianOpsProviderModel) GetEmailAddress() string {
	return receiver.emailAddress
}

func (receiver AtlassianOpsProviderModel) GetToken() string {
	return receiver.token
}

func (receiver AtlassianOpsProviderModel) GetOrgAdminToken() string {
	return receiver.orgAdminToken
}

func (receiver AtlassianOpsProviderModel) GetApiRetryCount() int {
	return receiver.apiRetryCount
}

func (receiver AtlassianOpsProviderModel) GetApiRetryWait() time.Duration {
	return receiver.apiRetryWait
}

func (receiver AtlassianOpsProviderModel) GetApiRetryWaitMax() time.Duration {
	return receiver.apiRetryWaitMax
}

func (receiver AtlassianOpsProviderModel) GetIsStaging() bool {
	return receiver.isStaging
}
