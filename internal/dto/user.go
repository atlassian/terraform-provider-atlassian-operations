package dto

const (
	AccountTypeAtlassian = AccountType("atlassian")
	AccountTypeApp       = AccountType("app")
	AccountTypeCustomer  = AccountType("customer")
	Unknown              = AccountType("unknown")
)

type (
	AccountType  string
	GroupNameDto struct {
		GroupId string `json:"groupId"`
		Name    string `json:"name"`
		Self    string `json:"self"`
	}
	ApplicationRoleDto struct {
		DefaultGroups        []string       `json:"defaultGroups"`
		DefaultGroupsDetails []GroupNameDto `json:"defaultGroupsDetails"`
		Defined              bool           `json:"defined"`
		GroupDetails         []GroupNameDto `json:"groupDetails"`
		Groups               []string       `json:"groups"`
		HasUnlimitedSeats    bool           `json:"hasUnlimitedSeats"`
		Key                  string         `json:"key"`
		Name                 string         `json:"name"`
		NumberOfSeats        int32          `json:"numberOfSeats"`
		Platform             bool           `json:"platform"`
	}
	AvatarUrlsBeanDto struct {
		A16x16 string `json:"16x16"`
		A24x24 string `json:"24x24"`
		A32x32 string `json:"32x32"`
		A48x48 string `json:"48x48"`
	}
	UserDto struct {
		AccountId        string      `json:"accountId"`
		AccountType      AccountType `json:"AccountType"`
		Active           bool        `json:"active"`
		ApplicationRoles struct {
			Items      []ApplicationRoleDto `json:"items"`
			MaxResults int32                `json:"maxResults"`
			Size       int32                `json:"size"`
		} `json:"applicationRoles"`
		AvatarUrls   AvatarUrlsBeanDto `json:"avatarUrls"`
		DisplayName  string            `json:"displayName"`
		EmailAddress string            `json:"emailAddress"`
		Expand       string            `json:"expand"`
		Groups       struct {
			Items      []GroupNameDto `json:"items"`
			MaxResults int32          `json:"maxResults"`
			Size       int32          `json:"size"`
		} `json:"groups"`
		Locale   string `json:"locale"`
		TimeZone string `json:"timeZone"`
	}
)
