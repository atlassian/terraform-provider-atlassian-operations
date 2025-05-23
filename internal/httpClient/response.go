package httpClient

import (
	"io"
	"net/http"
)

type Response struct {
	nativeResponse *http.Response
	errorBody      *string
}

func (receiver *Response) Discard() error {
	if receiver.nativeResponse != nil {
		err := receiver.nativeResponse.Body.Close()
		if err != nil {
			return err
		}
		receiver.nativeResponse = nil
	}
	return nil
}

func (receiver *Response) parseErrorBody() error {
	body, err := receiver.Body()
	if err != nil {
		return err
	}
	bodyString := string(body)
	receiver.errorBody = &bodyString
	return nil
}

func (receiver *Response) Body() ([]byte, error) {
	if receiver.nativeResponse != nil {
		defer receiver.nativeResponse.Body.Close()
		return io.ReadAll(receiver.nativeResponse.Body)
	}
	return nil, nil
}

func (receiver *Response) GetErrorBody() *string {
	return receiver.errorBody
}

func (receiver *Response) GetStatusCode() int {
	if receiver.nativeResponse == nil {
		return -1
	}
	return receiver.nativeResponse.StatusCode
}

func (receiver *Response) IsError() bool {
	return receiver.GetStatusCode() == -1 || receiver.GetStatusCode() > 399
}

func (receiver *Response) GetNativeResponse() *http.Response {
	return receiver.nativeResponse
}
