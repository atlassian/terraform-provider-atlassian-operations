package aggClient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto/agg"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
)

// gateway is a fake Teams GraphQL gateway. It answers every request with the body registered for its operation, or
// with status when that is set, and records the requests.
type gateway struct {
	bodies    map[string]string
	status    int  // answer every request with this HTTP status
	first     int  // answer the first request with this HTTP status, like a flaky gateway
	dropFirst bool // close the connection of the first request without answering it
	url       string
	// Requests are answered one at a time and calls is read only after Send returned, so no lock.
	calls []agg.Request
}

func newGateway(t *testing.T, bodies map[string]string) *gateway {
	t.Helper()
	g := &gateway{bodies: bodies}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req agg.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("gateway: cannot decode request: %v", err)
		}
		g.calls = append(g.calls, req)
		if g.dropFirst && len(g.calls) == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("gateway: cannot drop the connection: %v", err)
				return
			}
			_ = conn.Close()
			return
		}
		status := g.status
		if g.first != 0 && len(g.calls) == 1 {
			status = g.first
		}
		if status != 0 {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"errors":[{"message":"gateway error"}]}`))
			return
		}
		body, ok := g.bodies[req.OperationName]
		if !ok {
			t.Errorf("gateway: unexpected %s request", req.OperationName)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	g.url = server.URL
	t.Cleanup(server.Close)
	return g
}

// client returns a client that sends its requests to the gateway with the model's retry settings.
func (g *gateway) client(model dto.AtlassianOpsProviderModel) Client {
	c := New(model)
	c.newRequest = func(model dto.AtlassianOpsProviderModel) *httpClient.Request {
		req := httpClient.NewRequest()
		req.SetUrl(g.url)
		req.SetRetryCount(model.GetApiRetryCount())
		req.SetRetryWaitTime(model.GetApiRetryWait())
		req.SetRetryMaxWaitTime(model.GetApiRetryWaitMax())
		return req
	}
	return c
}

func (g *gateway) operations() string {
	names := make([]string, len(g.calls))
	for i, c := range g.calls {
		names[i] = c.OperationName
	}
	return strings.Join(names, ",")
}

func testModel() dto.AtlassianOpsProviderModel {
	return dto.NewAtlassianOpsProviderModel("jira-service-desk", "cloud-id", "mysite.atlassian.net", "bot@serviceaccount.atlassian.com",
		"token", "", "", dto.AuthTypeScoped, 0, 0, 0, false)
}

const (
	activeTeam = `{"data":{"team":{"teamV2":{"id":"ari:cloud:identity::team/team-1","displayName":"Ops","state":"ACTIVE","membershipSettings":"OPEN","members":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}}`
	goneTeam   = `{"data":{"team":{"teamV2":null}},"errors":[{"message":"gone","path":["team","teamV2"],"extensions":{"classification":"Gone"}}]}`
)

func TestARIsAreNeverPrefixedTwice(t *testing.T) {
	for _, c := range [][2]string{
		{teamIdFromARI(teamARIPrefix + "t"), "t"},
		{teamIdFromARI("t"), "t"},
		{teamARI("t"), teamARIPrefix + "t"},
		{teamARI(teamARIPrefix + "t"), teamARIPrefix + "t"},
		{userARI("u"), userARIPrefix + "u"},
		{userARI(userARIPrefix + "u"), userARIPrefix + "u"},
	} {
		if c[0] != c[1] {
			t.Errorf("got %q, want %q", c[0], c[1])
		}
	}
}

func TestUserPermissions(t *testing.T) {
	if toUserPermissions(nil) != nil {
		t.Fatal("no permissions from AGG must stay nil, not all-false")
	}
	view := toUserPermissions([]agg.Permission{agg.PermissionViewTeam})
	if view.UpdateTeam || view.AddMembers || view.RemoveMembers || view.DeleteTeam {
		t.Errorf("view only: %+v", view)
	}
	member := toUserPermissions([]agg.Permission{agg.PermissionViewTeam, agg.PermissionManageTeam, agg.PermissionDeleteTeam})
	if !member.UpdateTeam || !member.AddMembers || !member.RemoveMembers || !member.DeleteTeam {
		t.Errorf("member of the team: %+v", member)
	}
}

func TestCreateTeam(t *testing.T) {
	ctx := context.Background()
	g := newGateway(t, map[string]string{"CreateTeam": `{"data":{"team":{"createTeam":{"team":{"id":"ari:cloud:identity::team/team-1","displayName":"Ops","description":"d","membershipSettings":"MEMBER_INVITE","state":"ACTIVE","permissions":["CAN_VIEW_TEAM","CAN_MANAGE_TEAM","CAN_DELETE_TEAM"]},"success":true,"errors":[]}}}}`})

	team, err := g.client(testModel()).CreateTeam(ctx, dto.TeamDto{DisplayName: "Ops", Description: "d", OrganizationId: "org-1", TeamType: dto.MEMBER_INVITE}, "", []string{"acc-1", userARIPrefix + "acc-2"})
	if err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}
	if team.TeamId != "team-1" || team.OrganizationId != "org-1" || team.TeamType != dto.MEMBER_INVITE {
		t.Errorf("unexpected team: %+v", team)
	}
	if team.UserPermissions == nil || !team.UserPermissions.UpdateTeam || !team.UserPermissions.DeleteTeam {
		t.Errorf("permissions from the mutation payload expected, got %+v", team.UserPermissions)
	}
	if g.operations() != "CreateTeam" {
		t.Fatalf("expected a single CreateTeam call, got %s", g.operations())
	}
	input := g.calls[0].Variables["input"].(map[string]any)
	members, _ := input["members"].([]any)
	if input["scopeId"] != siteARIPrefix+"cloud-id" || len(members) != 2 || members[0] != userARIPrefix+"acc-1" || members[1] != userARIPrefix+"acc-2" {
		t.Errorf("scopeId must be the cloud_id site ARI and members user ARIs, prefixed once: %v", input)
	}
}

func TestCreateTeamErrors(t *testing.T) {
	ctx := context.Background()
	newTeam := dto.TeamDto{DisplayName: "Ops", TeamType: dto.OPEN}

	t.Run("an HTTP error", func(t *testing.T) {
		g := newGateway(t, nil)
		g.status = http.StatusBadGateway
		_, err := g.client(testModel()).CreateTeam(ctx, newTeam, "", []string{"acc-1"})
		if err == nil || !strings.Contains(err.Error(), "returned status 502") || !strings.Contains(err.Error(), `a team named "Ops" may have been created and left outside Terraform state; check the Teams UI before retrying`) {
			t.Errorf("expected the API's error with the hint that the team may exist, got %v", err)
		}
	})

	t.Run("a 4xx answer is a rejection", func(t *testing.T) {
		g := newGateway(t, nil)
		g.status = http.StatusForbidden
		_, err := g.client(testModel()).CreateTeam(ctx, newTeam, "", []string{"acc-1"})
		if err == nil || !strings.Contains(err.Error(), "returned status 403") || !strings.Contains(err.Error(), "rejected") || strings.Contains(err.Error(), "may have been created") {
			t.Errorf("expected the plain rejection, got %v", err)
		}
	})

	t.Run("top-level errors only", func(t *testing.T) {
		g := newGateway(t, map[string]string{"CreateTeam": `{"data":{"team":{"createTeam":null}},"errors":[{"message":"scope"}]}`})
		if _, err := g.client(testModel()).CreateTeam(ctx, newTeam, "", []string{"acc-1"}); err == nil || !strings.Contains(err.Error(), "CreateTeam errors: scope") {
			t.Errorf("expected the API's error, got %v", err)
		}
	})

	t.Run("payload errors without a team", func(t *testing.T) {
		g := newGateway(t, map[string]string{"CreateTeam": `{"data":{"team":{"createTeam":{"team":null,"success":false,"errors":[{"message":"member rejected"}]}}}}`})
		if _, err := g.client(testModel()).CreateTeam(ctx, newTeam, "", []string{"acc-1"}); err == nil || !strings.Contains(err.Error(), "CreateTeam errors: member rejected") {
			t.Errorf("expected the API's error, got %v", err)
		}
		if g.operations() != "CreateTeam" {
			t.Errorf("no team came back, so nothing is deleted, got %s", g.operations())
		}
	})

	t.Run("a partially created team is deleted again", func(t *testing.T) {
		g := newGateway(t, map[string]string{
			"CreateTeam": `{"data":{"team":{"createTeam":{"team":{"id":"ari:cloud:identity::team/team-1"},"success":false,"errors":[{"message":"member rejected"}]}}}}`,
			"DeleteTeam": `{"data":{"team":{"deleteTeam":{"success":true,"errors":[]}}}}`,
		})
		if _, err := g.client(testModel()).CreateTeam(ctx, newTeam, "", []string{"acc-1"}); err == nil || !strings.Contains(err.Error(), "partially created team was deleted") {
			t.Errorf("expected the rollback note, got %v", err)
		}
		if g.operations() != "CreateTeam,DeleteTeam" {
			t.Errorf("expected CreateTeam then DeleteTeam, got %s", g.operations())
		}
	})
}

func TestGetTeam(t *testing.T) {
	ctx := context.Background()
	g := newGateway(t, map[string]string{"TeamV2": activeTeam})
	c := g.client(testModel())

	withSite, err := c.GetTeam(ctx, teamARIPrefix+"team-1", "site-1")
	if err != nil || withSite.TeamId != "team-1" || withSite.SiteId == nil || *withSite.SiteId != "site-1" {
		t.Errorf("a configured site_id must be echoed back: %+v, %v", withSite, err)
	}
	withoutSite, err := c.GetTeam(ctx, "team-1", "")
	if err != nil || withoutSite.SiteId != nil {
		t.Errorf("an unconfigured site_id must stay nil, not the cloud_id fallback: %+v, %v", withoutSite, err)
	}
	if g.calls[0].Variables["id"] != "team-1" || g.calls[0].Variables["siteId"] != "site-1" || g.calls[1].Variables["siteId"] != "cloud-id" {
		t.Errorf("the query must use the raw team ID and the resolved site: %v, %v", g.calls[0].Variables, g.calls[1].Variables)
	}
}

func TestGetTeamNotFound(t *testing.T) {
	ctx := context.Background()
	bodies := map[string]string{
		"deleted":   goneTeam,
		"unknown":   `{"data":{"team":{"teamV2":null}},"errors":[{"message":"missing","path":["team","teamV2"],"extensions":{"classification":"NotFound"}}]}`,
		"disbanded": `{"data":{"team":{"teamV2":{"id":"x","state":"DISBANDED"}}}}`,
	}
	for name, body := range bodies {
		g := newGateway(t, map[string]string{"TeamV2": body})
		if _, err := g.client(testModel()).GetTeam(ctx, "team-1", ""); !errors.Is(err, ErrTeamNotFound) {
			t.Errorf("%s team: expected ErrTeamNotFound, got %v", name, err)
		}
	}

	g := newGateway(t, nil)
	g.status = http.StatusUnauthorized
	if _, err := g.client(testModel()).GetTeam(ctx, "team-1", ""); err == nil || errors.Is(err, ErrTeamNotFound) {
		t.Errorf("a 401 must be an error, not a missing team: %v", err)
	}
}

func TestListTeamMembers(t *testing.T) {
	ctx := context.Background()
	g := newGateway(t, map[string]string{"TeamMembers": `{"data":{"team":{"teamV2":{"state":"ACTIVE","members":{"nodes":[{"member":{"accountId":"ari:cloud:identity::user/a1"}},{"member":{"accountId":"a2"}}],"pageInfo":{"hasNextPage":true,"endCursor":"cursor-1"}}}}}}`})

	page, err := g.client(testModel()).ListTeamMembers(ctx, "team-1", "site-1", dto.TeamMemberListRequest{First: 50, After: "cursor-0"})
	if err != nil {
		t.Fatalf("ListTeamMembers: %v", err)
	}
	if len(page.Results) != 2 || page.Results[0].AccountId != "a1" || page.Results[1].AccountId != "a2" || !page.PageInfo.HasNextPage || page.PageInfo.EndCursor != "cursor-1" {
		t.Errorf("unexpected page: %+v", page)
	}
	if g.calls[0].Variables["after"] != "cursor-0" {
		t.Errorf("the cursor must be sent: %v", g.calls[0].Variables)
	}

	g = newGateway(t, map[string]string{"TeamMembers": `{"data":{"team":{"teamV2":{"state":"ACTIVE","members":null}}}}`})
	if _, err = g.client(testModel()).ListTeamMembers(ctx, "team-1", "site-1", dto.TeamMemberListRequest{First: 50}); err == nil {
		t.Error("a null members field must be an error, not an empty last page")
	}
}

func TestUpdateTeam(t *testing.T) {
	ctx := context.Background()
	g := newGateway(t, map[string]string{"UpdateTeam": `{"data":{"team":{"updateTeam":{"team":{"id":"ari:cloud:identity::team/team-1","displayName":"Renamed","description":"d","membershipSettings":"OPEN","state":"ACTIVE","permissions":["CAN_VIEW_TEAM"]},"success":true,"errors":[]}}}}`})

	site := "site-1"
	out, err := g.client(testModel()).UpdateTeam(ctx, dto.TeamDto{TeamId: "team-1", DisplayName: "Renamed", Description: "d", OrganizationId: "org-1", SiteId: &site, TeamType: dto.OPEN}, site)
	if err != nil {
		t.Fatalf("UpdateTeam: %v", err)
	}
	if out.TeamId != "team-1" || out.DisplayName != "Renamed" || out.OrganizationId != "org-1" || *out.SiteId != "site-1" {
		t.Errorf("unexpected team: %+v", out)
	}
	if out.UserPermissions == nil || out.UserPermissions.UpdateTeam {
		t.Errorf("view-only permissions from the mutation payload expected, got %+v", out.UserPermissions)
	}
	if input := g.calls[0].Variables["input"].(map[string]any); input["teamId"] != teamARIPrefix+"team-1" || input["membershipSettings"] != nil {
		t.Errorf("the input must carry the team ARI and no team type: %v", input)
	}
}

func TestMutationsWhenTheTeamIsGone(t *testing.T) {
	ctx := context.Background()
	for _, extensions := range []string{`{"statusCode":404,"errorType":"NotFound"}`, `{"statusCode":410,"errorType":"Gone"}`} {
		payload := `{"success":false,"errors":[{"message":"Team not found","extensions":` + extensions + `}]}`
		g := newGateway(t, map[string]string{
			"UpdateTeam":    `{"data":{"team":{"updateTeam":` + payload + `}}}`,
			"DeleteTeam":    `{"data":{"team":{"deleteTeam":` + payload + `}}}`,
			"AddMembers":    `{"data":{"team":{"addMembers":` + payload + `}}}`,
			"RemoveMembers": `{"data":{"team":{"removeMembers":` + payload + `}}}`,
		})
		c := g.client(testModel())
		if _, err := c.UpdateTeam(ctx, dto.TeamDto{TeamId: "team-1"}, ""); !errors.Is(err, ErrTeamNotFound) {
			t.Errorf("UpdateTeam with %s: expected ErrTeamNotFound, got %v", extensions, err)
		}
		if err := c.AddMembers(ctx, "team-1", "", []string{"a1"}); !errors.Is(err, ErrTeamNotFound) {
			t.Errorf("AddMembers with %s: expected ErrTeamNotFound, got %v", extensions, err)
		}
		if err := c.RemoveMembers(ctx, "team-1", "", []string{"a1"}); !errors.Is(err, ErrTeamNotFound) {
			t.Errorf("RemoveMembers with %s: expected ErrTeamNotFound, got %v", extensions, err)
		}
		if err := c.DeleteTeam(ctx, "team-1", ""); err != nil {
			t.Errorf("DeleteTeam with %s: a team that is gone counts as deleted, got %v", extensions, err)
		}
	}

	payload := `{"success":false,"errors":[{"message":"not allowed","extensions":{"statusCode":403,"errorType":"Forbidden"}}]}`
	g := newGateway(t, map[string]string{
		"UpdateTeam": `{"data":{"team":{"updateTeam":` + payload + `}}}`,
		"DeleteTeam": `{"data":{"team":{"deleteTeam":` + payload + `}}}`,
		"AddMembers": `{"data":{"team":{"addMembers":` + payload + `}}}`,
	})
	c := g.client(testModel())
	if _, err := c.UpdateTeam(ctx, dto.TeamDto{TeamId: "team-1"}, ""); err == nil || errors.Is(err, ErrTeamNotFound) {
		t.Errorf("UpdateTeam: any other failure is an error, got %v", err)
	}
	if err := c.AddMembers(ctx, "team-1", "", []string{"a1"}); err == nil || errors.Is(err, ErrTeamNotFound) {
		t.Errorf("AddMembers: any other failure is an error, got %v", err)
	}
	if err := c.DeleteTeam(ctx, "team-1", ""); err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("DeleteTeam: any other failure is an error, got %v", err)
	}
}

func TestAddAndRemoveMembers(t *testing.T) {
	ctx := context.Background()
	g := newGateway(t, map[string]string{
		"AddMembers":    `{"data":{"team":{"addMembers":{"success":false,"errors":[],"failedMembers":[{"accountId":"a2","code":"ADD_MEMBER_FAILED","message":"[MemberNotFound]"}]}}}}`,
		"RemoveMembers": `{"data":{"team":{"removeMembers":{"success":true,"errors":[]}}}}`,
	})
	c := g.client(testModel())

	if err := c.RemoveMembers(ctx, "team-1", "site-1", []string{userARIPrefix + "a1"}); err != nil {
		t.Fatalf("RemoveMembers: %v", err)
	}
	if err := c.AddMembers(ctx, "team-1", "site-1", []string{"a2"}); err == nil || !strings.Contains(err.Error(), "member a2: ADD_MEMBER_FAILED ([MemberNotFound])") {
		t.Errorf("a rejected member must be reported with its account ID and reason, got %v", err)
	}
	for _, call := range g.calls {
		input := call.Variables["input"].(map[string]any)
		if input["teamId"] != teamARIPrefix+"team-1" || input["scopeId"] != siteARIPrefix+"site-1" {
			t.Errorf("%s: teamId must be the team ARI and scopeId the site ARI, got %v", call.OperationName, input)
		}
		if ids := input["memberIds"].([]any); strings.HasPrefix(ids[0].(string), userARIPrefix) {
			t.Errorf("%s: memberIds must be raw account IDs, got %v", call.OperationName, ids)
		}
	}
}

func retryingModel() dto.AtlassianOpsProviderModel {
	return dto.NewAtlassianOpsProviderModel("jira-service-desk", "cloud-id", "example.atlassian.net", "me@example.com",
		"token", "", "", dto.AuthTypeScoped, 2, time.Millisecond, 5*time.Millisecond, false)
}

// A 429 means the gateway did not run the mutation, so it is safe to send it again.
func TestMutationsAreRetriedOnTooManyRequests(t *testing.T) {
	g := newGateway(t, map[string]string{"AddMembers": `{"data":{"team":{"addMembers":{"success":true,"errors":[]}}}}`})
	g.first = http.StatusTooManyRequests

	if err := g.client(retryingModel()).AddMembers(context.Background(), "team-1", "site-1", []string{"a1"}); err != nil || len(g.calls) != 2 {
		t.Errorf("a mutation must be retried after a 429, got %d calls and error %v", len(g.calls), err)
	}
}

// A dropped connection leaves it unknown whether the gateway ran the mutation, so it is not sent again.
func TestMutationsAreNotRetriedAfterADroppedConnection(t *testing.T) {
	g := newGateway(t, map[string]string{"AddMembers": `{"data":{"team":{"addMembers":{"success":true,"errors":[]}}}}`})
	g.dropFirst = true

	if err := g.client(retryingModel()).AddMembers(context.Background(), "team-1", "site-1", []string{"a1"}); err == nil || len(g.calls) != 1 {
		t.Errorf("a mutation must be sent once and fail after a dropped connection, got %d calls and error %v", len(g.calls), err)
	}
}

func TestReadsAreRetriedAndMutationsAreNot(t *testing.T) {
	ctx := context.Background()
	retrying := retryingModel()
	calls := map[string]func(c Client) error{
		"TeamV2": func(c Client) error { _, err := c.GetTeam(ctx, "team-1", "site-1"); return err },
		"TeamMembers": func(c Client) error {
			_, err := c.ListTeamMembers(ctx, "team-1", "site-1", dto.TeamMemberListRequest{First: 50})
			return err
		},
		"CreateTeam": func(c Client) error {
			_, err := c.CreateTeam(ctx, dto.TeamDto{DisplayName: "Ops"}, "", []string{"a1"})
			return err
		},
		"UpdateTeam":    func(c Client) error { _, err := c.UpdateTeam(ctx, dto.TeamDto{TeamId: "team-1"}, ""); return err },
		"DeleteTeam":    func(c Client) error { return c.DeleteTeam(ctx, "team-1", "site-1") },
		"AddMembers":    func(c Client) error { return c.AddMembers(ctx, "team-1", "site-1", []string{"a1"}) },
		"RemoveMembers": func(c Client) error { return c.RemoveMembers(ctx, "team-1", "site-1", []string{"a1"}) },
	}
	operations := make([]string, 0, len(calls))
	for operation := range calls {
		operations = append(operations, operation)
	}
	slices.Sort(operations)

	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			g := newGateway(t, map[string]string{"TeamV2": activeTeam, "TeamMembers": activeTeam})
			g.first = http.StatusInternalServerError

			err := calls[operation](g.client(retrying))
			isRead := operation == "TeamV2" || operation == "TeamMembers"
			if isRead && (err != nil || len(g.calls) != 2) {
				t.Errorf("a read must be retried after a 500, got %d calls and error %v", len(g.calls), err)
			}
			if !isRead && (err == nil || len(g.calls) != 1) {
				t.Errorf("a write must be sent once and fail after a 500, got %d calls and error %v", len(g.calls), err)
			}
		})
	}
}
