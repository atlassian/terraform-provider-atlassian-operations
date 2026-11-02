package provider

import (
	"net/http"
	"strings"
	"testing"
)

// With auth_type = "scoped" the user lookup goes to api.atlassian.com, which answers a wrong or expired token
// with 401 instead of Jira's "200 + X-Seraph-LoginReason". The diagnostic must then tell the user what to check,
// and the scoped and classic paths get different guidance: the scoped path already has auth_type = "scoped", so
// pointing at that setting again would be useless, while the classic path's token may actually be a scoped one.
func TestUserLookupAuthHint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		scoped         bool
		status         int
		body           string
		wantContains   []string
		wantNotContain []string
		wantEmptyHint  bool
	}{
		{
			name:           "401 on the scoped path points at the token, not at auth_type = scoped again",
			scoped:         true,
			status:         http.StatusUnauthorized,
			body:           `{"code":401,"message":"Unauthorized"}`,
			wantContains:   []string{"Check `email_address` and `token`", "classic API token", `auth_type = "classic"`},
			wantNotContain: []string{`auth_type = "scoped"`},
		},
		{
			name:         "401 on the classic path points at auth_type = scoped",
			scoped:       false,
			status:       http.StatusUnauthorized,
			body:         `{"code":401,"message":"Unauthorized"}`,
			wantContains: []string{"Check `email_address` and `token`", `auth_type = "scoped"`},
		},
		{
			name:         "401 scope mismatch on the classic path points at the token scopes",
			scoped:       false,
			status:       http.StatusUnauthorized,
			body:         `{"code":401,"message":"Unauthorized; scope does not match"}`,
			wantContains: []string{"rejected because of its scopes", "add the missing scopes"},
		},
		{
			name:           "401 scope mismatch on the scoped path blames the API, not the token's auth_type",
			scoped:         true,
			status:         http.StatusUnauthorized,
			body:           `{"code":401,"message":"Unauthorized; scope does not match"}`,
			wantContains:   []string{"rejected because of its scopes", "Compass Operations API accepts classic API tokens only"},
			wantNotContain: []string{"add the missing scopes", `auth_type = "scoped"`},
		},
		{
			name:         "401 without a body still gets the classic path's credentials hint",
			scoped:       false,
			status:       http.StatusUnauthorized,
			body:         "",
			wantContains: []string{"Check `email_address` and `token`"},
		},
		{
			name:          "other errors are reported as they are",
			scoped:        true,
			status:        http.StatusNotFound,
			body:          `{"message":"not found"}`,
			wantEmptyHint: true,
		},
		{
			name:          "403 is a permission problem, not a credentials one",
			scoped:        false,
			status:        http.StatusForbidden,
			body:          `{"message":"forbidden"}`,
			wantEmptyHint: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := userLookupAuthHint(tc.scoped, tc.status, tc.body)
			if tc.wantEmptyHint {
				if got != "" {
					t.Fatalf("expected no hint, got %q", got)
				}
				return
			}
			if got == "" {
				t.Fatalf("expected a hint for status %d", tc.status)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("hint %q does not contain %q", got, want)
				}
			}
			for _, notWant := range tc.wantNotContain {
				if strings.Contains(got, notWant) {
					t.Errorf("hint %q must not contain %q", got, notWant)
				}
			}
		})
	}
}

// The org-admin directory search always authenticates with org_admin_token (a bearer token); auth_type and
// email_address/token have no effect on it, so a 401 there must not reuse the Jira user lookup's hint.
func TestOrgAdminAuthHint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		status         int
		body           string
		wantContains   []string
		wantNotContain []string
		wantEmptyHint  bool
	}{
		{
			name:           "401 asks to check org_admin_token, not email_address/token or auth_type",
			status:         http.StatusUnauthorized,
			body:           `{"message":"Unauthorized"}`,
			wantContains:   []string{"org_admin_token"},
			wantNotContain: []string{"email_address", "auth_type"},
		},
		{
			name:          "other errors are reported as they are",
			status:        http.StatusNotFound,
			body:          `{"message":"not found"}`,
			wantEmptyHint: true,
		},
		{
			name:          "403 is a permission problem, not a credentials one",
			status:        http.StatusForbidden,
			body:          `{"message":"forbidden"}`,
			wantEmptyHint: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := orgAdminAuthHint(tc.status, tc.body)
			if tc.wantEmptyHint {
				if got != "" {
					t.Fatalf("expected no hint, got %q", got)
				}
				return
			}
			if got == "" {
				t.Fatalf("expected a hint for status %d", tc.status)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("hint %q does not contain %q", got, want)
				}
			}
			for _, notWant := range tc.wantNotContain {
				if strings.Contains(got, notWant) {
					t.Errorf("hint %q must not contain %q: the org-admin path uses a different credential", got, notWant)
				}
			}
		})
	}
}

// The org-admin directory search (product_type = "compass") cannot look up a service account by email at all; an
// empty result for one must say so instead of the message for other accounts.
func TestEmptyOrgAdminSearchError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		emailAddress string
		wantContains string
	}{
		{"a service account address", "my-bot@serviceaccount.atlassian.com", "service accounts cannot be looked up by email on Compass"},
		{"matched case-insensitively with surrounding whitespace", "  My-Bot@ServiceAccount.Atlassian.Com  ", "service accounts cannot be looked up by email on Compass"},
		{"a human address gets the directory message", "person@example.com", "no account with that email address was found in the organization's directory"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := emptyOrgAdminSearchError(tc.emailAddress)
			if !strings.Contains(got, tc.wantContains) {
				t.Errorf("error %q does not contain %q", got, tc.wantContains)
			}
		})
	}
}
