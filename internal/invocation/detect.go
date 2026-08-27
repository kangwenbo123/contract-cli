package invocation

import (
	"fmt"
	"path/filepath"
	"strings"
)

type applicationRule struct {
	ID                        string
	Channel                   string
	BundleID                  string
	TeamID                    string
	WindowsPackageFamilyNames []string
	WindowsCertificateSHA256  []string
	ExecutableMarkers         []string
	ProcessNames              []string
}

var applicationRules = []applicationRule{
	{
		ID:       "client.doubao",
		Channel:  "doubao",
		BundleID: "com.bot.pc.doubao",
		TeamID:   "96L78H6LMH",
		// Leaf certificate from the official Doubao Windows installer 1.81.6.
		// Treat rotation as an explicit registry update instead of trusting a same-name executable.
		WindowsCertificateSHA256: []string{
			"f05e610036eddb254d1d9344b824ac969cce3e5b5658485cd2167767623262dc",
		},
		ExecutableMarkers: []string{"/applications/doubao.app/", `\doubao\`, `\doubao.exe`},
		ProcessNames:      []string{"doubao", "doubao.exe"},
	},
	{
		ID:       "client.workbuddy",
		Channel:  "workbuddy",
		BundleID: "com.workbuddy.workbuddy",
		TeamID:   "FN2V63AD2J",
		// Leaf certificate from the official WorkBuddy Windows installer 5.3.14.36279234.
		WindowsCertificateSHA256: []string{
			"a7d0aff6774068a4f37485b7e61cbf9d31b65190aaedfe8cb79ebd3c65cbce76",
		},
		ExecutableMarkers: []string{"/applications/workbuddy.app/", `\workbuddy\`, `\codebuddy\`, `\workbuddy.exe`, `\codebuddy.exe`},
		ProcessNames:      []string{"workbuddy", "workbuddy.exe", "codebuddy", "codebuddy.exe"},
	},
	{
		ID:       "client.codex",
		Channel:  "codex",
		BundleID: "com.openai.codex",
		TeamID:   "2DC432GLL2",
		// Package identity published by OpenAI for Codex in Microsoft Store product 9PLM9XGG6VKS.
		WindowsPackageFamilyNames: []string{
			"OpenAI.Codex_2p2nqsd0c76g0",
		},
		ExecutableMarkers: []string{"/applications/chatgpt.app/", "/applications/codex.app/", `\chatgpt\`, `\codex\`, `\chatgpt.exe`, `\codex.exe`},
		ProcessNames:      []string{"chatgpt", "chatgpt.exe", "codex", "codex.exe"},
	},
}

func Analyze(chain []Process, identities []ApplicationIdentity) Result {
	result := Result{
		RequestSourceType: "cli",
		ChannelType:       "unknown",
		EvidenceType:      "none",
		Confidence:        "unknown",
		DetectorVersion:   DetectorVersion,
		Reason:            "no registered client matched the process ancestry",
	}

	for _, identity := range identities {
		if !identity.SignatureValid {
			continue
		}
		for _, rule := range applicationRules {
			if strings.EqualFold(identity.BundleID, rule.BundleID) && strings.EqualFold(identity.TeamID, rule.TeamID) {
				result.ChannelType = rule.Channel
				result.EvidenceType = "macos_code_signature"
				result.Confidence = "high"
				result.RuleID = rule.ID + ".signed-bundle"
				result.Reason = fmt.Sprintf("matched verified bundle id %s and team id %s", identity.BundleID, identity.TeamID)
				result.Application = copyApplicationIdentity(identity)
				result.MatchedProcess = processAtDepth(chain, identity.ProcessDepth)
				return result
			}
		}
	}

	for _, identity := range identities {
		if identity.PackageFamilyName == "" {
			continue
		}
		for _, rule := range applicationRules {
			if matchesString(identity.PackageFamilyName, rule.WindowsPackageFamilyNames) {
				result.ChannelType = rule.Channel
				result.EvidenceType = "windows_package_identity"
				result.Confidence = "high"
				result.RuleID = rule.ID + ".package-family"
				result.Reason = fmt.Sprintf("matched Windows package family name %s", identity.PackageFamilyName)
				result.Application = copyApplicationIdentity(identity)
				result.MatchedProcess = processAtDepth(chain, identity.ProcessDepth)
				return result
			}
		}
	}

	for _, identity := range identities {
		if !identity.SignatureValid || identity.CertificateSHA256 == "" {
			continue
		}
		executable := normalizeExecutable(identity.ExecutablePath)
		for _, rule := range applicationRules {
			if matchesExecutableMarker(executable, rule.ExecutableMarkers) &&
				matchesCertificateSHA256(identity.CertificateSHA256, rule.WindowsCertificateSHA256) {
				result.ChannelType = rule.Channel
				result.EvidenceType = "windows_authenticode"
				result.Confidence = "high"
				result.RuleID = rule.ID + ".authenticode"
				result.Reason = fmt.Sprintf("matched a trusted Authenticode certificate for %s", rule.Channel)
				result.Application = copyApplicationIdentity(identity)
				result.MatchedProcess = processAtDepth(chain, identity.ProcessDepth)
				return result
			}
		}
	}

	for _, identity := range identities {
		if identity.BundlePath == "" {
			continue
		}
		bundlePath := normalizeExecutable(identity.BundlePath)
		for _, rule := range applicationRules {
			if !matchesExecutableMarker(bundlePath, rule.ExecutableMarkers) {
				continue
			}
			result.EvidenceType = "macos_code_signature_mismatch"
			result.RuleID = rule.ID + ".signed-bundle-mismatch"
			if identity.SignatureValid {
				result.Reason = fmt.Sprintf("application path matched %s but its verified bundle identity was unavailable or did not match the registered bundle and team ids", rule.Channel)
			} else {
				result.Reason = fmt.Sprintf("application path matched %s but its code signature could not be verified", rule.Channel)
			}
			result.Application = copyApplicationIdentity(identity)
			result.MatchedProcess = processAtDepth(chain, identity.ProcessDepth)
			return result
		}
	}

	for _, identity := range identities {
		if identity.ExecutablePath == "" {
			continue
		}
		executable := normalizeExecutable(identity.ExecutablePath)
		for _, rule := range applicationRules {
			if !matchesExecutableMarker(executable, rule.ExecutableMarkers) && !matchesProcessNameAtDepth(chain, identity.ProcessDepth, rule.ProcessNames) {
				continue
			}

			result.Application = copyApplicationIdentity(identity)
			result.MatchedProcess = processAtDepth(chain, identity.ProcessDepth)
			result.RuleID = rule.ID + ".windows-identity-mismatch"
			switch {
			case identity.PackageFamilyName != "":
				result.EvidenceType = "windows_package_identity_mismatch"
				result.Reason = fmt.Sprintf("application path matched %s but package family name %s is not registered", rule.Channel, identity.PackageFamilyName)
			case !identity.SignatureValid:
				result.EvidenceType = "windows_authenticode_mismatch"
				result.Reason = fmt.Sprintf("application path matched %s but its Authenticode signature could not be verified", rule.Channel)
			default:
				result.EvidenceType = "windows_authenticode_mismatch"
				result.Reason = fmt.Sprintf("application path matched %s but its verified signing certificate is not registered", rule.Channel)
			}
			return result
		}
	}

	for index := 1; index < len(chain); index++ {
		current := chain[index]
		executable := normalizeExecutable(current.Executable)
		for _, rule := range applicationRules {
			if matchesExecutableMarker(executable, rule.ExecutableMarkers) {
				result.ChannelType = rule.Channel
				result.EvidenceType = "process_executable_path"
				result.Confidence = "medium"
				result.RuleID = rule.ID + ".executable-path"
				result.Reason = "ancestor executable matched a registered path marker"
				result.MatchedProcess = copyProcess(current)
				return result
			}
		}
	}

	for index := 1; index < len(chain); index++ {
		current := chain[index]
		name := strings.ToLower(strings.TrimSpace(current.Name))
		if name == "" {
			name = strings.ToLower(filepath.Base(current.Executable))
		}
		for _, rule := range applicationRules {
			for _, registered := range rule.ProcessNames {
				if name == strings.ToLower(registered) {
					result.ChannelType = rule.Channel
					result.EvidenceType = "process_name"
					result.Confidence = "low"
					result.RuleID = rule.ID + ".process-name"
					result.Reason = fmt.Sprintf("ancestor process name matched %q", registered)
					result.MatchedProcess = copyProcess(current)
					return result
				}
			}
		}
	}

	return result
}

func matchesExecutableMarker(executable string, markers []string) bool {
	candidateWithSeparator := strings.TrimSuffix(executable, "/") + "/"
	for _, marker := range markers {
		normalizedMarker := normalizeExecutable(marker)
		if strings.Contains(executable, normalizedMarker) || strings.Contains(candidateWithSeparator, normalizedMarker) {
			return true
		}
	}
	return false
}

func matchesString(value string, registered []string) bool {
	for _, candidate := range registered {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func matchesCertificateSHA256(value string, registered []string) bool {
	return matchesString(normalizeCertificateSHA256(value), registered)
}

func normalizeCertificateSHA256(value string) string {
	replacer := strings.NewReplacer(":", "", " ", "", "-", "")
	return strings.ToLower(replacer.Replace(strings.TrimSpace(value)))
}

func matchesProcessNameAtDepth(chain []Process, depth int, names []string) bool {
	current := processAtDepth(chain, depth)
	if current == nil {
		return false
	}
	name := strings.TrimSpace(current.Name)
	if name == "" {
		name = filepath.Base(current.Executable)
	}
	return matchesString(name, names)
}

func matchingWindowsRule(current Process) *applicationRule {
	executable := normalizeExecutable(current.Executable)
	name := strings.TrimSpace(current.Name)
	if name == "" {
		name = filepath.Base(current.Executable)
	}
	for index := range applicationRules {
		rule := &applicationRules[index]
		if matchesExecutableMarker(executable, rule.ExecutableMarkers) || matchesString(name, rule.ProcessNames) {
			return rule
		}
	}
	return nil
}

func normalizeExecutable(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), `\`, "/"))
}

func processAtDepth(chain []Process, depth int) *Process {
	for _, current := range chain {
		if current.Depth == depth {
			return copyProcess(current)
		}
	}
	return nil
}

func copyProcess(value Process) *Process {
	copy := value
	return &copy
}

func copyApplicationIdentity(value ApplicationIdentity) *ApplicationIdentity {
	copy := value
	return &copy
}
