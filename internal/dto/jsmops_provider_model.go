package dto

import "time"

type JsmopsProviderModel struct {
	cloudId         string
	domainName      string
	emailAddress    string
	token           string
	apiRetryCount   int
	apiRetryWait    time.Duration
	apiRetryWaitMax time.Duration
	isStaging       bool
}

func NewJsmopsProviderModel(
	cloudId string,
	domainName string,
	emailAddress string,
	token string,
	apiRetryCount int,
	apiRetryWait time.Duration,
	apiRetryWaitMax time.Duration,
	isStaging bool,
) JsmopsProviderModel {
	return JsmopsProviderModel{
		cloudId:         cloudId,
		domainName:      domainName,
		emailAddress:    emailAddress,
		token:           token,
		apiRetryCount:   apiRetryCount,
		apiRetryWait:    apiRetryWait,
		apiRetryWaitMax: apiRetryWaitMax,
		isStaging:       isStaging,
	}
}

func (receiver JsmopsProviderModel) GetCloudId() string {
	return receiver.cloudId
}

func (receiver JsmopsProviderModel) GetDomainName() string {
	return receiver.domainName
}

func (receiver JsmopsProviderModel) GetEmailAddress() string {
	return receiver.emailAddress
}

func (receiver JsmopsProviderModel) GetToken() string {
	return receiver.token
}

func (receiver JsmopsProviderModel) GetApiRetryCount() int {
	return receiver.apiRetryCount
}

func (receiver JsmopsProviderModel) GetApiRetryWait() time.Duration {
	return receiver.apiRetryWait
}

func (receiver JsmopsProviderModel) GetApiRetryWaitMax() time.Duration {
	return receiver.apiRetryWaitMax
}

func (receiver JsmopsProviderModel) GetIsStaging() bool {
	return receiver.isStaging
}
