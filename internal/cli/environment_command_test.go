package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/invocation"
	updatecheck "cn.qfei/contract-cli/internal/update"
)

func TestEnvironmentInspectReleaseVersionRemainsLocal(t *testing.T) {
	for _, cached := range []bool{false, true} {
		name := "missing-cache"
		if cached {
			name = "stale-cache"
		}
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			cachePath := filepath.Join(directory, "update-check.json")
			var originalCache []byte
			if cached {
				if err := updatecheck.SaveCache(cachePath, updatecheck.Cache{
					CheckedAt: fixedCLINow().Add(-25 * time.Hour), Channel: "latest",
					CurrentVersion: "1.8.3", LatestVersion: "1.8.4", UpdateAvailable: true,
				}); err != nil {
					t.Fatal(err)
				}
				var err error
				originalCache, err = os.ReadFile(cachePath)
				if err != nil {
					t.Fatal(err)
				}
			}
			var stdout bytes.Buffer
			requests, inspections := 0, 0
			app := cli.New(cli.Options{
				Stdout: &stdout, Stderr: &bytes.Buffer{},
				Store: config.NewStore(directory), Secrets: config.NewSecretsStore(directory),
				UpdateCurrentVersion: "1.8.3", Now: fixedCLINow,
				LookupEnv: func(string) (string, bool) { return "", false },
				HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					requests++
					return jsonResponse(`{"dist-tags":{"latest":"1.8.4"}}`), nil
				})},
				InspectEnvironment: func(context.Context, int) invocation.Result {
					inspections++
					return sampleEnvironmentReport()
				},
			})
			if err := app.Run(context.Background(), []string{"environment", "inspect", "--output", "json"}); err != nil {
				t.Fatal(err)
			}
			if requests != 0 || inspections != 1 {
				t.Errorf("HTTP requests = %d, inspections = %d; want 0 and 1", requests, inspections)
			}
			cache, err := os.ReadFile(cachePath)
			if cached {
				if err != nil || !bytes.Equal(cache, originalCache) {
					t.Errorf("diagnostic changed existing update cache: %v", err)
				}
			} else if !os.IsNotExist(err) {
				t.Errorf("diagnostic created update cache: %v", err)
			}
			var report invocation.Result
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if report.AgentSourceType != "codex" || report.Confidence != "high" || report.RuleID != "client.codex.signed-bundle" {
				t.Fatalf("diagnostic changed attribution: %+v", report)
			}
		})
	}
}

func TestEnvironmentInspectText(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Store:  config.NewStore(t.TempDir()),
		InspectEnvironment: func(_ context.Context, depth int) invocation.Result {
			if depth != invocation.DefaultMaxDepth {
				t.Fatalf("depth = %d", depth)
			}
			return sampleEnvironmentReport()
		},
	})

	if err := app.Run(context.Background(), []string{"environment", "inspect"}); err != nil {
		t.Fatalf("Run(environment inspect) error = %v", err)
	}
	output := stdout.String()
	for _, want := range []string{
		"Channel: cli",
		"Agent Source: codex",
		"Product: contract",
		"Evidence: macos_code_signature",
		"Confidence: high",
		"Rule: client.codex.signed-bundle",
		"bundle_id=com.openai.codex",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Processes:") {
		t.Fatalf("default output should not include process chain:\n%s", output)
	}
}

func TestEnvironmentInspectJSONCanIncludeProcesses(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout:             stdout,
		Stderr:             &bytes.Buffer{},
		Store:              config.NewStore(t.TempDir()),
		InspectEnvironment: func(context.Context, int) invocation.Result { return sampleEnvironmentReport() },
	})

	err := app.Run(context.Background(), []string{"environment", "inspect", "--output", "json", "--include-processes"})
	if err != nil {
		t.Fatalf("Run(environment inspect) error = %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, `"channel_type": "cli"`) ||
		!strings.Contains(output, `"agent_source_type": "codex"`) ||
		!strings.Contains(output, `"product_code": "contract"`) ||
		!strings.Contains(output, `"processes":`) {
		t.Fatalf("unexpected json output:\n%s", output)
	}
}

func TestEnvironmentInspectRejectsInvalidDepth(t *testing.T) {
	app := cli.New(cli.Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  config.NewStore(t.TempDir()),
	})

	err := app.Run(context.Background(), []string{"environment", "inspect", "--depth", "0"})
	if err == nil || !strings.Contains(err.Error(), "between 1 and 128") {
		t.Fatalf("error = %v", err)
	}
}

func sampleEnvironmentReport() invocation.Result {
	matched := invocation.Process{Depth: 2, PID: 10, PPID: 1, Name: "codex", Executable: "/Applications/ChatGPT.app/Contents/Resources/codex"}
	application := invocation.ApplicationIdentity{
		ProcessDepth: 2, BundlePath: "/Applications/ChatGPT.app", BundleID: "com.openai.codex", TeamID: "2DC432GLL2", Version: "1.0.0", SignatureValid: true,
	}
	return invocation.Result{
		ChannelType:     "cli",
		AgentSourceType: "codex",
		ProductCode:     invocation.ProductCodeContract,
		EvidenceType:    "macos_code_signature",
		Confidence:      "high",
		DetectorVersion: invocation.DetectorVersion,
		Platform:        "darwin",
		RuleID:          "client.codex.signed-bundle",
		Reason:          "matched verified bundle",
		MatchedProcess:  &matched,
		Application:     &application,
		Processes: []invocation.Process{
			{Depth: 0, PID: 30, PPID: 20, Name: "contract-cli"},
			matched,
		},
	}
}
