package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
)

func TestContractSearchPreservesJSONNumbersOnWire(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value string
	}{
		{"integer above float64 precision", `[9007199254740993]`},
		{"high precision decimal", `[0.123456789012345678901234567890,5000.000000000000000001]`},
		{"nested arrays and mixed types", `[[9007199254740993,-9007199254740993],{"amount":[0.123456789012345678901234567890,null],"id":"9007199254740993","enabled":true}]`},
		{"scientific notation", `[9.007199254740993e+15,1.234567890123456789E-25,1e400]`},
		{"integer beyond float64 range", `[` + strings.Repeat("9", 400) + `]`},
	}
	for _, tc := range cases {
		for _, inputFlag := range []string{"--data", "--input-file"} {
			t.Run(tc.name+"/"+inputFlag, func(t *testing.T) {
				body := `{"search_tab_code":0,"filter_units":[{"search_field":"CONTRACT_FORM_FIELDS_OPTION","filter_unique_key":"fixture_numeric_field","search_value":` + tc.value + `,"union_type":"MUST"}]}`
				calls := 0
				app := newJSONNumberInputTestApp(t, func(req *http.Request) (*http.Response, error) {
					calls++
					if req.Method != http.MethodPost || req.URL.Path != "/open-apis/contract/v1/mcp/contracts/search" {
						t.Fatalf("unexpected route: %s %s", req.Method, req.URL.Path)
					}
					assertJSONNumberRequestBody(t, req, body)
					return jsonResponse(`{"code":0,"success":true,"data":{"items":[],"has_more":false,"page_token":""}}`), nil
				})
				args := []string{"contract", "search", "--profile", "contract", "--as", "user"}
				args = append(args, jsonNumberInputArgs(t, inputFlag, body)...)
				if err := app.Run(context.Background(), args); err != nil {
					t.Fatalf("valid JSON number input failed: %v", err)
				}
				if calls != 1 {
					t.Fatalf("HTTP calls = %d, want 1", calls)
				}
			})
		}
	}
}

func TestJSONObjectInputRejectsExtraDocumentsAndInvalidNumbersBeforeHTTP(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body string
	}{
		{"second object", `{"amount":9007199254740993} {"amount":2}`},
		{"second number", `{"amount":1} 2`},
		{"second null", `{"amount":1} null`},
		{"trailing garbage", `{"amount":9007199254740993} trailing`},
		{"leading zero", `{"amount":01}`},
		{"NaN", `{"amount":NaN}`},
		{"infinity", `{"amount":Infinity}`},
		{"incomplete fraction", `{"amount":1.}`},
		{"incomplete exponent", `{"amount":1e}`},
		{"array instead of object", `[9007199254740993]`},
		{"number instead of object", `9007199254740993`},
		{"null instead of object", `null`},
		{"whitespace only", " \n\t"},
	}
	for _, tc := range cases {
		for _, inputFlag := range []string{"--data", "--input-file"} {
			t.Run(tc.name+"/"+inputFlag, func(t *testing.T) {
				calls := 0
				app := newJSONNumberInputTestApp(t, func(*http.Request) (*http.Response, error) {
					calls++
					return jsonResponse(`{"code":0,"success":true,"data":{"items":[]}}`), nil
				})
				args := []string{"contract", "search", "--profile", "contract", "--as", "user"}
				args = append(args, jsonNumberInputArgs(t, inputFlag, tc.body)...)
				err := app.Run(context.Background(), args)
				if err == nil {
					t.Fatalf("invalid JSON object was accepted: %q", tc.body)
				}
				if !strings.Contains(err.Error(), "json") && !strings.Contains(err.Error(), "JSON object") {
					t.Fatalf("expected JSON input error, got %v", err)
				}
				if calls != 0 {
					t.Fatalf("invalid input sent %d HTTP requests", calls)
				}
			})
		}
	}
}

