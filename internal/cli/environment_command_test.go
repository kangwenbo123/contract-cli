package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/invocation"
)

func TestEnvironmentInspectText(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := cli.New(cli.Options{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Store:  config.NewStore(t.TempDir()),
		InspectEnvironment: func(depth int) invocation.Result {
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
		"Request Source: cli",
		"Channel: codex",
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
		InspectEnvironment: func(int) invocation.Result { return sampleEnvironmentReport() },
	})

	err := app.Run(context.Background(), []string{"environment", "inspect", "--output", "json", "--include-processes"})
	if err != nil {
		t.Fatalf("Run(environment inspect) error = %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, `"channel_type": "codex"`) || !strings.Contains(output, `"processes":`) {
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
		RequestSourceType: "cli",
		ChannelType:       "codex",
		EvidenceType:      "macos_code_signature",
		Confidence:        "high",
		DetectorVersion:   invocation.DetectorVersion,
		Platform:          "darwin",
		RuleID:            "client.codex.signed-bundle",
		Reason:            "matched verified bundle",
		MatchedProcess:    &matched,
		Application:       &application,
		Processes: []invocation.Process{
			{Depth: 0, PID: 30, PPID: 20, Name: "contract-cli"},
			matched,
		},
	}
}
