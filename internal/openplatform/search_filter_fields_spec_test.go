package openplatform_test

import (
	"net/http"
	"testing"

	"cn.qfei/contract-cli/internal/openplatform"
)

func TestSearchFilterFieldsSpecMatchesMCP(t *testing.T) {
	t.Parallel()

	spec, ok := openplatform.ContractMCPToolSpec("list-contract-search-filter-fields")
	if !ok {
		t.Fatal("MCP search filter fields tool is missing")
	}
	if spec.Method != http.MethodGet || spec.Path != "/open-apis/contract/v1/mcp/contracts/search/filter_fields" {
		t.Fatalf("route = %s %s", spec.Method, spec.Path)
	}
	if spec.IdentityPolicy != openplatform.IdentityPolicyUserOnly {
		t.Fatalf("identity policy = %s, want user only", spec.IdentityPolicy)
	}
	if spec.OperationKind != openplatform.OperationRead {
		t.Fatalf("operation = %s, want read", spec.OperationKind)
	}
	if len(spec.FixedQuery) != 0 {
		t.Fatalf("unexpected fixed query: %v", spec.FixedQuery)
	}
}
