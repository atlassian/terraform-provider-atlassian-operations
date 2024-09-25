package dto

import "encoding/json"

type ErrorResponse struct {
	Message string
	Code    int32
}

func (e *ErrorResponse) Error() string {
	res, _ := json.Marshal(e)
	return string(res)
}
