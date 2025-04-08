package dto

type UserContactDto struct {
	ID      string `json:"id,omitempty"`
	Method  string `json:"method"`
	To      string `json:"to"`
	Enabled bool   `json:"enabled"`
}

type UserContactCUDResponseDto struct {
	Message string                        `json:"message"`
	Data    UserContactDataCUDResponseDto `json:"data"`
}

type UserContactDataCUDResponseDto struct {
	ID string `json:"id"`
}

type UserContactDataStatusReadResponseDto struct {
	Enabled bool `json:"enabled"`
}

type UserContactDataReadResponseDto struct {
	ID     string                               `json:"id"`
	Method string                               `json:"method"`
	To     string                               `json:"to"`
	Status UserContactDataStatusReadResponseDto `json:"status"`
}