func TestJSONObjectInputSharedCommandsPreserveNumbersAndFlagOverrides(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		args     []string
		input    string
		wantPath string
		wantBody string
	}{
		{
			name:     "user search overrides only explicit flags",
			args:     []string{"contract", "search", "--as", "user", "--contract-number", "CLI-001", "--page-size", "20", "--page-token", "next-page"},
			input:    `{"contract_number":"BODY-001","page_size":5,"page_token":"old-page","search_tab_code":1,"filter_units":[{"search_field":"CONTRACT_AMOUNT","search_value":[9007199254740993,null],"union_type":"MUST"}]}`,
			wantPath: "/open-apis/contract/v1/mcp/contracts/search",
			wantBody: `{"contract_number":"CLI-001","page_size":20,"page_token":"next-page","search_tab_code":1,"filter_units":[{"search_field":"CONTRACT_AMOUNT","search_value":[9007199254740993,null],"union_type":"MUST"}]}`,
		},
		{
			name:     "app search retains legacy route",
			args:     []string{"contract", "search", "--as", "app"},
			input:    `{"combine_condition":{"contract_amount":9007199254740993},"page_size":20}`,
			wantPath: "/open-apis/contract/v1/contracts/search",
		},
		{
			name:     "user contract create",
			args:     []string{"contract", "create", "--as", "user"},
			input:    `{"title":"numeric fixture","contract_amount":0.123456789012345678901234567890}`,
			wantPath: "/open-apis/contract/v1/mcp/contracts",
		},
		{
			name:     "app contract create",
			args:     []string{"contract", "create", "--as", "app"},
			input:    `{"title":"numeric fixture","contract_amount":9007199254740993}`,
			wantPath: "/open-apis/contract/v1/contracts",
		},
		{
			name:     "user template instantiate",
			args:     []string{"contract", "template", "instantiate", "--as", "user"},
			input:    `{"template_number":"TMP001","template_data":{"amount":1.234567890123456789e+25}}`,
			wantPath: "/open-apis/contract/v1/mcp/template_instances",
		},
		{
			name:     "app template instantiate",
			args:     []string{"contract", "template", "instantiate", "--as", "app"},
			input:    `{"template_number":"TMP001","template_data":{"amount":9007199254740993}}`,
			wantPath: "/open-apis/contract/v1/template_instances",
		},
		{
			name:     "approval task list keeps body numeric fields",
			args:     []string{"contract", "approval", "task", "list", "--as", "user"},
			input:    `{"query":"采购","task_type_code":0,"page_index":2,"page_size":20}`,
			wantPath: "/open-apis/contract/v1/mcp/tasks",
		},
		{
			name:     "approval task list overrides body numeric fields",
			args:     []string{"contract", "approval", "task", "list", "--as", "user", "--task-type", "done", "--page-index", "3", "--page-size", "50"},
			input:    `{"query":"采购","task_type_code":0,"page_index":2,"page_size":20}`,
			wantPath: "/open-apis/contract/v1/mcp/tasks",
			wantBody: `{"query":"采购","task_type_code":1,"page_index":3,"page_size":50}`,
		},
		{
			name:     "trailing whitespace remains valid",
			args:     []string{"contract", "search", "--as", "user"},
			input:    "{\"page_size\":20} \n\t\r",
			wantPath: "/open-apis/contract/v1/mcp/contracts/search",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantBody := tc.wantBody
			if wantBody == "" {
				wantBody = tc.input
			}
			calls := 0
			app := newJSONNumberInputTestApp(t, func(req *http.Request) (*http.Response, error) {
				calls++
				if req.Method != http.MethodPost || req.URL.Path != tc.wantPath {
					t.Fatalf("unexpected route: %s %s, want POST %s", req.Method, req.URL.Path, tc.wantPath)
				}
				assertJSONNumberRequestBody(t, req, wantBody)
				return jsonResponse(`{"code":0,"success":true,"data":{"items":[],"has_more":false}}`), nil
			})
			args := append(append([]string{}, tc.args...), "--profile", "contract", "--data", tc.input)
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("HTTP calls = %d, want 1", calls)
			}
		})
	}
}

func newJSONNumberInputTestApp(t *testing.T, transport roundTripFunc) *cli.App {
	t.Helper()
	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
		t.Fatal(err)
	}
	return cli.New(cli.Options{
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		Store:      store,
		HTTPClient: &http.Client{Transport: transport},
	})
}

func jsonNumberInputArgs(t *testing.T, flag, body string) []string {
	t.Helper()
	if flag == "--data" {
		return []string{flag, body}
	}
	path := filepath.Join(t.TempDir(), "numeric-input.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return []string{flag, path}
}

func assertJSONNumberRequestBody(t *testing.T, req *http.Request, want string) {
	t.Helper()
	got, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	decode := func(body []byte) any {
		t.Helper()
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			t.Fatalf("invalid request JSON %q: %v", body, err)
		}
		return value
	}
	if !reflect.DeepEqual(decode(got), decode([]byte(want))) {
		t.Errorf("request JSON values changed:\n got: %s\nwant: %s", got, want)
	}
}
