package dto

import "time"

// Values of the auth_type provider setting.
const (
	// AuthTypeClassic authenticates with a classic (unscoped) API token.
	AuthTypeClassic = "classic"
	// AuthTypeScoped authenticates with a scoped API token of a user or a service account.
	AuthTypeScoped = "scoped"
)

// ServiceAccountEmailSuffix ends the email address of every Atlassian service account. A service account never has a
// classic API token, so auth_type defaults to AuthTypeScoped for such an address.
const ServiceAccountEmailSuffix = "@serviceaccount.atlassian.com"

type AtlassianOpsProviderModel struct {
	productType     string
	cloudId         string
	domainName      string
	emailAddress    string
	token           string
	teamsToken      string
	orgAdminToken   string
	authType        string
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
	teamsToken string,
	orgAdminToken string,
	authType string,
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
		teamsToken:      teamsToken,
		orgAdminToken:   orgAdminToken,
		authType:        authType,
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

// GetTeamsToken returns the token for the Teams GraphQL API (auth_type = scoped): teams_token when set, otherwise token.
// A user's scoped API tokens are created per product and cannot carry the Teams scopes together with the JSM
// Operations scopes, so such a user needs a second token for teams. Service accounts and classic tokens use token.
func (receiver AtlassianOpsProviderModel) GetTeamsToken() string {
	if receiver.teamsToken != "" {
		return receiver.teamsToken
	}
	return receiver.token
}

func (receiver AtlassianOpsProviderModel) GetOrgAdminToken() string {
	return receiver.orgAdminToken
}

// GetAuthType returns how the token authenticates: AuthTypeClassic or AuthTypeScoped.
func (receiver AtlassianOpsProviderModel) GetAuthType() string {
	return receiver.authType
}

// IsScopedToken reports whether the configured token is a scoped API token (auth_type = "scoped") rather than a
// classic (unscoped) one.
func (receiver AtlassianOpsProviderModel) IsScopedToken() bool {
	return receiver.authType == AuthTypeScoped
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
