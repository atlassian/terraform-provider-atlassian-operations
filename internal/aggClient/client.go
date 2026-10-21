// Package aggClient calls the Atlassian GraphQL Gateway (AGG). The provider uses it to manage teams when the
// configured token is a scoped API token, which the Teams REST API rejects.
package aggClient

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto/agg"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrTeamNotFound is returned when AGG explicitly reports a team as missing or deleted.
var ErrTeamNotFound = errors.New("team not found")

type Client struct {
	config dto.AtlassianOpsProviderModel
}

func New(config dto.AtlassianOpsProviderModel) Client {
	return Client{config: config}
}

// send posts one GraphQL operation and decodes the response.
func (c Client) send(ctx context.Context, operationName, query string, variables map[string]any) (agg.Response, error) {
	tflog.Trace(ctx, fmt.Sprintf("Sending AGG %s request", operationName))

	var resp agg.Response
	httpResp, err := httpClientHelpers.
		GenerateAGGClientRequest(c.config).
		Method(httpClient.POST).
		SetBody(agg.Request{Query: query, OperationName: operationName, Variables: variables}).
		SetBodyParseObject(&resp).
		Send()

	if err != nil {
		return resp, fmt.Errorf("AGG %s request failed: %w", operationName, err)
	}
	if httpResp.IsError() {
		if errBody := httpResp.GetErrorBody(); errBody != nil {
			return resp, fmt.Errorf("AGG %s returned status %d: %s", operationName, httpResp.GetStatusCode(), *errBody)
		}
		return resp, fmt.Errorf("AGG %s returned status %d", operationName, httpResp.GetStatusCode())
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

// checkMutation turns a mutation payload into an error when the mutation did not succeed.
func checkMutation(operationName string, payload *agg.MutationPayload, topLevel []agg.Error) error {
	if payload == nil {
		if len(topLevel) > 0 {
			return fmt.Errorf("AGG %s errors: %s", operationName, errorMessages(topLevel))
		}
		return fmt.Errorf("AGG %s returned no result", operationName)
	}
	if len(payload.Errors) > 0 {
		return fmt.Errorf("AGG %s errors: %s", operationName, errorMessages(payload.Errors))
	}
	if !payload.Success {
		return fmt.Errorf("AGG %s was not successful", operationName)
	}
	return nil
}

// isTeamGone reports whether an error says teamV2 found no team: unknown (NotFound) or deleted (Gone / TEAM_DELETED).
// Errors on other fields, or without that classification, do not count.
func isTeamGone(e agg.Error) bool {
	if len(e.Path) == 0 || fmt.Sprint(e.Path[len(e.Path)-1]) != "teamV2" {
		return false
	}
	ext := e.Extensions
	return ext.Classification == "NotFound" || ext.Classification == "Gone" || ext.ErrorCode == "TEAM_DELETED"
}
