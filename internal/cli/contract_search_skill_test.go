package cli_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestContractSearchSkillMetadataRoutesSearchIntents(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract-search")
	metadata := readTextFile(t, filepath.Join(root, "agents", "openai.yaml"))
	skill := readTextFile(t, filepath.Join(root, "SKILL.md"))
	frontmatterParts := strings.SplitN(skill, "---", 3)
	if len(frontmatterParts) != 3 {
		t.Fatal("contract search skill must contain YAML frontmatter")
	}
	discoveryDescription := frontmatterParts[1]

	for _, fragment := range []string{
		"搜索和筛选合同",
		"人员、部门、交易方",
		"自定义字段",
		"$contract-cli-contract-search",
		"user/app",
		"查询候选",
		"字段元数据",
		"完整分页",
		"allow_implicit_invocation: true",
	} {
		if !strings.Contains(metadata, fragment) {
			t.Errorf("contract search agent metadata missing discovery route %q", fragment)
		}
	}

	for _, fragment := range []string{
		"查合同",
		"合同文本包含某词",
		"按日期或状态筛选",
		"申请或需求",
		"部门",
		"交易方",
		"我方主体",
		"自定义字段",
	} {
		if !strings.Contains(discoveryDescription, fragment) {
			t.Errorf("contract search skill discovery description missing search intent %q", fragment)
		}
	}

	if strings.Contains(metadata, "有歧义时用业务字段和值向我澄清") {
		t.Errorf("contract search metadata must not encourage clarification before agent-side candidate and field discovery")
	}
}

func TestContractSearchSkillKeepsGuidanceWithoutStealingDefaultKeywords(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract-search")
	metadata := readTextFile(t, filepath.Join(root, "agents", "openai.yaml"))
	skill := readTextFile(t, filepath.Join(root, "SKILL.md"))
	guided := readTextFile(t, filepath.Join(root, "references", "guided-search.md"))
	allKeywords := readTextFile(t, filepath.Join(root, "references", "all-keyword-search.md"))

	for name, content := range map[string]string{
		"agents/openai.yaml":    metadata,
		"SKILL.md":              skill,
		"guided-search.md":      guided,
		"all-keyword-search.md": allKeywords,
	} {
		for _, fragment := range []string{"显式关键词", "页面默认", "业务角色", "推测"} {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing hybrid keyword guidance %q", name, fragment)
			}
		}
	}

	keywordRoute := strings.Index(skill, "显式关键词意图")
	guidedRoute := strings.Index(skill, "推测是人员、部门、交易方、我方主体或自定义字段")
	fallbackRoute := strings.Index(skill, "其他带搜索词的请求")
	if keywordRoute < 0 || guidedRoute < 0 || fallbackRoute < 0 ||
		keywordRoute >= guidedRoute || guidedRoute >= fallbackRoute {
		t.Fatalf("hybrid routes must be explicit-keyword, guided-inference, then default fallback: keyword=%d guided=%d fallback=%d", keywordRoute, guidedRoute, fallbackRoute)
	}

	for _, fragment := range []string{
		"“关键词搜索毛鹏”",
		"“搜索毛鹏”",
		"“毛鹏的合同”",
		"“项目区域为华东”",
		"“查华东”",
	} {
		if !strings.Contains(guided, fragment) {
			t.Errorf("guided-search.md missing routing example %s", fragment)
		}
	}

	if strings.Contains(skill, "| 明确某个人、某个部门、某个交易方或我方主体 |") {
		t.Error("entity-looking terms must not automatically become exact-object filters")
	}
	if strings.Contains(guided, "问：‘华东’是哪个字段的值") ||
		strings.Contains(guided, "问：“『华东』是哪个字段的值") {
		t.Error("a value-only search term must fall back to page default keyword search instead of blocking for a field")
	}
}

