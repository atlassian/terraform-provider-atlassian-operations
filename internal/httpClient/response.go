package httpClient

import (
	"io"
	"net/http"
)

type Response struct {
	nativeResponse *http.Response
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

func (receiver *Response) Body() ([]byte, error) {
	if receiver.nativeResponse != nil {
		defer receiver.nativeResponse.Body.Close()
		return io.ReadAll(receiver.nativeResponse.Body)
	}
	return nil, nil
}

func (receiver *Response) GetStatusCode() int {
	return receiver.nativeResponse.StatusCode
}

func (receiver *Response) IsError() bool {
	if receiver.nativeResponse != nil {
		return receiver.GetStatusCode() > 399
	} else {
		return true
	}
}

func (receiver *Response) GetNativeResponse() *http.Response {
	return receiver.nativeResponse
}
