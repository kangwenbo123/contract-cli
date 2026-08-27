package cli

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/invocation"
	"cn.qfei/contract-cli/internal/openplatform"
)

func TestBeforeOpenPlatformRequestDetectsAgainForEveryRequest(t *testing.T) {
	reports := []invocation.Result{
		{
			RequestSourceType: "cli",
			ChannelType:       "doubao",
			EvidenceType:      "macos_code_signature",
			Confidence:        "high",
			DetectorVersion:   invocation.DetectorVersion,
			RuleID:            "client.doubao.signed-bundle",
		},
		{
			RequestSourceType: "cli",
			ChannelType:       "workbuddy",
			EvidenceType:      "macos_code_signature",
			Confidence:        "high",
			DetectorVersion:   invocation.DetectorVersion,
			RuleID:            "client.workbuddy.signed-bundle",
		},
	}
	calls := 0
	app := &App{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		inspectEnvironment: func(depth int) invocation.Result {
			if depth != invocation.DefaultMaxDepth {
				t.Fatalf("depth = %d", depth)
			}
			result := reports[calls]
			calls++
			return result
		},
	}

	for index, wantChannel := range []string{"doubao", "workbuddy"} {
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://open.qtech.cn/open-apis/test", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}
		if err := app.beforeOpenPlatformRequest(context.Background(), request); err != nil {
			t.Fatalf("beforeOpenPlatformRequest() error = %v", err)
		}
		if got := request.Header.Get(invocation.HeaderChannelType); got != wantChannel {
			t.Fatalf("request %d channel = %q, want %q", index, got, wantChannel)
		}
	}

	if calls != 2 {
		t.Fatalf("detector calls = %d, want 2", calls)
	}
}

func TestOpenPlatformClientWiresFreshEnvironmentHookIntoEveryAttempt(t *testing.T) {
	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(config.Profile{
		Name:                "contract",
		Environment:         "prod",
		OpenPlatformBaseURL: "https://open.qfei.cn",
		Resource:            "https://open.qfei.cn",
		DefaultIdentity:     config.IdentityUser,
		Identities: config.Identities{
			User: config.UserIdentity{Token: &config.Token{
				AccessToken: "user-token",
				Expiry:      time.Now().Add(time.Hour),
			}},
		},
	}, true); err != nil {
		t.Fatalf("UpsertProfile() error = %v", err)
	}

	channels := []string{"doubao", "workbuddy"}
	detectorCalls := 0
	receivedChannels := make([]string, 0, len(channels))
	app := New(Options{
		Stdout: io.Discard,
		Stderr: io.Discard,
		Store:  store,
		HTTPClient: &http.Client{Transport: roundTripFuncInternal(func(request *http.Request) (*http.Response, error) {
			receivedChannels = append(receivedChannels, request.Header.Get(invocation.HeaderChannelType))
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0}`)),
				Request:    request,
			}, nil
		})},
		InspectEnvironment: func(int) invocation.Result {
			channel := channels[detectorCalls]
			detectorCalls++
			return invocation.Result{
				RequestSourceType: "cli",
				ChannelType:       channel,
				EvidenceType:      "test",
				Confidence:        "high",
				DetectorVersion:   invocation.DetectorVersion,
			}
		},
	})

	client, requestContext, err := app.openPlatformClientAndContext("contract", "user", "/open-apis/contract/v1/mcp/contracts/search", openplatform.IdentityPolicyUserOnly)
	if err != nil {
		t.Fatalf("openPlatformClientAndContext() error = %v", err)
	}
	for range channels {
		if _, err := client.Do(context.Background(), requestContext, openplatform.Request{
			Method: http.MethodGet,
			Path:   "/open-apis/contract/v1/mcp/contracts/search",
		}); err != nil {
			t.Fatalf("Do() error = %v", err)
		}
	}

	if detectorCalls != 2 {
		t.Fatalf("detector calls = %d, want 2", detectorCalls)
	}
	if len(receivedChannels) != 2 || receivedChannels[0] != "doubao" || receivedChannels[1] != "workbuddy" {
		t.Fatalf("received channels = %#v", receivedChannels)
	}
}

type roundTripFuncInternal func(*http.Request) (*http.Response, error)

func (function roundTripFuncInternal) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
