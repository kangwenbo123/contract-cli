package cli_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
)

func TestContractSearchUserIDTypeFollowsResolvedIdentity(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name            string
		defaultIdentity config.IdentityKind
		args            []string
		wantIdentity    config.IdentityKind
		wantIDType      string
		wantError       bool
	}{
		{"explicit user accepts user_id", config.IdentityApp, []string{"--as", "user", "--user-id-type", "user_id"}, config.IdentityUser, "user_id", false},
		{"explicit user rejects union_id", config.IdentityApp, []string{"--as", "user", "--user-id-type", "union_id"}, config.IdentityUser, "", true},
		{"explicit user rejects employee_id", config.IdentityUser, []string{"--as", "user", "--user-id-type", "employee_id"}, config.IdentityUser, "", true},
		{"default user rejects open_id", config.IdentityUser, []string{"--user-id-type", "open_id"}, config.IdentityUser, "", true},
		{"default user rejects empty type", config.IdentityUser, []string{"--user-id-type", ""}, config.IdentityUser, "", true},
		{"default user accepts omitted type", config.IdentityUser, nil, config.IdentityUser, "user_id", false},
		{"default user accepts user_id", config.IdentityUser, []string{"--user-id-type", "user_id"}, config.IdentityUser, "user_id", false},
		{"explicit app preserves union_id", config.IdentityUser, []string{"--as", "app", "--user-id-type", "union_id"}, config.IdentityApp, "union_id", false},
		{"default app preserves union_id", config.IdentityApp, []string{"--user-id-type", "union_id"}, config.IdentityApp, "union_id", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := config.NewStore(t.TempDir())
			if err := store.UpsertProfile(uploadProfile(tc.defaultIdentity), true); err != nil {
				t.Fatal(err)
			}
			calls := 0
			app := cli.New(cli.Options{Store: store, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
				HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					wantPath := "/open-apis/contract/v1/contracts/search"
					if tc.wantIdentity == config.IdentityUser {
						wantPath = "/open-apis/contract/v1/mcp/contracts/search"
					}
					if req.URL.Path != wantPath {
						t.Errorf("path = %q, want %q", req.URL.Path, wantPath)
					}
					if !tc.wantError && req.URL.Query().Get("user_id_type") != tc.wantIDType {
						t.Errorf("user_id_type = %q, want %q", req.URL.Query().Get("user_id_type"), tc.wantIDType)
					}
					return jsonResponse(`{"code":0,"success":true,"data":{"items":[],"has_more":false}}`), nil
				})}})
			args := append([]string{"contract", "search", "--profile", "contract", "--data", `{"page_size":1}`}, tc.args...)
			err := app.Run(context.Background(), args)
			if tc.wantError {
				if err == nil || !strings.Contains(err.Error(), "user identity requires --user-id-type user_id") {
					t.Errorf("error = %v, want fixed MCP user_id guidance", err)
				}
				if calls != 0 {
					t.Errorf("rejected ID type sent %d HTTP requests", calls)
				}
				return
			}
			if err != nil || calls != 1 {
				t.Errorf("error = %v, HTTP requests = %d, want one successful request", err, calls)
			}
		})
	}
}

func TestContractSearchHelpExplainsFixedUserIDTypeAndFieldDiscovery(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{Store: config.NewStore(t.TempDir()), Stdout: stdout, Stderr: &bytes.Buffer{}})
	if err := app.Run(context.Background(), []string{"contract", "search", "--help"}); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"user 固定 user_id", "app", "contract search-fields", "业务 code"} {
		if !strings.Contains(stdout.String(), required) {
			t.Errorf("search help missing %q", required)
		}
	}
	if strings.Contains(stdout.String(), "不传默认 user_id，传了则覆盖默认值") {
		t.Error("search help incorrectly promises an override for user identity")
	}
	stdout.Reset()
	if err := app.Run(context.Background(), []string{"contract", "get", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "不传默认 user_id，传了则覆盖默认值") {
		t.Error("other commands must retain their existing ID type help")
	}
}
