package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/atlassian/terraform-provider-atlassian-operations/internal/httpClient"
)

func orderingTestAPI(t *testing.T, handler http.HandlerFunc) *policyOrderingAPI {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &policyOrderingAPI{policyType: policyTypeAlert, requestFactory: func() *httpClient.Request {
		req := httpClient.NewRequest().SetUrl(server.URL + "/jsm/ops/api/cloud").SetRetryCount(0)
		// Use the loopback server's transport rather than host proxy settings.
		req.GetInnerClient().HTTPClient = server.Client()
		return req
	}}
}

func TestPolicyOrderingAPIListAndMove(t *testing.T) {
	for _, tc := range []struct {
		pt     policyType
		teamID string
	}{{policyTypeAlert, ""}, {policyTypeAlert, "team-id"}, {policyTypeNotification, "team-id"}} {
		t.Run(string(tc.pt)+"/"+tc.teamID, func(t *testing.T) {
			teamID := tc.teamID
			endpoint, _ := policyEndpoint(tc.pt, teamID)
			prefix := "/jsm/ops/api/cloud" + endpoint
			calls := 0
			client := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == prefix:
					if r.URL.Query().Get("type") != string(tc.pt) {
						t.Error("missing policy type")
					}
					if r.URL.Query().Get("offset") == "2" {
						_, _ = w.Write([]byte(`{"values":[{"id":"a","order":1}],"links":{}}`))
					} else {
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"values": []map[string]interface{}{{"id": "b", "order": 2}},
							"links":  map[string]string{"next": prefix + "?offset=2&type=" + string(tc.pt)},
						})
					}
				case r.Method == http.MethodPost && r.URL.Path == prefix+"/b/change-order":
					var body struct {
						Order int64 `json:"order"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body.Order != 0 {
						t.Errorf("index=%d", body.Order)
					}
					_, _ = w.Write([]byte(`{}`))
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(400)
				}
			})
			client.policyType = tc.pt
			got, err := client.List(context.Background(), teamID)
			if err != nil {
				t.Fatal(err)
			}
			want := []policyPosition{{ID: "a", Order: 1}, {ID: "b", Order: 2}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v", got)
			}
			if err := client.Move(context.Background(), teamID, "b", 0); err != nil {
				t.Fatal(err)
			}
			if calls != 3 {
				t.Fatalf("requests=%d", calls)
			}
		})
	}
}

func TestPolicyOrderingAPIRejectsInvalidResponses(t *testing.T) {
	tests := map[string]string{
		"missing values":  `{}`,
		"null values":     `{"values":null}`,
		"missing order":   `{"values":[{"id":"a"}]}`,
		"zero order":      `{"values":[{"id":"a","order":0}]}`,
		"fractional":      `{"values":[{"id":"a","order":1.5}]}`,
		"overflow":        `{"values":[{"id":"a","order":1e30}]}`,
		"empty ID":        `{"values":[{"id":"","order":1}]}`,
		"duplicate ID":    `{"values":[{"id":"a","order":1},{"id":"a","order":2}]}`,
		"duplicate order": `{"values":[{"id":"a","order":1},{"id":"b","order":1}]}`,
		"malformed JSON":  `{`,
		"cycle":           `{"values":[],"links":{"next":"?type=alert"}}`,
		"foreign host":    `{"values":[],"links":{"next":"https://other.example/v1/alerts/policies?offset=2"}}`,
		"wrong scope":     `{"values":[],"links":{"next":"/v1/teams/other/policies?offset=2"}}`,
		"wrong type":      `{"values":[],"links":{"next":"?offset=2&type=notification"}}`,
		"invalid link":    `{"values":[],"links":{"next":"%zz"}}`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			client := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) { calls++; _, _ = w.Write([]byte(body)) })
			if _, err := client.List(context.Background(), ""); err == nil {
				t.Fatal("expected error")
			}
			if calls != 1 {
				t.Fatalf("unexpected retries/requests: %d", calls)
			}
		})
	}
}

func TestPolicyOrderingAPIErrorsAndCancellation(t *testing.T) {
	client := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"sensitive-body"}`))
	})
	for _, err := range []error{
		func() error { _, err := client.List(context.Background(), ""); return err }(),
		client.Move(context.Background(), "", "a", 0),
	} {
		if err == nil || !strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "sensitive-body") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.List(ctx, ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if err := client.Move(ctx, "", "a", 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if err := client.Move(context.Background(), "", "a", -1); err == nil {
		t.Fatal("negative index accepted")
	}
	if err := client.Move(context.Background(), "../team", "a", 0); err == nil {
		t.Fatal("invalid scope accepted")
	}
	if err := client.Move(context.Background(), "", "../a", 0); err == nil {
		t.Fatal("invalid ID accepted")
	}
}

func TestPolicyOrderingAPIRejectsRedirectsAndNonSuccess(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusNotModified} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			redirected := 0
			api := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/unexpected-scope" {
					redirected++
					_, _ = w.Write([]byte(`{}`))
					return
				}
				w.Header().Set("Location", "/unexpected-scope")
				w.WriteHeader(status)
			})
			if err := api.Move(context.Background(), "team", "a", 0); err == nil {
				t.Error("non-2xx move response accepted")
			}
			if _, err := api.List(context.Background(), "team"); err == nil {
				t.Fatal("non-2xx list response accepted")
			}
			if redirected != 0 {
				t.Fatalf("followed %d redirects outside the selected endpoint", redirected)
			}
		})
	}
}

func TestPolicyOrderingAPIGappedPositions(t *testing.T) {
	for _, body := range []string{
		`{"values":[{"id":"a","order":2}]}`,
		`{"values":[{"id":"a","order":1},{"id":"b","order":3}]}`,
	} {
		api := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) })
		if _, err := api.List(context.Background(), "team"); err == nil {
			t.Fatal("incomplete or gapped response could falsely satisfy a contiguous prefix")
		}
	}
}

func TestPolicyOrderingAPIRetriesAndDeadline(t *testing.T) {
	for _, operation := range []string{"list", "move"} {
		for _, status := range []int{http.StatusTooManyRequests, http.StatusServiceUnavailable} {
			t.Run(operation+"/"+http.StatusText(status), func(t *testing.T) {
				calls := 0
				api := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
					calls++
					if calls == 1 {
						w.WriteHeader(status)
						return
					}
					_, _ = w.Write([]byte(`{"values":[]}`))
				})
				factory := api.requestFactory
				api.requestFactory = func() *httpClient.Request {
					return factory().SetRetryCount(1).SetRetryWaitTime(time.Millisecond).SetRetryMaxWaitTime(time.Millisecond)
				}
				var err error
				if operation == "list" {
					_, err = api.List(context.Background(), "team")
				} else {
					err = api.Move(context.Background(), "team", "a", 0)
				}
				if err != nil || calls != 2 {
					t.Fatalf("calls=%d err=%v", calls, err)
				}
			})
		}
	}
	api := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := api.List(ctx, "team"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("request did not honor context deadline: %v", err)
	}
}

func TestPolicyOrderingAPITypeIsolation(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"wrong response type", `{"values":[{"id":"a","type":"alert","order":1}],"links":{}}`},
		{"pagination type switch", `{"values":[],"links":{"next":"?type=alert&offset=1"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := orderingTestAPI(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(tc.body)) })
			api.policyType = policyTypeNotification
			if _, err := api.List(context.Background(), "team"); err == nil {
				t.Fatal("cross-type response accepted")
			}
		})
	}
}
