package dto

type CustomRoleDto struct {
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name"`
	GrantedRights    []string `json:"grantedRights"`
	DisallowedRights []string `json:"disallowedRights"`
}

type CustomRoleCUDResponseDto struct {
	Message string                       `json:"message"`
	Data    CustomRoleDataCUDResponseDto `json:"data"`
}

type CustomRoleDataCUDResponseDto struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
