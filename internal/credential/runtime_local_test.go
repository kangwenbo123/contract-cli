package credential

import (
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/invocation"
)

func TestLocalUserRuntimeRequiresRecognizedDesktopEvidence(t *testing.T) {
	for _, sample := range []struct {
		platform, source, evidence, rule string
	}{
		{"darwin", "doubao", "macos_code_signature", "client.doubao.signed-bundle"},
		{"darwin", "doubaoWork", "process_executable_path", "client.doubao_work.executable-path"},
		{"darwin", "workbuddy", "macos_runtime_environment", "client.workbuddy.brokered-runtime"},
		{"darwin", "doubaoWork", "macos_signed_host_runtime", "client.doubao_work.feishu-runtime"},
		{"darwin", "doubaoWork", "host_path_runtime_environment", "client.doubao_work.feishu-runtime-path"},
		{"darwin", "codex", "macos_code_signature", "client.codex.signed-bundle"},
		{"windows", "doubao", "windows_authenticode", "client.doubao.authenticode"},
		{"windows", "doubaoWork", "windows_runtime_environment", "client.doubao_work.environment"},
		{"windows", "workbuddy", "process_executable_path", "client.workbuddy.bundled-git-ancestor"},
		{"windows", "workbuddy", "windows_authenticode", "client.workbuddy_international.authenticode"},
		{"windows", "doubaoWork", "windows_signed_host_runtime", "client.doubao_work.feishu-runtime"},
		{"windows", "codex", "windows_package_identity", "client.codex.package-family"},
	} {
		t.Run(sample.platform+"/"+sample.rule, func(t *testing.T) {
			report := invocation.Result{Platform: sample.platform, AgentSourceType: sample.source, EvidenceType: sample.evidence, RuleID: sample.rule}
			resolved, err := ResolveDeviceRuntimeWithEvidence(envLookup(map[string]string{"SESSION_ID": "task-a"}), report)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.Kind != DeviceRuntimeLocalUser || resolved.SessionID != "task-a" {
				t.Fatalf("runtime = %+v, want local user preserving task ID", resolved)
			}
		})
	}
}

func TestLocalUserRuntimeSupportsDesktopWithoutSessionID(t *testing.T) {
	resolved, err := ResolveDeviceRuntimeWithEvidence(envLookup(nil), desktopReport())
	if err != nil || resolved.Kind != DeviceRuntimeLocalUser || resolved.SessionID != "" {
		t.Fatalf("runtime = %+v, error = %v", resolved, err)
	}
}

func TestLocalUserRuntimeRejectsWeakConflictingOrUnknownEvidence(t *testing.T) {
	for _, sample := range []struct {
		name   string
		report invocation.Result
	}{
		{"unknown", invocation.Result{Platform: "darwin", AgentSourceType: "unknown", EvidenceType: "none"}},
		{"OS only", invocation.Result{Platform: "darwin"}},
		{"name only", invocation.Result{Platform: "darwin", AgentSourceType: "doubao", EvidenceType: "process_name", RuleID: "client.doubao.process-name"}},
		{"linux", invocation.Result{Platform: "linux", AgentSourceType: "doubao", EvidenceType: "process_executable_path", RuleID: "client.doubao.executable-path"}},
		{"cloud rule", invocation.Result{Platform: "darwin", AgentSourceType: "doubaoWork", EvidenceType: "linux_runtime_environment", RuleID: "client.doubao_work.cloud-runtime"}},
		{"unregistered runtime", invocation.Result{Platform: "darwin", AgentSourceType: "doubao", EvidenceType: "macos_runtime_environment", RuleID: "client.doubao.future-runtime"}},
		{"process conflict", invocation.Result{Platform: "darwin", AgentSourceType: "unknown", EvidenceType: "conflicting_process_evidence"}},
		{"runtime conflict", invocation.Result{Platform: "darwin", AgentSourceType: "doubao", EvidenceType: "macos_code_signature", RuleID: "client.doubao.signed-bundle", Warnings: []string{"conflicting runtime product markers"}}},
		{"source rule mismatch", invocation.Result{Platform: "darwin", AgentSourceType: "workbuddy", EvidenceType: "macos_code_signature", RuleID: "client.doubao.signed-bundle"}},
	} {
		t.Run(sample.name, func(t *testing.T) {
			resolved, err := ResolveDeviceRuntimeWithEvidence(envLookup(map[string]string{"SESSION_ID": "task-a"}), sample.report)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.Kind != DeviceRuntimeDoubaoWorkTask {
				t.Fatalf("weak evidence widened credential scope: %+v", resolved)
			}
		})
	}
}

func TestCloudMarkersOverrideDesktopEvidence(t *testing.T) {
	workspace := t.TempDir()
	for _, marker := range []string{"SKILL_SESSION_WORKSPACE", "DOUBAO_SANDBOX_TYPE", "CODEBUDDY_AGENTOS_SESSION_ID", "AGENTOS_RUNTIME_ID"} {
		t.Run(marker, func(t *testing.T) {
			environment := map[string]string{"SESSION_ID": "task-a", marker: workspace}
			resolved, err := ResolveDeviceRuntimeWithEvidence(envLookup(environment), desktopReport())
			if err != nil {
				t.Fatal(err)
			}
			if resolved.Kind == DeviceRuntimeLocalUser {
				t.Fatalf("cloud marker %s widened credential scope", marker)
			}
			if marker == "SKILL_SESSION_WORKSPACE" && resolved.Kind != DeviceRuntimeDoubaoCloud {
				t.Fatalf("cloud workspace precedence lost: %+v", resolved)
			}
		})
	}
}

func TestUnknownRuntimeDoesNotGainLocalUserStore(t *testing.T) {
	_, err := ResolveDeviceRuntimeWithEvidence(envLookup(nil), invocation.Result{Platform: "darwin", AgentSourceType: "unknown"})
	if err == nil || !strings.Contains(err.Error(), "SESSION_ID") {
		t.Fatalf("error = %v, want existing runtime isolation requirement", err)
	}
}

func desktopReport() invocation.Result {
	return invocation.Result{Platform: "darwin", AgentSourceType: "doubao", EvidenceType: "macos_code_signature", RuleID: "client.doubao.signed-bundle"}
}
