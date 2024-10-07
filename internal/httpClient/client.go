package httpClient

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/go-retryablehttp"
	"net/url"
	"time"
)

type (
	authMethod string
	user       struct {
		Username string
		Password string
		Token    string
	}
	OnRetryFunc        func(*Request) error
	RetryConditionFunc func(*Response, error) bool
	HttpClient         struct {
		innerClient     *retryablehttp.Client
		baseUrl         *url.URL
		authMethod      authMethod
		authUser        *user
		onRetryFuncs    []OnRetryFunc
		retryConditions []RetryConditionFunc
	}
)

const (
	NoAuth     authMethod = ""
	BasicAuth  authMethod = "Basic"
	BearerAuth authMethod = "Bearer"
	OAuth2     authMethod = "OAuth2"
)

func NewHttpClient() *HttpClient {
	innerClient := retryablehttp.NewClient()
	client := &HttpClient{
		authMethod:  NoAuth,
		innerClient: innerClient,
		authUser: &user{
			Username: "",
			Password: "",
			Token:    "",
		},
		onRetryFuncs:    make([]OnRetryFunc, 0),
		retryConditions: make([]RetryConditionFunc, 0),
	}
	return client
}

func (receiver *HttpClient) GetInnerClient() *retryablehttp.Client {
	if receiver.innerClient == nil {
		receiver.innerClient = retryablehttp.NewClient()
	}
	return receiver.innerClient
}

func (receiver *HttpClient) SetRetryCount(count int) *HttpClient {
	receiver.innerClient.RetryMax = count
	return receiver
}

func (receiver *HttpClient) SetRetryWaitTime(waitTime time.Duration) *HttpClient {
	receiver.innerClient.RetryWaitMin = waitTime
	return receiver
}

func (receiver *HttpClient) SetRetryMaxWaitTime(maxWaitTime time.Duration) *HttpClient {
	receiver.innerClient.RetryWaitMax = maxWaitTime
	return receiver
}

func (receiver *HttpClient) SetBaseUrl(baseUrl string) *HttpClient {
	parsedURL, _ := url.Parse(baseUrl)
	receiver.baseUrl = parsedURL
	return receiver
}

func (receiver *HttpClient) AddRetryHook(hook OnRetryFunc) *HttpClient {
	receiver.onRetryFuncs = append(receiver.onRetryFuncs, hook)
	return receiver
}

func (receiver *HttpClient) AddRetryCondition(condition RetryConditionFunc) *HttpClient {
	receiver.retryConditions = append(receiver.retryConditions, condition)
	return receiver
}

func (receiver *HttpClient) SetDefaultBasicAuth(username string, password string) *HttpClient {
	receiver.authMethod = BasicAuth
	receiver.authUser.Username = username
	receiver.authUser.Password = password
	return receiver
}

func (receiver *HttpClient) SetDefaultBearerAuth(token string) *HttpClient {
	receiver.authMethod = BearerAuth
	receiver.authUser.Token = token
	return receiver
}

func (receiver *HttpClient) shouldRetryBecauseCondition(ctx context.Context, resp *Response, err error) (bool, error) {
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

func (receiver *HttpClient) NewRequest() *Request {
	req := NewRequest(receiver)
	if receiver.baseUrl != nil {
		req.innerRequest.URL = receiver.baseUrl
	}
	if receiver.authMethod != NoAuth && receiver.authUser != nil {
		switch receiver.authMethod {
		case BasicAuth:
			req.SetBasicAuth(receiver.authUser.Username, receiver.authUser.Password)
		case BearerAuth:
			req.SetBearerAuth(receiver.authUser.Token)
		case OAuth2:
			req.SetOAuth2Auth(receiver.authUser.Token)
		}
	}
	return req
}
