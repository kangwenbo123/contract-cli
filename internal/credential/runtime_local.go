package credential

import (
	"os"
	"strings"

	"cn.qfei/contract-cli/internal/invocation"
)

// ResolveDeviceRuntimeWithEvidence uses invocation evidence only to choose a
// credential storage scope. It does not authenticate the invoking application
// or change the OAuth permissions granted by the user.
func ResolveDeviceRuntimeWithEvidence(lookupEnv func(string) (string, bool), report invocation.Result) (DeviceRuntime, error) {
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	if hasCloudRuntimeMarker(lookupEnv) || !hasLocalDesktopEvidence(report) {
		return ResolveDeviceRuntime(lookupEnv)
	}
	sessionID := lookupTrimmedEnv(lookupEnv, envWorkBuddySessionID)
	if sessionID == "" {
		sessionID = lookupTrimmedEnv(lookupEnv, envDoubaoWorkTaskSession)
	}
	resolved := DeviceRuntime{Kind: DeviceRuntimeLocalUser, SessionID: sessionID}
	if sessionID != "" {
		resolved.SessionNamespace = deviceSessionNamespace(sessionID)
	}
	return resolved, nil
}

func hasCloudRuntimeMarker(lookupEnv func(string) (string, bool)) bool {
	for _, name := range []string{
		envSkillSessionWorkspace, "DOUBAO_SANDBOX_TYPE", "CODEBUDDY_AGENTOS_SESSION_ID", "AGENTOS_RUNTIME_ID",
	} {
		if lookupTrimmedEnv(lookupEnv, name) != "" {
			return true
		}
	}
	return lookupTrimmedEnv(lookupEnv, "CLIENT_INFO_PLATFORM") == "web_agents" ||
		lookupTrimmedEnv(lookupEnv, "CLIENT_INFO_IDE_TYPE") == "WorkBuddy_Web" ||
		lookupTrimmedEnv(lookupEnv, "CODEBUDDY_SESSION_BIZ_SOURCE") == "agent-server"
}

func hasLocalDesktopEvidence(report invocation.Result) bool {
	if report.Platform != "darwin" && report.Platform != "windows" {
		return false
	}
	for _, warning := range report.Warnings {
		if strings.HasPrefix(warning, "conflicting ") {
			return false
		}
	}
	for _, rule := range localDesktopRules(report.AgentSourceType) {
		if matchesLocalDesktopRule(report, rule) {
			return true
		}
	}
	return matchesKnownLocalRuntime(report)
}

func localDesktopRules(source string) []string {
	switch source {
	case "doubao":
		return []string{"client.doubao"}
	case "doubaoWork":
		return []string{"client.doubao_work"}
	case "workbuddy":
		return []string{"client.workbuddy", "client.workbuddy_international"}
	case "codex":
		return []string{"client.codex"}
	}
	return nil
}

func matchesLocalDesktopRule(report invocation.Result, rule string) bool {
	switch report.EvidenceType {
	case "macos_code_signature":
		return report.Platform == "darwin" && report.RuleID == rule+".signed-bundle"
	case "windows_authenticode":
		return report.Platform == "windows" && report.RuleID == rule+".authenticode"
	case "windows_package_identity":
		return report.Platform == "windows" && report.RuleID == rule+".package-family"
	case "process_executable_path":
		return report.RuleID == rule+".executable-path"
	}
	return false
}

func matchesKnownLocalRuntime(report invocation.Result) bool {
	selector := report.Platform + ":" + report.AgentSourceType + ":" + report.EvidenceType + ":" + report.RuleID
	switch selector {
	case "darwin:doubaoWork:macos_signed_host_runtime:client.doubao_work.feishu-runtime",
		"windows:doubaoWork:windows_signed_host_runtime:client.doubao_work.feishu-runtime",
		"darwin:doubaoWork:host_path_runtime_environment:client.doubao_work.feishu-runtime-path",
		"windows:doubaoWork:host_path_runtime_environment:client.doubao_work.feishu-runtime-path",
		"darwin:workbuddy:macos_runtime_environment:client.workbuddy.brokered-runtime",
		"windows:workbuddy:windows_runtime_environment:client.workbuddy.environment",
		"windows:doubao:windows_runtime_environment:client.doubao.environment",
		"windows:doubaoWork:windows_runtime_environment:client.doubao_work.environment",
		"windows:workbuddy:process_executable_path:client.workbuddy.bundled-git-ancestor":
		return true
	}
	return false
}
