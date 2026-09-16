package invocation

import (
	"net/http"
	"testing"
)

func TestApplyHeadersOverwritesCallerValuesAndRemovesEmptyRule(t *testing.T) {
	headers := http.Header{
		HeaderChannelType:     {"spoofed-channel"},
		HeaderAgentSourceType: {"spoofed-agent"},
		HeaderProductCode:     {"spoofed-product"},
		legacyRequestSource:   {"legacy"},
		HeaderRuleID:          {"stale-rule"},
	}
	ApplyHeaders(headers, Result{
		ChannelType:     "cli",
		AgentSourceType: "workbuddy",
		ProductCode:     ProductCodeContract,
		EvidenceType:    "process_executable_path",
		Confidence:      "medium",
		DetectorVersion: DetectorVersion,
	})

	if got := headers.Get(HeaderChannelType); got != "cli" {
		t.Fatalf("channel header = %q", got)
	}
	if got := headers.Get(HeaderAgentSourceType); got != "workbuddy" {
		t.Fatalf("agent source header = %q", got)
	}
	if got := headers.Get(HeaderProductCode); got != ProductCodeContract {
		t.Fatalf("product code header = %q", got)
	}
	if got := headers.Get(legacyRequestSource); got != "" {
		t.Fatalf("legacy request source header = %q", got)
	}
	if got := headers.Get(HeaderRuleID); got != "" {
		t.Fatalf("rule header = %q", got)
	}
}
