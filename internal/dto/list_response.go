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
)
