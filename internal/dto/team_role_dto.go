// Copyright (c) HashiCorp, Inc.

package dto

type TeamRoleDto struct {
	ID     string          `json:"id,omitempty"`
	Name   string          `json:"name"`
	Rights []TeamRoleRight `json:"rights"`
}

type TeamRoleRight struct {
	Right   string `json:"right"`
	Granted bool   `json:"granted"`
}

type TeamRoleResponseDto struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Rights []TeamRoleRight `json:"rights"`
}
