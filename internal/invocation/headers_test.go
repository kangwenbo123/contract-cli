package invocation

import (
	"net/http"
	"testing"
)

func TestApplyHeadersOverwritesCallerValuesAndRemovesEmptyRule(t *testing.T) {
	headers := http.Header{
		HeaderChannelType: {"spoofed"},
		HeaderRuleID:      {"stale-rule"},
	}
	ApplyHeaders(headers, Result{
		RequestSourceType: "cli",
		ChannelType:       "workbuddy",
		EvidenceType:      "process_executable_path",
		Confidence:        "medium",
		DetectorVersion:   DetectorVersion,
	})

	if got := headers.Get(HeaderRequestSourceType); got != "cli" {
		t.Fatalf("request source header = %q", got)
	}
	if got := headers.Get(HeaderChannelType); got != "workbuddy" {
		t.Fatalf("channel header = %q", got)
	}
	if got := headers.Get(HeaderRuleID); got != "" {
		t.Fatalf("rule header = %q", got)
	}
}
