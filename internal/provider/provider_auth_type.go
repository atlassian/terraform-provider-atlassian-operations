package provider

import (
	"fmt"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
)

// resolveAuthType picks the auth_type from the environment variable first (like every setting of this provider),
// then the configuration value, then the default: "scoped" for a service account address, because a service account
// never has a classic API token, and "classic" otherwise.
//
// Both sources are compared case-insensitively with surrounding whitespace removed: the environment variable
// because the CI path and secret stores routinely append a newline, and the configuration value for the same
// reason a user might write ATLASSIAN_OPS_AUTH_TYPE=Scoped and expect auth_type = "Scoped" to work the same way.
// A value that is empty after trimming counts as unset. An invalid value is reported with the original text, so a
// stray character stays visible in the error.
func resolveAuthType(envValue, configValue, emailAddress string) (string, error) {
	raw, source := envValue, "ATLASSIAN_OPS_AUTH_TYPE"
	authType := strings.ToLower(strings.TrimSpace(envValue))
	if authType == "" {
		raw, source = configValue, "auth_type"
		authType = strings.ToLower(strings.TrimSpace(configValue))
	}
	if authType == "" {
		if strings.HasSuffix(strings.ToLower(strings.TrimSpace(emailAddress)), dto.ServiceAccountEmailSuffix) {
			return dto.AuthTypeScoped, nil
		}
		return dto.AuthTypeClassic, nil
	}
	if authType != dto.AuthTypeClassic && authType != dto.AuthTypeScoped {
		return "", fmt.Errorf("auth_type (or ATLASSIAN_OPS_AUTH_TYPE) must be \"classic\" or \"scoped\", got %q from %s", raw, source)
	}
	return authType, nil
}
