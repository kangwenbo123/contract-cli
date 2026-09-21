package contract_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/openplatform"
	"cn.qfei/contract-cli/internal/openplatform/contract"
)

func TestListSearchFilterFieldsUsesMCPQueryAndPreservesResponse(t *testing.T) {
	t.Parallel()

	for _, keyword := range []string{"", " 项目区域 &金额+税率/%?# "} {
		t.Run(keyword, func(t *testing.T) {
			t.Parallel()
			payload := []byte(`{"code":0,"data":{"fields":[{"id":9223372036854775807,"filter_unique_key":"custom_key","search_field":"CONTRACT_FORM_FIELDS_OPTION","field_name":"项目区域","options":[{"label":"华东","value":9007199254740993}],"unknown_metadata":{"supported":true}}],"future_response_key":"preserved"}}`)
			calls := 0
			client := openplatform.New(openplatform.Options{
				HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					if req.Method != http.MethodGet || req.URL.Path != "/open-apis/contract/v1/mcp/contracts/search/filter_fields" {
						t.Fatalf("route = %s %s", req.Method, req.URL.Path)
					}
					wantQuery := url.Values{}
					if keyword != "" {
						wantQuery.Set("keyword", keyword)
					}
					if req.URL.RawQuery != wantQuery.Encode() {
						t.Fatalf("query = %q, want %q", req.URL.RawQuery, wantQuery.Encode())
					}
					if req.ContentLength != 0 || req.Header.Get("Authorization") != "Bearer user-token" {
						t.Fatal("expected a body-free request authenticated as the current user")
					}
					response := jsonResponse(string(payload))
					response.Header.Set("X-Request-Id", "metadata-response")
					return response, nil
				})},
			})
			requestContext, err := client.RequestContext(profileWithUserToken(), config.IdentityUser)
			if err != nil {
				t.Fatal(err)
			}
			requestContext.CommonQuery = url.Values{"user_id_type": {"open_id"}, "lang": {"en-US"}, "keyword": {"wrong keyword"}}
			response, err := contract.NewService(client).ListSearchFilterFields(context.Background(), requestContext, keyword)
			if err != nil {
				t.Fatalf("ListSearchFilterFields() error = %v", err)
			}
			if calls != 1 || response.StatusCode != http.StatusOK || response.Headers.Get("X-Request-Id") != "metadata-response" || response.TraceID == "" {
				t.Fatalf("response metadata lost: calls=%d status=%d headers=%v trace=%q", calls, response.StatusCode, response.Headers, response.TraceID)
			}
			if !bytes.Equal(response.Body, payload) {
				t.Fatalf("response changed: %s", response.Body)
			}
			if !reflect.DeepEqual(requestContext.CommonQuery, url.Values{"user_id_type": {"open_id"}, "lang": {"en-US"}, "keyword": {"wrong keyword"}}) {
				t.Fatal("mutated the caller's request context")
			}
		})
	}
}

func TestListSearchFilterFieldsRejectsAppBeforeNetwork(t *testing.T) {
	t.Parallel()

	client := openplatform.New(openplatform.Options{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("app identity must not send a request")
			return nil, nil
		})},
	})
	requestContext, err := client.RequestContext(profileWithAppToken(), config.IdentityApp)
	if err != nil {
		t.Fatal(err)
	}
	requestContext.PrepareAccessToken = func(context.Context, string) (string, error) {
		t.Fatal("app identity must be rejected before token preparation")
		return "", nil
	}
	_, err = contract.NewService(client).ListSearchFilterFields(context.Background(), requestContext, "项目")
	if err == nil || !strings.Contains(err.Error(), "only supports --as user") {
		t.Fatalf("expected user-only error, got %v", err)
	}
}

func TestListSearchFilterFieldsReusesReadRetry(t *testing.T) {
	t.Parallel()

	calls := 0
	client := openplatform.New(openplatform.Options{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return nil, &net.DNSError{IsTimeout: true, Err: "temporary timeout"}
			}
			return jsonResponse(`{"code":0,"data":{"fields":[]}}`), nil
		})},
	})
	requestContext, err := client.RequestContext(profileWithUserToken(), config.IdentityUser)
	if err != nil {
		t.Fatal(err)
	}
	_, err = contract.NewService(client).ListSearchFilterFields(context.Background(), requestContext, "")
	if err != nil || calls != 2 {
		t.Fatalf("read retry: calls=%d error=%v", calls, err)
	}
}
