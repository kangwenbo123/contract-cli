package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
	contractskills "cn.qfei/contract-cli/skills"
)

func TestContractSearchSkillInstallsWithResolvableReferences(t *testing.T) {
	target := t.TempDir()
	app := cli.New(cli.Options{
		Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
		Store: config.NewStore(t.TempDir()), SkillsFS: contractskills.FS,
	})
	if err := app.Run(context.Background(), []string{"skills", "install", "--target", target}); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(target, "contract-cli-contract-search")
	for _, name := range []string{"SKILL.md", "agents/openai.yaml", "references/text-search.md",
		"references/search-filter-fields.md",
		"references/search-contract-fields.md", "references/search-user-parameters.md",
		"references/search-app-parameters.md", "references/search-v2-parameters.md"} {
		if content := readTextFile(t, filepath.Join(root, name)); strings.TrimSpace(content) == "" {
			t.Errorf("installed resource %s is empty", name)
		}
	}
	assertSearchSkillLinksResolve(t, root)
	for _, name := range []string{"search-contract-fields.md", "search-user-parameters.md",
		"search-app-parameters.md", "search-v2-parameters.md"} {
		assertSearchDocumentLinksResolve(t, filepath.Join(target, "contract-cli-contract", "references", name))
	}
}

func assertSearchSkillLinksResolve(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Ext(path) == ".md" {
			assertSearchDocumentLinksResolve(t, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertSearchDocumentLinksResolve(t *testing.T, path string) {
	t.Helper()
	links := regexp.MustCompile(`\]\(([^)]+)\)`).FindAllStringSubmatch(readTextFile(t, path), -1)
	for _, link := range links {
		target, _, _ := strings.Cut(link[1], "#")
		if target == "" || strings.Contains(target, "://") {
			continue
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(path), target)); err != nil {
			t.Errorf("installed document %s has broken link %s: %v", path, target, err)
		}
	}
}

type documentedSearchRequest struct {
	ContractNumber   string         `json:"contract_number"`
	CombineCondition map[string]any `json:"combine_condition"`
	SearchTabCode    int            `json:"search_tab_code"`
	PageSize         int            `json:"page_size"`
	PageToken        string         `json:"page_token"`
	ConditionUnits   []struct {
		SearchField string `json:"search_field"`
		SearchValue string `json:"search_value"`
		UnionType   string `json:"union_type"`
	} `json:"condition_units"`
	FilterUnits []struct {
		SearchField     string `json:"search_field"`
		FilterUniqueKey string `json:"filter_unique_key"`
		SearchValue     any    `json:"search_value"`
		UnionType       string `json:"union_type"`
	} `json:"filter_units"`
}

func TestContractTextSearchExamplesMatchRequestedScope(t *testing.T) {
	requests := documentedSearchRequests(t, "text-search.md")
	expected := [][]string{
		{"CONTRACT_SCAN_TEXT", "CONTRACT_TEXT_FIELD", "CONTRACT_ATTACHMENT_TEXT", "CONTRACT_ARCHIVE_ATTACHMENT", "CONTRACT_CAUSE_FIELD"},
		{"CONTRACT_TEXT_FIELD"},
	}
	if len(requests) != len(expected) {
		t.Fatalf("text search recipes = %d, want %d", len(requests), len(expected))
	}
	for index, request := range requests {
		assertTextSearchRecipe(t, request, expected[index])
	}
}

func assertTextSearchRecipe(t *testing.T, request documentedSearchRequest, expected []string) {
	t.Helper()
	if request.SearchTabCode != 0 || request.PageSize < 1 || request.PageSize > 50 {
		t.Fatalf("invalid search tab or page size in recipe: %+v", request)
	}
	fields := make([]string, 0, len(request.ConditionUnits))
	keyword := ""
	for _, unit := range request.ConditionUnits {
		if unit.SearchValue == "" || unit.UnionType != "SHOULD" {
			t.Errorf("text keyword unit must have a nonempty value and SHOULD: %+v", unit)
		}
		if keyword != "" && unit.SearchValue != keyword {
			t.Error("all text sources must search the same requested keyword")
		}
		keyword = unit.SearchValue
		fields = append(fields, unit.SearchField)
	}
	sort.Strings(fields)
	sort.Strings(expected)
	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("recipe text scope = %v, want %v", fields, expected)
	}
}

func TestContractSearchDocumentedKeywordRequestsHaveShouldAlternative(t *testing.T) {
	for _, request := range documentedSearchRequests(t, "search-user-parameters.md") {
		if len(request.ConditionUnits) == 0 {
			continue
		}
		hasAlternative := false
		for _, unit := range request.ConditionUnits {
			hasAlternative = hasAlternative || unit.UnionType == "" || unit.UnionType == "SHOULD"
		}
		if !hasAlternative {
			t.Errorf("documented keyword request relies on unsupported MUST-only semantics: %+v", request)
		}
	}
}

func documentedSearchRequests(t *testing.T, name string) []documentedSearchRequest {
	t.Helper()
	return documentedSearchRequestsAtPath(t, filepath.Join("..", "..", "skills", "contract-cli-contract-search", "references", name))
}

func documentedSearchRequestsAtPath(t *testing.T, path string) []documentedSearchRequest {
	t.Helper()
	content := readTextFile(t, path)
	blocks := regexp.MustCompile("(?s)```json\\s*\\n(.*?)\\n```").FindAllStringSubmatch(content, -1)
	if len(blocks) == 0 {
		t.Fatalf("%s has no JSON request examples", path)
	}
	requests := make([]documentedSearchRequest, 0, len(blocks))
	for _, block := range blocks {
		var request documentedSearchRequest
		if err := json.Unmarshal([]byte(block[1]), &request); err != nil {
			t.Fatalf("invalid JSON request example in %s: %v", path, err)
		}
		requests = append(requests, request)
	}
	return requests
}
