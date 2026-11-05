package httpClientHelpers

import (
	"encoding/base64"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
)

func newModel(scoped, isStaging bool) dto.AtlassianOpsProviderModel {
	return newModelWithTeamsToken(scoped, isStaging, "")
}

func newModelWithTeamsToken(scoped, isStaging bool, teamsToken string) dto.AtlassianOpsProviderModel {
	authType := dto.AuthTypeClassic
	if scoped {
		authType = dto.AuthTypeScoped
	}
	return dto.NewAtlassianOpsProviderModel(
		"jira-service-desk",
		"cloud-id",
		"mysite.atlassian.net",
		"me@example.com",
		"token",
		teamsToken,
		"",
		authType,
		0, 0, 0,
		isStaging,
	)
}

func basicAuth(email, token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(email+":"+token))
}

func TestTeamsRequestsUseTeamsTokenWhenSet(t *testing.T) {
	withTeams := newModelWithTeamsToken(true, false, "teams-token")
	without := newModelWithTeamsToken(true, false, "")

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"AGG uses teams_token", GenerateAGGClientRequest(withTeams).GetInnerRequest().Header.Get("Authorization"), basicAuth("me@example.com", "teams-token")},
		{"Teams REST keeps token even when teams_token is set", GenerateTeamsClientRequest(withTeams).GetInnerRequest().Header.Get("Authorization"), basicAuth("me@example.com", "token")},
		{"JSM Ops keeps token", GenerateJsmOpsClientRequest(withTeams).GetInnerRequest().Header.Get("Authorization"), basicAuth("me@example.com", "token")},
		{"user lookup keeps token", GenerateUserClientRequest(withTeams).GetInnerRequest().Header.Get("Authorization"), basicAuth("me@example.com", "token")},
		{"AGG falls back to token", GenerateAGGClientRequest(without).GetInnerRequest().Header.Get("Authorization"), basicAuth("me@example.com", "token")},
		{"Teams REST uses token without teams_token", GenerateTeamsClientRequest(without).GetInnerRequest().Header.Get("Authorization"), basicAuth("me@example.com", "token")},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: Authorization = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestGenerateAGGClientRequest_UsesAtlassianApiDomain(t *testing.T) {
	tests := []struct {
		name      string
		isStaging bool
		want      string
	}{
		{name: "production", isStaging: false, want: "https://api.atlassian.com/gateway/api/graphql"},
		{name: "staging", isStaging: true, want: "https://api.stg.atlassian.com/gateway/api/graphql"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateAGGClientRequest(newModel(true, tt.isStaging)).GetInnerRequest().URL.String()
			if got != tt.want {
				t.Errorf("AGG URL = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateUserClientRequest_ScopedTokenUsesGateway(t *testing.T) {
	tests := []struct {
		name      string
		scoped    bool
		isStaging bool
		want      string
	}{
		{name: "classic token stays on the site", scoped: false, want: "https://mysite.atlassian.net/rest/api/3/user/"},
		{name: "scoped token goes through ex/jira", scoped: true, want: "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/user/"},
		{name: "scoped token on staging", scoped: true, isStaging: true, want: "https://api.stg.atlassian.com/ex/jira/cloud-id/rest/api/3/user/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUserClientRequest(newModel(tt.scoped, tt.isStaging)).GetInnerRequest().URL.String()
			if got != tt.want {
				t.Errorf("user URL = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateTeamsClientRequest_StaysOnSiteDomain(t *testing.T) {
	got := GenerateTeamsClientRequest(newModel(false, true)).GetInnerRequest().URL.String()
	want := "https://mysite.atlassian.net/gateway/api/public/teams/v1/org/"
	if got != want {
		t.Errorf("Teams REST URL = %q, want %q", got, want)
	}
}
