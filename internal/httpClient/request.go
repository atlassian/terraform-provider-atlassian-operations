package httpClient

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/go-retryablehttp"
	"net/http"
	"time"
)

type (
	OnRetryFunc        func(*Request) error
	RetryConditionFunc func(*Response, error) bool
	RequestMethod      string
	Request            struct {
		innerRequest    *retryablehttp.Request
		innerClient     *retryablehttp.Client
		parseBodyObject any
		response        *Response
		onRetryFuncs    []OnRetryFunc
		retryConditions []RetryConditionFunc
	}
)

const (
	GET    RequestMethod = http.MethodGet
	POST   RequestMethod = http.MethodPost
	PUT    RequestMethod = http.MethodPut
	DELETE RequestMethod = http.MethodDelete
	PATCH  RequestMethod = http.MethodPatch
)

func NewRequest() *Request {

	inReq, _ := retryablehttp.NewRequest(http.MethodGet, "", nil)

	newReq := &Request{
		innerRequest:    inReq,
		innerClient:     retryablehttp.NewClient(),
		onRetryFuncs:    make([]OnRetryFunc, 0),
		retryConditions: make([]RetryConditionFunc, 0),
	}
	newReq.SetHeader("Content-Type", "application/json")
	newReq.innerClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		return newReq.shouldRetryBecauseCondition(ctx, &Response{nativeResponse: resp}, err)
	}
	newReq.innerClient.PrepareRetry = func(req *http.Request) error {
		for _, fun := range newReq.onRetryFuncs {
			err := fun(newReq)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return newReq
}

func (receiver *Request) GetInnerClient() *retryablehttp.Client {
	if receiver.innerClient == nil {
		receiver.innerClient = retryablehttp.NewClient()
	}
	return receiver.innerClient
}

func (receiver *Request) SetRetryCount(count int) *Request {
	receiver.innerClient.RetryMax = count
	return receiver
}

func (receiver *Request) SetRetryWaitTime(waitTime time.Duration) *Request {
	receiver.innerClient.RetryWaitMin = waitTime
	return receiver
}

func (receiver *Request) SetRetryMaxWaitTime(maxWaitTime time.Duration) *Request {
	receiver.innerClient.RetryWaitMax = maxWaitTime
	return receiver
}

func (receiver *Request) AddRetryHook(hook OnRetryFunc) *Request {
	receiver.onRetryFuncs = append(receiver.onRetryFuncs, hook)
	return receiver
}

func (receiver *Request) AddRetryCondition(condition RetryConditionFunc) *Request {
	receiver.retryConditions = append(receiver.retryConditions, condition)
	return receiver
}

func (receiver *Request) shouldRetryBecauseCondition(ctx context.Context, resp *Response, err error) (bool, error) {
	shouldRetry, _ := retryablehttp.DefaultRetryPolicy(ctx, resp.nativeResponse, err)
	if !shouldRetry {
		for _, fun := range receiver.retryConditions {
			if fun(resp, err) {
				shouldRetry = true
				break
			}
		}
	} else if err != nil {
		var invalidUnmarshalError *json.InvalidUnmarshalError
		if errors.As(err, &invalidUnmarshalError) {
			shouldRetry = false

		}
	}
	return shouldRetry, err
}

func (receiver *Request) GetInnerClient() *retryablehttp.Client {
	if receiver.innerClient == nil {
		receiver.innerClient = retryablehttp.NewClient()
	}
	return receiver.innerClient
}

func (receiver *Request) SetRetryCount(count int) *Request {
	receiver.innerClient.RetryMax = count
	return receiver
}

func (receiver *Request) SetRetryWaitTime(waitTime time.Duration) *Request {
	receiver.innerClient.RetryWaitMin = waitTime
	return receiver
}

func (receiver *Request) SetRetryMaxWaitTime(maxWaitTime time.Duration) *Request {
	receiver.innerClient.RetryWaitMax = maxWaitTime
	return receiver
}

func (receiver *Request) AddRetryHook(hook OnRetryFunc) *Request {
	receiver.onRetryFuncs = append(receiver.onRetryFuncs, hook)
	return receiver
}

func (receiver *Request) AddRetryCondition(condition RetryConditionFunc) *Request {
	receiver.retryConditions = append(receiver.retryConditions, condition)
	return receiver
}

func (receiver *Request) shouldRetryBecauseCondition(ctx context.Context, resp *Response, err error) (bool, error) {
	shouldRetry, _ := retryablehttp.DefaultRetryPolicy(ctx, resp.nativeResponse, err)
	if !shouldRetry {
		for _, fun := range receiver.retryConditions {
			if fun(resp, err) {
				shouldRetry = true
				break
			}
		}
	} else if err != nil {
		var invalidUnmarshalError *json.InvalidUnmarshalError
		if errors.As(err, &invalidUnmarshalError) {
			shouldRetry = false
		}
	}
	return shouldRetry, err
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

func (r *Request) SetUrl(url string) *Request {
	parse, err := r.innerRequest.URL.Parse(url)
	if err != nil {
		return nil
	}
	r.innerRequest.URL = parse
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

func (r *Request) Send() (*Response, error) {
	r.innerRequest.SetResponseHandler(func(resp *http.Response) error {
		var retErr error = nil
		clientResp := &Response{nativeResponse: resp}
		if clientResp.IsError() {
			err := clientResp.parseErrorBody()
			if err != nil {
				retErr = err
			}
		} else if r.parseBodyObject != nil {
			retErr = parseBody(r.parseBodyObject, clientResp)
		}
		r.response = clientResp
		return retErr
	})
	r.innerClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		if resp != nil {
			clientResp := Response{nativeResponse: resp}
			r.response = &clientResp
			parseErrorBodyError := clientResp.parseErrorBody()
			if parseErrorBodyError != nil {
				err = parseErrorBodyError
			}
		}
		return resp, err
	}

	_, err := r.innerClient.Do(r.innerRequest)
	return r.response, err
}
