package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/openplatform"
)

const directoryResponse = `{"code":0,"success":true,"data":{"items":[{"user_id":"9007199254740993","employee_name":"同名员工","status":0,"departments":[{"department_id":"od-test","department_name":"测试部门"}]},{"user_id":"user-active","employee_name":"同名员工","status":1,"future_field":9007199254740993}],"has_more":true,"page_token":" next+/%&= "}}`

func newDirectoryTestApp(t *testing.T, body string, inspect func(*http.Request)) (*cli.App, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	store := config.NewStore(t.TempDir())
	if err := store.UpsertProfile(uploadProfile(config.IdentityApp), true); err != nil {
		t.Fatal(err)
	}
	stdout, logs := &bytes.Buffer{}, &bytes.Buffer{}
	app := cli.New(cli.Options{Stdout: stdout, Stderr: logs, Store: store,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			inspect(req)
			return jsonResponse(body), nil
		})}})
	return app, stdout, logs
}

func TestDirectoryListRequestAndResponse(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, resource string
		args           []string
		query          url.Values
	}{
		{"employee name", "employee", []string{"--name", "姓名 &+%_?= 空格"}, url.Values{"name": {"姓名 &+%_?= 空格"}, "page_size": {"10"}, "user_id_type": {"user_id"}}},
		{"employee department next page", "employee", []string{"--department-id", "od-test", "--page-size", "200", "--page-token", " next+/%&= "}, url.Values{"department_collection": {"od-test"}, "page_size": {"200"}, "page_token": {" next+/%&= "}, "user_id_type": {"user_id"}}},
		{"department catalog", "department", nil, url.Values{"page_size": {"10"}}},
		{"department name", "department", []string{"--name", "研发 &+% 部", "--page-size", "1"}, url.Values{"name": {"研发 &+% 部"}, "page_size": {"1"}}},
		{"department ID", "department", []string{"--department-id", "od-test", "--page-token", "next"}, url.Values{"department_collection": {"od-test"}, "page_size": {"10"}, "page_token": {"next"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			app, stdout, logs := newDirectoryTestApp(t, directoryResponse, func(req *http.Request) {
				calls++
				if req.Method != http.MethodGet || req.URL.Path != "/open-apis/contract/v1/mcp/"+tc.resource+"s" {
					t.Fatalf("unexpected route: %s %s", req.Method, req.URL.Path)
				}
				if !reflect.DeepEqual(req.URL.Query(), tc.query) {
					t.Fatalf("query = %v, want %v", req.URL.Query(), tc.query)
				}
				if req.Header.Get("Authorization") != "Bearer user-token" {
					t.Fatal("must use user identity even when profile defaults to app")
				}
				if req.Body != nil {
					body, err := io.ReadAll(req.Body)
					if err != nil || len(body) != 0 {
						t.Fatalf("GET body = %s, err=%v", body, err)
					}
				}
			})
			args := append([]string{tc.resource, "list", "--profile", "contract"}, tc.args...)
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("calls = %d", calls)
			}
			assertDirectoryJSONEqual(t, stdout.String(), directoryResponse)
			for _, sensitive := range []string{"姓名 &+%_?= 空格", "od-test", " next+/%&= ", "user-token"} {
				if strings.Contains(logs.String(), sensitive) {
					t.Errorf("logs expose query/token %q: %s", sensitive, logs)
				}
			}
		})
	}
}

func assertDirectoryJSONEqual(t *testing.T, actual, expected string) {
	t.Helper()
	decode := func(value string) any {
		decoder := json.NewDecoder(strings.NewReader(value))
		decoder.UseNumber()
		var decoded any
		if err := decoder.Decode(&decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	}
	if !reflect.DeepEqual(decode(actual), decode(expected)) {
		t.Fatalf("directory response changed: %s", actual)
	}
}

func TestDirectoryFailureDoesNotExposeQueryInFinalError(t *testing.T) {
	for _, kind := range []string{"network", "http"} {
		t.Run(kind, func(t *testing.T) {
			store := config.NewStore(t.TempDir())
			if err := store.UpsertProfile(uploadProfile(config.IdentityUser), true); err != nil {
				t.Fatal(err)
			}
			logs := &bytes.Buffer{}
			app := cli.New(cli.Options{Stdout: &bytes.Buffer{}, Stderr: logs, Store: store,
				HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					if kind == "network" {
						return nil, errors.New("connection unavailable")
					}
					response := jsonResponse(`{"msg":"姓名敏感 od-sensitive page-sensitive"}`)
					response.StatusCode = http.StatusServiceUnavailable
					return response, nil
				})}})
			err := app.Run(context.Background(), []string{"employee", "list", "--name", "姓名敏感", "--page-token", "page-sensitive"})
			if err == nil {
				t.Fatal("failure must return error")
			}
			printed := logs.String() + err.Error()
			for _, secret := range []string{"姓名敏感", url.QueryEscape("姓名敏感"), "od-sensitive", "page-sensitive"} {
				if strings.Contains(printed, secret) {
					t.Errorf("final error/log exposes %q: %s", secret, printed)
				}
			}
			if !strings.Contains(err.Error(), "trace_id=") {
				t.Errorf("failure lost trace ID: %v", err)
			}
			var traced *openplatform.TraceError
			if !errors.As(err, &traced) {
				t.Fatal("safe error lost the underlying trace")
			}
			if kind == "http" {
				var status *openplatform.HTTPStatusError
				if !errors.As(err, &status) || status.StatusCode != 503 {
					t.Fatal("safe error lost HTTP status")
				}
			}
		})
	}
}

