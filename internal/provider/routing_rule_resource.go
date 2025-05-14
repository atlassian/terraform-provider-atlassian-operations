package provider

import (
	"context"
	"fmt"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/dto"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient/httpClientHelpers"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/dataModels"
	"github.com/atlassian/terraform-provider-atlassian-operations/internal/provider/schemaAttributes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

var _ resource.Resource = &RoutingRuleResource{}
var _ resource.ResourceWithImportState = &RoutingRuleResource{}

func NewRoutingRuleResource() resource.Resource {
	return &RoutingRuleResource{}
}

type RoutingRuleResource struct {
	clientConfiguration dto.AtlassianOpsProviderModel
}

func (r *RoutingRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_rule"
}

func (r *RoutingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: schemaAttributes.RoutingRuleResourceAttributes,
	}
}

func (r *RoutingRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(dto.AtlassianOpsProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *JsmOpsClient, got: %T", req.ProviderData),
		)
		return
	}

	r.clientConfiguration = client
}

func (r *RoutingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dataModels.RoutingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	ruleDto := RoutingRuleModelToDto(ctx, data)

	// Create routing rule
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/routing-rules", data.TeamID.ValueString())).
		Method(httpClient.POST).
		SetBody(ruleDto).
		SetBodyParseObject(&ruleDto).
		Send()

	handleHttpResponse(httpResp, err, "create routing rule", &resp.Diagnostics, ctx)

	// Update state with response
	data = RoutingRuleDtoToModel(data.TeamID.ValueString(), ruleDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoutingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dataModels.RoutingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get routing rule
	var ruleDto dto.RoutingRuleDto
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/routing-rules/%s", data.TeamID.ValueString(), data.ID.ValueString())).
		Method(httpClient.GET).
		SetBodyParseObject(&ruleDto).
		Send()

	handleHttpResponse(httpResp, err, "read routing rule", &resp.Diagnostics, ctx)

	// Update state
	data = RoutingRuleDtoToModel(data.TeamID.ValueString(), ruleDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoutingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dataModels.RoutingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to DTO
	ruleDto := RoutingRuleModelToDto(ctx, data)

	// Update routing rule
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/routing-rules/%s", data.TeamID.ValueString(), data.ID.ValueString())).
		Method(httpClient.PATCH).
		SetBody(ruleDto).
		SetBodyParseObject(&ruleDto).
		Send()

	handleHttpResponse(httpResp, err, "update routing rule", &resp.Diagnostics, ctx)

	// Update state
	data = RoutingRuleDtoToModel(data.TeamID.ValueString(), ruleDto)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoutingRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dataModels.RoutingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete routing rule
	httpResp, err := httpClientHelpers.
		GenerateJsmOpsClientRequest(r.clientConfiguration).
		JoinBaseUrl(fmt.Sprintf("/v1/teams/%s/routing-rules/%s", data.TeamID.ValueString(), data.ID.ValueString())).
		Method(httpClient.DELETE).
		Send()

	handleHttpResponse(httpResp, err, "delete routing rule", &resp.Diagnostics, ctx)
}

func (r *RoutingRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id,team_id. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), idParts[1])...)
}

func handleHttpResponse(httpResp *httpClient.Response, err error, s string, d *diag.Diagnostics, ctx context.Context) {
	if httpResp == nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to %s, got nil response", s))
		d.AddError("Client Error", fmt.Sprintf("Unable to %s, got nil response", s))
	} else if httpResp.IsError() {
		statusCode := httpResp.GetStatusCode()
		errorResponse := httpResp.GetErrorBody()
		if errorResponse != nil {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to %s, status code: %d. Got response: %s", s, statusCode, *errorResponse))
			d.AddError("Client Error", fmt.Sprintf("Unable to %s, status code: %d. Got response: %s", s, statusCode, *errorResponse))
		} else {
			tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to %s, got http response: %d", s, statusCode))
			d.AddError("Client Error", fmt.Sprintf("Unable to %s, got http response: %d", s, statusCode))
		}
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Client Error. Unable to %s, got error: %s", s, err.Error()))
		d.AddError("Client Error", fmt.Sprintf("Unable to %s, got error: %s", s, err.Error()))
	}
	if d.HasError() {
		return
	}
}
