package dto

type (
	paginationLinks struct {
		Next string `json:"next"`
	}

	publicApiPageInfoAccountId struct {
		EndCursor   string `json:"endCursor"`
		HasNextPage bool   `json:"hasNextPage"`
	}

	ListResponse[T any] struct {
		Values  []T             `json:"values"`
		Links   paginationLinks `json:"links"`
		Expands []string        `json:"_expands"`
	}

	TeamMemberListResponse struct {
		PageInfo publicApiPageInfoAccountId `json:"pageInfo"`
		Results  []TeamMember               `json:"results"`
	}

	TeamMemberListRequest struct {
		After string `json:"after,omitempty"`
		First int32  `json:"first"`
	}

	PublicApiMembershipAddResponse struct {
		Members []TeamMember             `json:"members"`
		Errors  []map[string]interface{} `json:"errors"`
	}

	PublicApiMembershipRemoveResponse struct {
		Errors []map[string]interface{} `json:"errors"`
	}
)

func DefaultTeamMemberListRequest() TeamMemberListRequest {
	return TeamMemberListRequest{
		After: "",
		First: 50,
	}
}
