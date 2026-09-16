//go:build test_e2e

package cli

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/credential"
)

func TestTestE2EEnvironmentPreset(t *testing.T) {
	preset, err := resolveEnvironment("test")
	if err != nil {
		t.Fatal(err)
	}
	if preset.OpenPlatformBaseURL != "https://test-open.qtech.cn" ||
		preset.AuthorizationServerMetadataURL != "https://test-myaccount.qtech.cn/.well-known/oauth-authorization-server/contract" ||
		preset.Resource != "https://test-open.qtech.cn" {
		t.Fatalf("unexpected Test preset: %+v", preset)
	}
	if _, err := resolveEnvironment("prod"); err == nil {
		t.Fatal("Test E2E build must reject prod environment")
	}
}

func TestTestE2EProfileValidationRejectsNonTestEndpoints(t *testing.T) {
	profile := testE2EProfileFixture()
	if err := validateProductionProfile(profile); err != nil {
		t.Fatalf("valid Test profile rejected: %v", err)
	}

	profile.OpenPlatformBaseURL = "https://open.qfei.cn"
	if err := validateProductionProfile(profile); err == nil {
		t.Fatal("Test E2E build must reject a production profile")
	}
}

func TestTestE2EPendingAuthorizationAllowsOnlyTestAccountHost(t *testing.T) {
	testPending := &credential.PendingTransaction{
		TokenEndpoint:           "https://test-myaccount.qtech.cn/api/public/oauth/token/contract",
		VerificationURIComplete: "https://test-myaccount.qtech.cn/device?code=redacted",
	}
	if err := validateProductionPendingTransaction("contract-test", testPending); err != nil {
		t.Fatalf("valid Test pending authorization rejected: %v", err)
	}

	productionPending := *testPending
	productionPending.TokenEndpoint = "https://myaccount.qfei.cn/api/public/oauth/token/contract"
	if err := validateProductionPendingTransaction("contract-test", &productionPending); err == nil {
		t.Fatal("Test E2E build must reject production pending authorization")
	}
}

func TestTestE2ENetworkGuardAllowsOnlyTestHosts(t *testing.T) {
	transport := productionGuardTransport{
		next: testE2ERoundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("{}"))}, nil
		}),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, rawURL := range []string{"https://test-open.qtech.cn/open-apis/ping", "https://test-myaccount.qtech.cn/ping"} {
		request, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := transport.RoundTrip(request); err != nil {
			t.Fatalf("allowed Test URL %s rejected: %v", rawURL, err)
		}
	}

	for _, rawURL := range []string{
		"https://open.qfei.cn/open-apis/ping",
		"https://myaccount.qfei.cn/ping",
		"https://dev-open.qtech.cn/open-apis/ping",
		"https://example.com/ping",
	} {
		request, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := transport.RoundTrip(request); err == nil {
			t.Fatalf("non-Test URL %s was allowed", rawURL)
		}
	}
}

type testE2ERoundTripFunc func(*http.Request) (*http.Response, error)

func (function testE2ERoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func testE2EProfileFixture() config.Profile {
	return config.Profile{
		Name:                           "contract-test",
		Environment:                    "test",
		OpenPlatformBaseURL:            "https://test-open.qtech.cn",
		AppTokenEndpoint:               "https://test-open.qtech.cn/open-apis/auth/v3/tenant_access_token/internal",
		AuthorizationServerMetadataURL: "https://test-myaccount.qtech.cn/.well-known/oauth-authorization-server/contract",
		Resource:                       "https://test-open.qtech.cn",
		Identities: config.Identities{User: config.UserIdentity{
			AuthorizationEndpoint:       "https://test-myaccount.qtech.cn/api/public/oauth/authorize/contract",
			DeviceAuthorizationEndpoint: "https://test-myaccount.qtech.cn/api/public/oauth/device-authorization/contract",
			TokenEndpoint:               "https://test-myaccount.qtech.cn/api/public/oauth/token/contract",
			RevocationEndpoint:          "https://test-myaccount.qtech.cn/api/public/oauth/revoke/contract",
			RegistrationEndpoint:        "https://test-myaccount.qtech.cn/api/public/oauth/register/contract",
		}},
	}
}