func TestContractSearchSkillSeparatesUserAndAppContracts(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract-search")
	skillContent := readTextFile(t, filepath.Join(root, "SKILL.md"))

	references := []struct {
		file      string
		fragments []string
	}{
		{
			file: "search-contract-fields.md",
			fragments: []string{
				"contract search --as user",
				"contract search --as app",
				"contract search-v2 --as app",
				"search-user-parameters.md",
				"search-app-parameters.md",
				"search-v2-parameters.md",
			},
		},
		{
			file: "search-user-parameters.md",
			fragments: []string{
				"contract-cli contract search --profile contract --as user",
				"POST /open-apis/contract/v1/mcp/contracts/search",
				"## CLI 参数映射",
				"## 请求体字段",
				"condition_units",
				"filter_units",
				"search_tab_code",
				"sort_type",
				"user_id_type",
				"filter_unique_key",
				"CONTRACT_GROUP",
				"## 枚举与约束",
				"## 示例",
			},
		},
		{
			file: "search-app-parameters.md",
			fragments: []string{
				"contract-cli contract search --profile contract --as app",
				"POST /open-apis/contract/v1/contracts/search",
				"## CLI 参数映射",
				"## 请求体字段",
				"精确查询",
				"110107",
				"combine_condition",
				"logic_search",
				"## 枚举与约束",
				"## 示例",
			},
		},
		{
			file: "search-v2-parameters.md",
			fragments: []string{
				"contract-cli contract search-v2 --profile contract --as app",
				"POST /open-apis/contract/v1/contracts/searchV2",
				"ES 模糊查询",
				"page_size",
				"10000",
				"condition_units",
				"filter_units",
				"忽略",
			},
		},
	}

	for _, reference := range references {
		reference := reference
		t.Run(reference.file, func(t *testing.T) {
			t.Parallel()

			link := "references/" + reference.file
			if !strings.Contains(skillContent, link) {
				t.Fatalf("contract skill must link %s", link)
			}
			content := readTextFile(t, filepath.Join(root, "references", reference.file))
			for _, fragment := range reference.fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", reference.file, fragment)
				}
			}
		})
	}

	if strings.Contains(skillContent, "想用新版搜索条件：用 `contract search-v2 --as app`") {
		t.Fatalf("contract skill must not describe search-v2 as the generic newer search contract")
	}
}

func TestContractSearchReferencesUseRunnableIdentitySpecificExamples(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract-search", "references")
	checks := []struct {
		file      string
		required  []string
		forbidden []string
	}{
		{
			file: "search-user-parameters.md",
			required: []string{
				"--as user --input-file",
				"--as user --data",
			},
		},
		{
			file: "search-app-parameters.md",
			required: []string{
				"--as app --input-file",
				"--as app --data",
			},
		},
		{
			file: "search-v2-parameters.md",
			required: []string{
				"--as app --input-file",
			},
			forbidden: []string{
				`"page_size": 0`,
			},
		},
	}

	for _, check := range checks {
		content := readTextFile(t, filepath.Join(root, check.file))
		for _, fragment := range check.required {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing runnable example fragment %q", check.file, fragment)
			}
		}
		for _, fragment := range check.forbidden {
			if strings.Contains(content, fragment) {
				t.Errorf("%s contains invalid example fragment %q", check.file, fragment)
			}
		}
	}
}

func TestContractSearchUserFilterValueContractsMatchCurrentCLIProfile(t *testing.T) {
	t.Parallel()

	content := readTextFile(t, filepath.Join(
		"..", "..", "skills", "contract-cli-contract-search", "references", "search-user-parameters.md",
	))

	required := []string{
		"当前 `contract-cli` 不发送 `X-MCP-Response-Profile`",
		"`CONTRACT_SUBMIT_ID` / `submitterEmployeeId` | `string` / `array<string>` | 飞书 `user_id`",
		"`CONTRACT_AMOUNT` / `contractAmount` | `array` | 恰好两个元素 `[start,end]`",
		"`CONTRACT_CURRENCY` / `contractCurrency` | `string` / `integer` / `array`",
		"`CONTRACT_SEAL_NUMBER` / `contractSealNumber` | `integer` / `array<integer>`",
		"`CONTRACT_FORM_FIELDS_OPTION` / `contractFormFieldsOption` | `string` / `array<string>` / `array<integer>`",
		"OPTION_LABEL_ARRAY",
		"CURRENCY_ID_ARRAY",
		"`CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT_ID` / `contractFormFieldsEmployeeDepartmentId` | `string` / `array<string>`",
		"JSON integer `0` 或 `1`",
		"不传 JSON boolean",
		"盖章份数",
	}
	for _, fragment := range required {
		if !strings.Contains(content, fragment) {
			t.Errorf("search-user-parameters.md missing filter value contract %q", fragment)
		}
	}

	forbidden := []string{
		"`0/1` 或 boolean",
		"印章编号 string",
		"单个 employeeId",
	}
	for _, fragment := range forbidden {
		if strings.Contains(content, fragment) {
			t.Errorf("search-user-parameters.md contains obsolete filter value contract %q", fragment)
		}
	}
}

func TestContractSearchFieldDiscoveryExplainsMetadataValueContracts(t *testing.T) {
	t.Parallel()

	content := readTextFile(t, filepath.Join(
		"..", "..", "skills", "contract-cli-contract-search", "references", "search-filter-fields.md",
	))
	for _, fragment := range []string{
		"data.items[]",
		"examples[].request_fragment",
		"语义类型",
		"OPTION_LABEL_ARRAY",
		"CURRENCY_ID_ARRAY",
		"DATETIME_STRING_RANGE",
		"MILLIS_RANGE_ARRAY",
		"空候选",
		"REPLACE_WITH_FILTER_UNIQUE_KEY",
	} {
		if !strings.Contains(content, fragment) {
			t.Errorf("search-filter-fields.md missing metadata conversion contract %q", fragment)
		}
	}
}
