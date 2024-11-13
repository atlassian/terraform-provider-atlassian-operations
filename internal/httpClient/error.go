package httpClient

import (
	"encoding/json"
	"fmt"
)

type (
	HttpErrorResponse interface {
		Error() string
		UnmarshalJSON([]byte) error
	}

	opsClientDefaultErrorResponse struct {
		Errors []opsClientDefaultError `json:"errors"`
	}

	opsClientDefaultError struct {
		Title string `json:"title"`
		Code  string  `json:"code"`
	}

	opsClientUnauthorizedErrorResponse struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	}

	teamClientDefaultErrorResponse struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	teamEnableOpsClientDefaultErrorResponse struct {
		Message string `json:"message"`
	}

	userClientDefaultErrorResponse struct {
		ErrorMessages []string `json:"errorMessages"`
		Errors        any      `json:"errors"`
		Status        int32    `json:"status"`
	}
)

func (e *opsClientDefaultErrorResponse) Error() string {
	errMsg := ""
	for _, err := range e.Errors {
		errMsg += fmt.Sprintf("Error: %s, Code: %s\n", err.Title, err.Code)
	}
	return errMsg
}

func (e *opsClientUnauthorizedErrorResponse) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

func (e *teamClientDefaultErrorResponse) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s", e.Code, e.Message)
}

func (e *teamEnableOpsClientDefaultErrorResponse) Error() string {
	return fmt.Sprintf("Message: %s", e.Message)
}

func (e *userClientDefaultErrorResponse) Error() string {
	errMsg := ""
	for _, err := range e.ErrorMessages {
		errMsg += fmt.Sprintf("Error: %s\n", err)
	}
	if e.Errors != nil {
		res, err := json.Marshal(e.Errors)
		if err == nil {
			errMsg += fmt.Sprintf("Errors: %s\n", string(res))
		}
	}
	return errMsg
}

func (e *opsClientDefaultErrorResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v["errors"] != nil {
		for err2Key, err2RawVal := range v["errors"].(map[string]interface{}) {
			errParsed, ok := err2RawVal.(map[string]interface{})
			if !ok {
				e.Errors = append(e.Errors, struct {
					Title string `json:"title"`
					Code  string `json:"code"`
				}{
					Title: err2Key,
					Code:  err2RawVal.(string),
				})
			} else {
				e.Errors = append(e.Errors, struct {
					Title string `json:"title"`
					Code  string `json:"code"`
				}{
					Title: errParsed["title"].(string),
					Code:  errParsed["code"].(string),
				})
			}
		}
	}
	return nil
}

func (e *opsClientUnauthorizedErrorResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v["code"] != nil {
		e.Code = int32(v["code"].(float64))
	}
	if v["message"] != nil {
		e.Message = v["message"].(string)
	}

	return nil
}

func (e *teamClientDefaultErrorResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v["code"] != nil {
		e.Code = v["code"].(string)
	}
	if v["message"] != nil {
		e.Message = v["message"].(string)
	}

	return nil
}

func (e *teamEnableOpsClientDefaultErrorResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v["message"] != nil {
		e.Message = v["message"].(string)
	}

	return nil
}

func (e *userClientDefaultErrorResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v["status"] != nil {
		e.Status = v["status"].(int32)
	}
	if v["errorMessages"] != nil {
		rawErrors := v["errorMessages"].(map[string]interface{})
		e.ErrorMessages = make([]string, len(rawErrors))
		for _, item := range rawErrors {
			e.ErrorMessages = append(e.ErrorMessages, item.(string))
		}
	}
	if v["errors"] != nil {
		e.Errors = v["errors"]
	}

	return nil
}

func NewOpsClientErrorMap() ErrorCodeToObjectMap {
	return ErrorCodeToObjectMap{
		400: &opsClientDefaultErrorResponse{},
		401: &opsClientUnauthorizedErrorResponse{},
		402: &opsClientDefaultErrorResponse{},
		403: &opsClientDefaultErrorResponse{},
		404: &opsClientDefaultErrorResponse{},
		409: &opsClientDefaultErrorResponse{},
		422: &opsClientDefaultErrorResponse{},
		429: &opsClientDefaultErrorResponse{},
	}
}

func NewTeamClientErrorMap() ErrorCodeToObjectMap {
	return ErrorCodeToObjectMap{
		400: &teamClientDefaultErrorResponse{},
		403: &teamClientDefaultErrorResponse{},
		404: &teamClientDefaultErrorResponse{},
		410: &teamClientDefaultErrorResponse{},
		413: &teamClientDefaultErrorResponse{},
		415: &teamClientDefaultErrorResponse{},
		422: &teamClientDefaultErrorResponse{},
	}
}

func NewTeamEnableOpsClientErrorMap() ErrorCodeToObjectMap {
	return ErrorCodeToObjectMap{
		400: &teamEnableOpsClientDefaultErrorResponse{},
		403: &teamEnableOpsClientDefaultErrorResponse{},
		404: &teamEnableOpsClientDefaultErrorResponse{},
		410: &teamEnableOpsClientDefaultErrorResponse{},
		413: &teamEnableOpsClientDefaultErrorResponse{},
		415: &teamEnableOpsClientDefaultErrorResponse{},
		422: &teamEnableOpsClientDefaultErrorResponse{},
	}
}

func NewUserClientErrorMap() ErrorCodeToObjectMap {
	return ErrorCodeToObjectMap{
		400: &userClientDefaultErrorResponse{},
		401: &userClientDefaultErrorResponse{},
		429: &userClientDefaultErrorResponse{},
	}
}
