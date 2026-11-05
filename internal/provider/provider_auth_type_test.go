package provider

import (
	"strings"
	"testing"
)

func TestResolveAuthType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, env, config, email string
		want                     string
		wantErr                  []string
	}{
		{"configured value", "", "scoped", "me@example.com", "scoped", nil},
		{"environment variable wins over the configuration", "classic", "scoped", "me@example.com", "classic", nil},
		{"trailing newline from a secret store is ignored", "scoped\n", "", "me@example.com", "scoped", nil},
		{"surrounding spaces are ignored", "  classic  ", "", "me@example.com", "classic", nil},
		{"the environment variable is case-insensitive", "Scoped", "", "me@example.com", "scoped", nil},
		{"the configuration value is case-insensitive too", "", "Classic", "me@example.com", "classic", nil},
		{"surrounding spaces and upper case in the configuration are both ignored", "", "  SCOPED  ", "me@example.com", "scoped", nil},
		{"a whitespace-only variable counts as unset", "   \n", "scoped", "me@example.com", "scoped", nil},
		{"default for a user is classic", "", "", "me@example.com", "classic", nil},
		{"default for a service account is scoped", "", "", "bot@serviceaccount.atlassian.com", "scoped", nil},
		{"the service account suffix is matched case-insensitively", "", "", "Bot@ServiceAccount.Atlassian.com ", "scoped", nil},
		{"default without an email is classic", "", "", "", "classic", nil},
		{"invalid variable names the variable and quotes the raw value", "oauth\n", "", "me@example.com", "", []string{`must be "classic" or "scoped"`, `"oauth\n"`, "ATLASSIAN_OPS_AUTH_TYPE"}},
		{"invalid configuration value names the attribute", "", "token", "me@example.com", "", []string{`got "token" from auth_type`}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveAuthType(tc.env, tc.config, tc.email)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				for _, part := range tc.wantErr {
					if !strings.Contains(err.Error(), part) {
						t.Errorf("error %q does not contain %q", err.Error(), part)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
