package httpClient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"net/http"
	urlpkg "net/url"
)

type (
	RequestMethod string
	Request       struct {
		innerRequest    *retryablehttp.Request
		parseBodyObject any
		response        *Response
		Client          *HttpClient
	}
)

const (
	GET    RequestMethod = http.MethodGet
	POST   RequestMethod = http.MethodPost
	PUT    RequestMethod = http.MethodPut
	DELETE RequestMethod = http.MethodDelete
	PATCH  RequestMethod = http.MethodPatch
)

func NewRequest(client *HttpClient) *Request {
	inReq, _ := retryablehttp.NewRequest("", "", nil)
	newReq := &Request{
		innerRequest: inReq,
		Client:       client,
	}
	newReq.SetHeader("Content-Type", "application/json")
	client.innerClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		return client.shouldRetryBecauseCondition(ctx, &Response{nativeResponse: resp}, err)
	}
	client.innerClient.PrepareRetry = func(req *http.Request) error {
		for _, fun := range client.onRetryFuncs {
			err := fun(newReq)
			if err != nil {
				return err
			}
		}
		return nil
	}
	client.innerClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		if err == nil {
			return resp, fmt.Errorf("%s request giving up after %d attempt(s)", resp.Request.Method, numTries)
		}
		return resp, err
	}
	return newReq
}

func (r *Request) SetBasicAuth(username, password string) *Request {
	r.innerRequest.SetBasicAuth(username, password)
	return r
}

func (r *Request) SetBearerAuth(token string) *Request {
	r.innerRequest.Header.Set("Authorization", "Bearer "+token)
	return r
}

func (r *Request) SetOAuth2Auth(token string) *Request {
	r.innerRequest.Header.Set("Authorization", "OAuth2 "+token)
	return r
}

func (r *Request) Url(url string) *Request {
	parsedURL, _ := urlpkg.Parse(url)
	parsedURL.RawQuery = r.innerRequest.URL.RawQuery
	r.innerRequest.URL = parsedURL
	return r
}

func (r *Request) JoinBaseUrl(url string) *Request {
	innerReq := r.GetInnerRequest()
	innerReq.URL = innerReq.URL.JoinPath(url)
	return r
}

func (r *Request) Method(method RequestMethod) *Request {
	r.innerRequest.Method = string(method)
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	r.innerRequest.Header.Set(key, value)
	return r
}

func (r *Request) SetQueryParam(param, value string) *Request {
	queries := r.innerRequest.URL.Query()
	if value == "" {
		queries.Del(param)
	} else {
		queries.Set(param, value)
	}
	r.innerRequest.URL.RawQuery = queries.Encode()
	return r
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

func (r *Request) SetBody(body interface{}) *Request {
	rawBody, _ := json.Marshal(body)
	_ = r.innerRequest.SetBody(rawBody)
	return r
}

func (r *Request) GetInnerRequest() *retryablehttp.Request {
	return r.innerRequest
}

func (r *Request) SetBodyParseObject(t interface{}) *Request {
	r.parseBodyObject = t
	return r
}

func parseBody(t interface{}, resp *Response) error {
	if t != nil {
		body, err := resp.Body()
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, t)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseErrorBody(resp *Response) (*string, error) {
	body, err := resp.Body()
	if err != nil {
		return nil, err
	}
	bodyString := string(body)
	return &bodyString, nil
}

func (r *Request) Send() (*Response, error) {
	r.innerRequest.SetResponseHandler(func(resp *http.Response) error {
		var retErr error = nil
		clientResp := Response{nativeResponse: resp}
		if clientResp.IsError() {
			errorBody, err := parseErrorBody(&clientResp)
			if err != nil {
				retErr = err
			} else {
				clientResp.errorBody = errorBody
			}
		} else if r.parseBodyObject != nil {
			retErr = parseBody(r.parseBodyObject, &clientResp)
		}
		r.response = &clientResp
		return retErr
	})
	client := r.Client.GetInnerClient()
	_, err := client.Do(r.innerRequest)
	return r.response, err
}