func TestDirectoryListRejectsInvalidInputBeforeHTTP(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"employee"}, {"employee", "get"}, {"department", "get"},
		{"employee", "list"}, {"employee", "list", "--name", ""},
		{"department", "list", "--name", "   "},
		{"employee", "list", "--department-id", ""},
		{"employee", "list", "--name", "test", "--department-id", "od-test"},
		{"department", "list", "--name", "test", "--department-id", "od-test"},
		{"employee", "list", "--name", "test", "--name", "other"},
		{"department", "list", "--department-id", "od-a", "--department-id", "od-b"},
		{"employee", "list", "--name", "test", "--as", "app"},
		{"department", "list", "--as", "app"},
		{"department", "list", "--page-size", "0"}, {"department", "list", "--page-size", "201"},
		{"department", "list", "--page-size", "-1"}, {"department", "list", "--page-size", "1.5"},
		{"department", "list", "--page-size", ""}, {"department", "list", "--page-size"},
		{"department", "list", "--data", "{}"}, {"department", "list", "--input-file", "not-read.json"},
		{"employee", "list", "--name", "test", "--user-id", "someone"},
		{"employee", "list", "--name", "test", "--user-id-type", "union_id"},
		{"department", "list", "--department-id-type", "department_id"},
		{"department", "list", "unexpected"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			app, _, _ := newDirectoryTestApp(t, directoryResponse, func(*http.Request) { t.Fatal("invalid command sent HTTP") })
			if err := app.Run(context.Background(), args); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

func TestDirectoryListBusinessErrorsAndEmptyResults(t *testing.T) {
	t.Parallel()
	for _, resource := range []string{"employee", "department"} {
		for _, body := range []string{`{"code":110000,"success":false,"msg":"invalid"}`, `{"code":0,"success":false}`, `{"success":true}`, `{"code":0,"success":true,"data":{"items":[],"has_more":false,"page_token":""}}`} {
			app, stdout, _ := newDirectoryTestApp(t, body, func(*http.Request) {})
			err := app.Run(context.Background(), []string{resource, "list", "--name", "test", "--raw"})
			wantSuccess := strings.Contains(body, `"items":[]`)
			if (err == nil) != wantSuccess {
				t.Fatalf("body=%s, err=%v", body, err)
			}
			if stdout.String() != body {
				t.Fatalf("raw response changed: %s", stdout)
			}
		}
	}
}

func TestDirectoryLookupIDsFeedContractSearch(t *testing.T) {
	t.Parallel()
	cases := []struct{ resource, field, idKey, id string }{
		{"employee", "CONTRACT_DEMAND_PERSON", "user_id", "9007199254740993"},
		{"department", "CONTRACT_DEMAND_PERSON_DEPARTMENT", "department_id", "od-test"},
	}
	for _, tc := range cases {
		t.Run(tc.resource, func(t *testing.T) {
			calls := 0
			body := fmt.Sprintf(`{"code":0,"success":true,"data":{"items":[{"%s":"%s","status":1}],"has_more":false}}`, tc.idKey, tc.id)
			app, stdout, _ := newDirectoryTestApp(t, body, func(req *http.Request) {
				calls++
				if calls == 1 {
					return
				}
				if req.Method != http.MethodPost || req.URL.Path != "/open-apis/contract/v1/mcp/contracts/search" {
					t.Fatalf("unexpected search route: %s", req.URL)
				}
				var request struct {
					FilterUnits []struct {
						SearchField string   `json:"search_field"`
						SearchValue []string `json:"search_value"`
					} `json:"filter_units"`
				}
				if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if len(request.FilterUnits) != 1 || request.FilterUnits[0].SearchField != tc.field || !reflect.DeepEqual(request.FilterUnits[0].SearchValue, []string{tc.id}) {
					t.Fatalf("ID changed or wrong role: %+v", request)
				}
			})
			if err := app.Run(context.Background(), []string{tc.resource, "list", "--name", "test"}); err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Data struct {
					Items []map[string]any `json:"items"`
				} `json:"data"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			id, ok := envelope.Data.Items[0][tc.idKey].(string)
			if !ok {
				t.Fatal("ID is not a string")
			}
			request := fmt.Sprintf(`{"filter_units":[{"search_field":%q,"search_value":[%q]}]}`, tc.field, id)
			if err := app.Run(context.Background(), []string{"contract", "search", "--as", "user", "--data", request}); err != nil {
				t.Fatal(err)
			}
			if calls != 2 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}

func TestDirectoryCommandsHelpAndSkillsInstall(t *testing.T) {
	app, stdout, _ := newDirectoryTestApp(t, directoryResponse, func(*http.Request) { t.Fatal("help/skills must not send business HTTP") })
	for _, resource := range []string{"employee", "department"} {
		stdout.Reset()
		if err := app.Run(context.Background(), []string{resource, "list", "--help"}); err != nil {
			t.Fatal(err)
		}
		for _, flag := range []string{"--name", "--department-id", "--page-size", "--page-token"} {
			if !strings.Contains(stdout.String(), flag) {
				t.Errorf("missing help flag %s", flag)
			}
		}
	}
	target := t.TempDir()
	if err := app.Run(context.Background(), []string{"skills", "install", "--target", target}); err != nil {
		t.Fatal(err)
	}
	for _, skill := range []string{"contract-cli-employee", "contract-cli-department"} {
		for _, file := range []string{"SKILL.md", "agents/openai.yaml"} {
			if _, err := os.Stat(filepath.Join(target, skill, file)); err != nil {
				t.Fatal(err)
			}
		}
		assertSearchSkillLinksResolve(t, filepath.Join(target, skill))
	}
}
