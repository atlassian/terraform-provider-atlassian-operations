package provider

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
)

// policyType is a constraint type for alert and notification policies
type policyType string

const (
	policyTypeAlert        policyType = "alert"
	policyTypeNotification policyType = "notification"
)

type policyPosition struct {
	ID    string
	Order int64
}

type policyOrderingClient interface {
	List(ctx context.Context, teamID string) ([]policyPosition, error)
	Move(ctx context.Context, teamID, policyID string, index int64) error
}

type policyOrderingAPI struct {
	policyType     policyType
	requestFactory func() *httpClient.Request
}

func newPolicyOrderingClient(configuration dto.AtlassianOpsProviderModel, pt policyType) policyOrderingClient {
	return &policyOrderingAPI{
		policyType: pt,
		requestFactory: func() *httpClient.Request {
			return httpClientHelpers.GenerateJsmOpsClientRequest(configuration)
		},
	}
}

// Scope-bound calls must not follow redirects, which can change their path,
// policy type, or even rewrite a POST to GET before we validate the response.
func (c *policyOrderingAPI) newRequest(ctx context.Context) *httpClient.Request {
	req := c.requestFactory().SetContext(ctx)
	client := *req.GetInnerClient().HTTPClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req.GetInnerClient().HTTPClient = &client
	return req
}

// IDs are path segments. Reject separators rather than allowing JoinPath to
// normalize a user-supplied ID into a different API endpoint.
func validatePolicyPathID(id string) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(id) != id || id == "." || id == ".." || strings.ContainsAny(id, "/\\?#%") {
		return fmt.Errorf("invalid policy or team ID %q", id)
	}
	return nil
}

func policyEndpoint(pt policyType, teamID string) (string, error) {
	if pt != policyTypeAlert && pt != policyTypeNotification {
		return "", fmt.Errorf("unsupported policy type %q", pt)
	}
	if teamID == "" {
		// Global scope only supports alert policies
		if pt != policyTypeAlert {
			return "", fmt.Errorf("global scope only supports alert policies")
		}
		return "/v1/alerts/policies", nil
	}
	if err := validatePolicyPathID(teamID); err != nil {
		return "", err
	}
	return "/v1/teams/" + teamID + "/policies", nil
}

// List follows pagination using only its query parameters. The request origin
// and endpoint are always generated from the provider configuration and scope.
func (c *policyOrderingAPI) List(ctx context.Context, teamID string) ([]policyPosition, error) {
	pt := c.policyType
	endpoint, err := policyEndpoint(pt, teamID)
	if err != nil {
		return nil, err
	}
	query := url.Values{"type": {string(pt)}}
	seenPages := make(map[string]bool)
	seenIDs := make(map[string]bool)
	seenOrders := make(map[int64]bool)
	positions := make([]policyPosition, 0)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		pageKey := query.Encode()
		if seenPages[pageKey] {
			return nil, fmt.Errorf("%s policy pagination repeated a page", pt)
		}
		seenPages[pageKey] = true
		var page dto.AlertPolicyListDto
		req := c.newRequest(ctx).JoinBaseUrl(endpoint).Method(httpClient.GET).SetBodyParseObject(&page)
		inner := req.GetInnerRequest()
		inner.URL.RawQuery = query.Encode()
		response, sendErr := req.Send()
		if err := policyOrderResponseError(ctx, response, sendErr, fmt.Sprintf("list %s policies", pt)); err != nil {
			return nil, err
		}
		if page.Values == nil {
			return nil, fmt.Errorf("%s policy list response omitted the values array", pt)
		}
		for _, policy := range page.Values {
			if policy.Type != "" && policy.Type != string(pt) {
				return nil, fmt.Errorf("policy list returned type %q instead of %q", policy.Type, pt)
			}
			if err := validatePolicyPathID(policy.ID); err != nil {
				return nil, fmt.Errorf("invalid %s policy list response: %w", pt, err)
			}
			if seenIDs[policy.ID] {
				return nil, fmt.Errorf("%s policy list returned duplicate ID %q", pt, policy.ID)
			}
			if math.IsNaN(policy.Order) || math.IsInf(policy.Order, 0) || policy.Order < 1 || policy.Order >= float64(math.MaxInt64) || math.Trunc(policy.Order) != policy.Order {
				return nil, fmt.Errorf("%s policy %q has a missing or invalid position", pt, policy.ID)
			}
			order := int64(policy.Order)
			if seenOrders[order] {
				return nil, fmt.Errorf("%s policy list returned duplicate position %d; retry after concurrent changes finish", pt, order)
			}
			seenIDs[policy.ID], seenOrders[order] = true, true
			positions = append(positions, policyPosition{ID: policy.ID, Order: order})
		}
		if page.Links.Next == "" {
			break
		}
		next, err := url.Parse(page.Links.Next)
		if err != nil || next.User != nil || next.Fragment != "" {
			return nil, fmt.Errorf("%s policy list returned an invalid pagination link", pt)
		}
		// Absolute, root-relative, and query-only links are supported, but
		// changing the authenticated destination or policy scope is not.
		currentURL := inner.URL
		if (next.Host != "" && next.Host != currentURL.Host) || (next.Scheme != "" && next.Scheme != currentURL.Scheme) || (next.Path != "" && next.Path != endpoint && next.Path != currentURL.Path) {
			return nil, fmt.Errorf("%s policy pagination link changed the API origin or scope", pt)
		}
		query, err = url.ParseQuery(next.RawQuery)
		if err != nil || len(query) == 0 {
			return nil, fmt.Errorf("%s policy list returned invalid pagination parameters", pt)
		}
		if values, ok := query["type"]; ok && (len(values) != 1 || values[0] != string(pt)) {
			return nil, fmt.Errorf("%s policy pagination link changed the policy type", pt)
		}
		query.Set("type", string(pt))
	}
	sort.Slice(positions, func(i, j int) bool { return positions[i].Order < positions[j].Order })
	for index, policy := range positions {
		if policy.Order != int64(index+1) {
			return nil, fmt.Errorf("%s policy list has missing positions; retry after concurrent changes finish", pt)
		}
	}
	return positions, nil
}

func (c *policyOrderingAPI) Move(ctx context.Context, teamID, policyID string, index int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	pt := c.policyType
	endpoint, err := policyEndpoint(pt, teamID)
	if err != nil {
		return err
	}
	if err := validatePolicyPathID(policyID); err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("%s policy position must be a nonnegative API index", pt)
	}
	req := c.newRequest(ctx).JoinBaseUrl(endpoint + "/" + policyID + "/change-order").
		Method(httpClient.POST).SetBody(map[string]int64{"order": index})
	response, sendErr := req.Send()
	if response != nil {
		defer response.Discard()
	}
	if err := policyOrderResponseError(ctx, response, sendErr, fmt.Sprintf("move %s policy", pt)); err != nil {
		return err
	}
	// Move has no decoded result, but its response body still needs draining
	// and closing so sequential moves can reuse connections.
	if _, err := response.Body(); err != nil {
		return fmt.Errorf("unable to read %s policy move response", pt)
	}
	return nil
}

func policyOrderResponseError(ctx context.Context, response *httpClient.Response, err error, operation string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if response != nil && (response.GetStatusCode() < http.StatusOK || response.GetStatusCode() >= http.StatusMultipleChoices) {
		return fmt.Errorf("unable to %s: HTTP %d", operation, response.GetStatusCode())
	}
	if err != nil || response == nil {
		// Do not include transport URLs or raw response bodies in diagnostics.
		return fmt.Errorf("unable to %s: request failed or response could not be decoded", operation)
	}
	return nil
}
