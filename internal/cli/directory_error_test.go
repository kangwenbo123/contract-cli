package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/openplatform"
)

func TestDirectoryErrorPreservesReauthorizationGuidance(t *testing.T) {
	cause := &openplatform.TraceError{TraceID: "trace-test", Cause: fmt.Errorf("refresh token: %w", ErrDeviceReauthorizationRequired)}
	err := safeDirectoryRequestError(cause)
	if !errors.Is(err, ErrDeviceReauthorizationRequired) || !strings.Contains(err.Error(), ErrDeviceReauthorizationRequired.Error()) || !strings.Contains(err.Error(), "trace-test") {
		t.Fatalf("reauthorization guidance was lost: %v", err)
	}
	if !strings.Contains(err.Error(), "user confirmation is required") {
		t.Fatal("reauthorization must retain user confirmation guidance")
	}
}
