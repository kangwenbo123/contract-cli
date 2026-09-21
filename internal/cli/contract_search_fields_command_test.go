package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
)

// Representative MCP metadata includes multiple request locations and typed values.
const searchFieldsResponse = `{"code":0,"success":true,"data":{"items":[
{"field_type":"BASIC_FIELD","field_display_name":"合同状态","request_location":"COMBINE_CONDITION","request_paths":["combine_condition.contract_status_in"],"search_value_type":"TEXT","value_scopes":[{"label":"审批中","value":"3"}]},
{"field_type":"CUSTOM_FIELD","field_display_name":"项目区域","search_field":"CONTRACT_FORM_FIELDS_OPTION","filter_unique_key":"custom_test_region","filter_unique_key_required":true,"request_location":"FILTER_UNITS","request_paths":["filter_units"],"search_value_type":"OPTION","value_description":"使用真实选项值","usage_hint":"保留唯一 key","value_scopes":[{"label":"华东","value":"east"},{"label":"数字选项","value":9007199254740993}],"examples":[{"request_fragment":{"filter_units":[{"search_field":"CONTRACT_FORM_FIELDS_OPTION","filter_unique_key":"custom_test_region","search_value":["east"]}]}}],"future_metadata":{"enabled":true}}
],"total_count":2,"has_more":false,"page_token":""}}`

func TestContractSearchFieldsUsesMCPKeywordQueryAndPreservesMetadata(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		args    []string
		keyword string
		raw     bool
	}{
		{name: "Chinese field name", args: []string{"--as", "user", "--keyword", "项目区域"}, keyword: "项目区域"},
		{name: "literal reserved characters", args: []string{"--keyword", "字段 &+%_*?= 空格"}, keyword: "字段 &+%_*?= 空格"},
		{name: "explicit full catalog", args: nil},
		{name: "empty keyword follows MCP", args: []string{"--keyword", ""}},
		{name: "raw metadata", args: []string{"--raw", "--keyword", "项目区域"}, keyword: "项目区域", raw: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := config.NewStore(t.TempDir())
			// This user-only command must not inherit the profile's app default.
			if err := store.UpsertProfile(uploadProfile(config.IdentityApp), true); err != nil {
				t.Fatal(err)
			}
			stdout := &bytes.Buffer{}
			calls := 0
			app := cli.New(cli.Options{Stdout: stdout, Stderr: &bytes.Buffer{}, Store: store,
				HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					if req.Method != http.MethodGet || req.URL.Path != "/open-apis/contract/v1/mcp/contracts/search/filter_fields" {
						t.Fatalf("unexpected MCP route: %s %s", req.Method, req.URL.Path)
					}
					wantQuery := url.Values{}
					if tc.keyword != "" {
						wantQuery.Set("keyword", tc.keyword)
					}
					if req.URL.RawQuery != wantQuery.Encode() {
						t.Fatalf("query = %q, want %q", req.URL.RawQuery, wantQuery.Encode())
					}
					if req.Header.Get("Authorization") != "Bearer user-token" {
						t.Fatal("must use authorized user identity")
					}
					if req.Header.Get("X-MCP-Response-Profile") != "" {
						t.Fatal("must not opt in to an unpublished response profile")
					}
					if req.Body != nil {
						body, err := io.ReadAll(req.Body)
						if err != nil {
							t.Fatal(err)
						}
						if len(body) != 0 {
							t.Fatalf("GET must not carry a body: %s", body)
						}
					}
					return jsonResponse(searchFieldsResponse), nil
				})}})
			args := append([]string{"contract", "search-fields", "--profile", "contract"}, tc.args...)
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("HTTP calls = %d, want 1", calls)
			}
			if tc.raw {
				if stdout.String() != searchFieldsResponse {
					t.Fatalf("raw response changed: %s", stdout)
				}
				return
			}
			decode := func(text string) any {
				decoder := json.NewDecoder(strings.NewReader(text))
				decoder.UseNumber()
				var value any
				if err := decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			if !reflect.DeepEqual(decode(stdout.String()), decode(searchFieldsResponse)) {
				t.Fatalf("metadata was lost or its value types changed: %s", stdout)
			}
		})
	}
}

func TestContractSearchFieldsRejectsUnsupportedInputsBeforeHTTP(t *testing.T) {
	t.Parallel()
	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"app identity", []string{"--as", "app"}, "only supports --as user"},
		{"positional", []string{"项目区域"}, "usage:"},
		{"JSON body", []string{"--data", `{}`}, "does not accept"},
		{"body file", []string{"--input-file", "must-not-read.json"}, "does not accept"},
		{"missing keyword value", []string{"--keyword"}, "--keyword"},
		{"unpublished pagination", []string{"--page-size", "20"}, "--page-size"},
		{"unpublished language", []string{"--lang", "en"}, "--lang"},
		{"caller override", []string{"--user-id", "someone-else"}, "--user-id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := cli.New(cli.Options{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, Store: store,
				HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					t.Fatal("invalid input must not send HTTP")
					return nil, nil
				})}})
			err := app.Run(context.Background(), append([]string{"contract", "search-fields", "--profile", "contract"}, tc.args...))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestContractSearchFieldsReportsBusinessFailureAndKeepsResponse(t *testing.T) {
	t.Parallel()
	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"code":110000,"success":false,"msg":"参数有误","data":null}`,
		`{"code":0,"success":false,"msg":"查询失败"}`,
		`{"success":true,"data":{"items":[]}}`,
	} {
		stdout := &bytes.Buffer{}
		app := cli.New(cli.Options{Stdout: stdout, Stderr: &bytes.Buffer{}, Store: store,
			HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return jsonResponse(body), nil })}})
		err := app.Run(context.Background(), []string{"contract", "search-fields", "--profile", "contract", "--keyword", "项目区域", "--raw"})
		if err == nil {
			t.Fatalf("invalid business envelope must fail: %s", body)
		}
		if stdout.String() != body {
			t.Fatalf("business error response lost: %s", stdout)
		}
	}
}

func TestContractSearchFieldsHelpMatchesPublishedMCPParameters(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{Stdout: stdout, Stderr: &bytes.Buffer{}, Store: config.NewStore(t.TempDir())})
	if err := app.Run(context.Background(), []string{"contract", "search-fields", "--help"}); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"--keyword", "list-contract-search-filter-fields", "GET /open-apis/contract/v1/mcp/contracts/search/filter_fields", "仅支持 user", "字段展示名", "request_location"} {
		if !strings.Contains(stdout.String(), required) {
			t.Errorf("help missing %q: %s", required, stdout)
		}
	}
	for _, unsupported := range []string{"--page-size", "--page-token", "--lang", "--user-id", "--data", "--input-file"} {
		if strings.Contains(stdout.String(), unsupported) {
			t.Errorf("help exposes unsupported flag %q", unsupported)
		}
	}
}
