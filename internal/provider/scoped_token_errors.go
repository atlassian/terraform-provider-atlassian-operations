package provider

import (
	"net/http"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
)

const (
	// seraphLoginReasonHeader is set by Jira when it could not authenticate the request. Jira then treats the request
	// as anonymous, so the user search answers 200 with an empty list instead of 401.
	seraphLoginReasonHeader    = "X-Seraph-LoginReason"
	seraphAuthenticationFailed = "AUTHENTICATED_FAILED"

	scopedTokenGuideHint = "If `token` is a scoped API token or a service account token, set `auth_type = \"scoped\"` in the provider block " +
		"(or ATLASSIAN_OPS_AUTH_TYPE=scoped); see the \"Scoped API Tokens and Service Accounts\" guide."
)

// scopedTokenHint returns guidance to append to an error message when the API rejected the token because of its
// scopes. On the classic path this means a scoped API token was sent to an API that rejects scoped tokens (the
// Teams REST API answers like that while auth_type is still classic), so the hint points at auth_type = "scoped".
// On the scoped path auth_type is already "scoped", so that hint would be wrong: some APIs behind these calls
// (JSM Operations' enable-ops, reached the same way on both paths) reject scoped tokens outright regardless of
// auth_type, for example the Compass Operations API, which accepts classic API tokens only.
func scopedTokenHint(scoped bool, statusCode int, errorResponse string) string {
	if statusCode != http.StatusUnauthorized || !strings.Contains(strings.ToLower(errorResponse), "scope does not match") {
		return ""
	}
	if scoped {
		return " The token was rejected because of its scopes: the API behind this call rejects this token's scopes (for example the Compass Operations API accepts classic API tokens only)."
	}
	return " The token was rejected because of its scopes. " + scopedTokenGuideHint + " Otherwise add the missing scopes to the token."
}

// authenticationFailed reports whether Jira answered the request as unauthenticated.
func authenticationFailed(resp *httpClient.Response) bool {
	return resp != nil && strings.EqualFold(resp.GetHeader(seraphLoginReasonHeader), seraphAuthenticationFailed)
}

// orgAdminAuthHint returns guidance to append when the org-admin directory search was rejected because of the
// credentials. Unlike the Jira user lookup, that call always authenticates with org_admin_token (a bearer token),
// never with email_address/token, and auth_type has no effect on it, so it needs its own hint rather than
// userLookupAuthHint's.
func orgAdminAuthHint(statusCode int, _ string) string {
	if statusCode == http.StatusUnauthorized {
		return " Check `org_admin_token`."
	}
	return ""
}
