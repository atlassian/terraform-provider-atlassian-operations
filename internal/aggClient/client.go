// Package aggClient calls the Atlassian GraphQL Gateway (AGG). The provider uses it to manage teams when the
// configured token is a scoped API token, which the Teams REST API rejects.
package aggClient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto/agg"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrTeamNotFound is returned when AGG explicitly reports a team as missing or deleted.
var ErrTeamNotFound = errors.New("team not found")

// errRejected marks an answer from the gateway that rejected the request (4xx): the mutation did not run.
var errRejected = errors.New("the gateway rejected the request")

type Client struct {
	config dto.AtlassianOpsProviderModel
	// newRequest builds the HTTP request for the gateway; tests replace it to point the client at a fake gateway.
	newRequest func(dto.AtlassianOpsProviderModel) *httpClient.Request
}

func New(config dto.AtlassianOpsProviderModel) Client {
	return Client{config: config, newRequest: httpClientHelpers.GenerateAGGClientRequest}
}

// query posts a read operation. Reads are retried according to api_retry_count, like every other request of the
// provider.
func (c Client) query(ctx context.Context, operationName, document string, variables map[string]any) (agg.Response, error) {
	return c.send(ctx, operationName, document, variables, true)
}

// mutate posts a write operation once, whatever api_retry_count says (it defaults to 3), except after a 429, which
// means the gateway did not run the request (retryablehttp's default backoff already honours Retry-After): a retry
// after a lost response would apply the change a second time, and a repeated CreateTeam would leave a second team
// behind. Every write goes through here, so the decision never depends on the text of the GraphQL document.
func (c Client) mutate(ctx context.Context, operationName, document string, variables map[string]any) (agg.Response, error) {
	return c.send(ctx, operationName, document, variables, false)
}

// send posts one GraphQL operation and decodes the response.
func (c Client) send(ctx context.Context, operationName, document string, variables map[string]any, retry bool) (agg.Response, error) {
	tflog.Trace(ctx, fmt.Sprintf("Sending AGG %s request", operationName))

	req := c.newRequest(c.config)
	req.SetContext(ctx)
	if !retry {
		req.SetRetryPolicy(func(resp *httpClient.Response, err error) bool {
			return err == nil && resp != nil && resp.GetStatusCode() == http.StatusTooManyRequests
		})
	}

	var resp agg.Response
	httpResp, err := req.
		Method(httpClient.POST).
		SetBody(agg.Request{Query: document, OperationName: operationName, Variables: variables}).
		SetBodyParseObject(&resp).
		Send()

	if err != nil {
		return resp, fmt.Errorf("Teams GraphQL API %s request failed: %w", operationName, err)
	}
	if httpResp.IsError() {
		status := httpResp.GetStatusCode()
		msg := fmt.Sprintf("Teams GraphQL API %s returned status %d", operationName, status)
		if errBody := httpResp.GetErrorBody(); errBody != nil {
			msg += ": " + *errBody
		}
		if status >= 400 && status < 500 && status != http.StatusRequestTimeout {
			return resp, fmt.Errorf("%s (%w)", msg, errRejected)
		}
		return resp, errors.New(msg)
	}
	// Top-level errors can be non-fatal warnings; callers decide based on whether the payload came back.
	if len(resp.Errors) > 0 {
		tflog.Warn(ctx, fmt.Sprintf("AGG %s returned messages: %s", operationName, errorMessages(resp.Errors)))
	}
	return resp, nil
}

func errorMessages(errs []agg.Error) string {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// payloadErrorText describes why a mutation failed: its errors, then the members it could not add or remove.
// addMembers and removeMembers report a rejected member only there, with success=false and an empty errors list.
func payloadErrorText(payload *agg.MutationPayload) string {
	parts := make([]string, 0, len(payload.Errors)+len(payload.FailedMembers))
	for _, e := range payload.Errors {
		parts = append(parts, e.Message)
	}
	for _, m := range payload.FailedMembers {
		parts = append(parts, fmt.Sprintf("member %s: %s (%s)", m.AccountId, m.Code, m.Message))
	}
	return strings.Join(parts, "; ")
}

// checkMutation turns a mutation payload into an error when the mutation did not succeed.
func checkMutation(operationName string, payload *agg.MutationPayload, topLevel []agg.Error) error {
	if payload == nil {
		if len(topLevel) > 0 {
			return fmt.Errorf("Teams GraphQL API %s errors: %s", operationName, errorMessages(topLevel))
		}
		return fmt.Errorf("Teams GraphQL API %s returned no result", operationName)
	}
	if text := payloadErrorText(payload); text != "" {
		return fmt.Errorf("Teams GraphQL API %s errors: %s", operationName, text)
	}
	if !payload.Success {
		return fmt.Errorf("Teams GraphQL API %s reported success=false without an error message", operationName)
	}
	return nil
}

// teamGone reports whether one of the errors says the team no longer exists: 404 (unknown) or 410 (deleted).
// The gateway sends the status on both top-level teamV2 errors and mutation payload errors (verified live 2026-11-03).
func teamGone(errs []agg.Error) bool {
	for _, e := range errs {
		ext := e.Extensions
		if ext.StatusCode == 404 || ext.StatusCode == 410 || ext.ErrorType == "Gone" || ext.ErrorType == "NotFound" ||
			ext.Classification == "Gone" || ext.Classification == "NotFound" {
			return true
		}
	}
	return false
}
