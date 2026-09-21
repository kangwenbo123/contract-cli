package cli_test

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestContractUserScenariosPreserveRequestedScopeAndFilters(t *testing.T) {
	requests := documentedSearchRequestsAtPath(t,
		filepath.Join("..", "..", "skills", "contract-cli-contract-search", "SKILL.md"))
	if len(requests) != 9 {
		t.Fatalf("main skill scenario examples = %d, want 9", len(requests))
	}
	t.Run("page text across five sources", func(t *testing.T) {
		assertTextSearchRecipe(t, requests[0], []string{"CONTRACT_SCAN_TEXT", "CONTRACT_TEXT_FIELD",
			"CONTRACT_ATTACHMENT_TEXT", "CONTRACT_ARCHIVE_ATTACHMENT", "CONTRACT_CAUSE_FIELD"})
	})
	t.Run("only body and CNY amount", func(t *testing.T) {
		assertTextSearchRecipe(t, requests[1], []string{"CONTRACT_TEXT_FIELD"})
		assertScenarioFilters(t, requests[1], map[string]any{
			"CONTRACT_AMOUNT":   []any{float64(1000), float64(5000)},
			"CONTRACT_CURRENCY": []any{"CNY"},
		})
	})
	t.Run("my contracts and name and approval status", func(t *testing.T) {
		request := requests[2]
		expected := map[string]any{"contract_name": "采购", "contract_status_in": "3"}
		if request.SearchTabCode != 1 || !reflect.DeepEqual(request.CombineCondition, expected) ||
			len(request.ConditionUnits) != 0 {
			t.Errorf("my-contract scope or verified name/status combination lost: %+v", request)
		}
	})
	t.Run("number keyword within user scope", func(t *testing.T) {
		if requests[3].ContractNumber != "HT2026" || requests[3].SearchTabCode != 0 || requests[3].PageSize != 20 {
			t.Errorf("number keyword recipe changed: %+v", requests[3])
		}
	})
	t.Run("submission dates and archived status", func(t *testing.T) {
		expected := map[string]any{"submited_time_start": "2026-09-01 00:00:00",
			"submited_time_end": "2026-09-15 23:59:59", "contract_status_in": "9"}
		if !reflect.DeepEqual(requests[4].CombineCondition, expected) {
			t.Errorf("date role, range or status lost: %+v", requests[4])
		}
	})
	t.Run("applicant name keyword needs no invented user ID", func(t *testing.T) {
		request := requests[5]
		if len(request.ConditionUnits) != 1 {
			t.Fatalf("applicant name requires one keyword condition: %+v", request)
		}
		unit := request.ConditionUnits[0]
		if unit.SearchField != "CONTRACT_SUBMIT_NAME" || unit.SearchValue != "张三" || unit.UnionType != "SHOULD" {
			t.Errorf("applicant name must not become owner, creator or ID: %+v", unit)
		}
		if request.SearchTabCode != 0 || len(request.FilterUnits) != 0 || len(request.CombineCondition) != 0 {
			t.Errorf("name keyword must not require ID filters or my-contract scope: %+v", request)
		}
	})
	t.Run("pagination preserves complete original request", func(t *testing.T) {
		firstPage, nextPage := requests[2], requests[8]
		if firstPage.PageToken != "" || nextPage.PageToken == "" {
			t.Errorf("invalid pagination token placement: first=%q next=%q", firstPage.PageToken, nextPage.PageToken)
		}
		nextPage.PageToken = ""
		if !reflect.DeepEqual(firstPage, nextPage) {
			t.Errorf("next page must preserve scope and all filters: first=%+v next=%+v", firstPage, nextPage)
		}
	})
}

func assertScenarioFilters(t *testing.T, request documentedSearchRequest, expected map[string]any) {
	t.Helper()
	actual := make(map[string]any, len(request.FilterUnits))
	for _, unit := range request.FilterUnits {
		if unit.UnionType != "MUST" {
			t.Errorf("additional filter must remain required: %+v", unit)
		}
		actual[unit.SearchField] = unit.SearchValue
	}
	if len(request.FilterUnits) != len(expected) || !reflect.DeepEqual(actual, expected) {
		t.Errorf("filters = %v, want %v", actual, expected)
	}
}
