package provider

import (
	"strings"
	"testing"
)

func TestScopedTokenHint(t *testing.T) {
	const mismatch = "Unauthorized; scope does not match"
	for _, tc := range []struct {
		name           string
		scoped         bool
		status         int
		body           string
		contains, omit string // contains == "" means no hint at all
	}{
		{"classic path suggests auth_type", false, 401, mismatch, `auth_type = "scoped"`, ""},
		{"classic path ignores the case", false, 401, "UNAUTHORIZED; SCOPE DOES NOT MATCH", `auth_type = "scoped"`, ""},
		{"classic path, another status", false, 403, mismatch, "", ""},
		{"classic path, another reason", false, 401, "token expired", "", ""},
		{"scoped path does not repeat auth_type", true, 401, mismatch, "Compass Operations API accepts classic API tokens only", `auth_type = "scoped"`},
		{"scoped path, another status", true, 403, mismatch, "", ""},
	} {
		hint := scopedTokenHint(tc.scoped, tc.status, tc.body)
		if tc.contains == "" && hint != "" {
			t.Errorf("%s: no hint expected, got %q", tc.name, hint)
		}
		if tc.contains != "" && !strings.Contains(hint, tc.contains) {
			t.Errorf("%s: the hint must contain %q, got %q", tc.name, tc.contains, hint)
		}
		if tc.omit != "" && strings.Contains(hint, tc.omit) {
			t.Errorf("%s: the hint must not contain %q, got %q", tc.name, tc.omit, hint)
		}
	}
}
