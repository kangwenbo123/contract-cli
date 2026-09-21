package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContractSkillFieldReferencesCoverDocumentedCommands(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	skillPath := filepath.Join(root, "skills", "contract-cli-contract", "SKILL.md")
	skillContent := readTextFile(t, skillPath)

	referenceFragments := map[string][]string{
		"contract-response-fields.md": {
			"contract-cli contract get",
			"contract_id",
			"counter_party_list",
			"payment_plan_list",
		},
		"patch-contract-fields.md": {
			"contract-cli contract patch",
			"ocr_file_id",
			"archive_attachment_map",
			"archive_attachment_file_ids",
		},
		"template-fields.md": {
			"contract-cli contract template list",
			"contract-cli contract template get",
			"template_fields",
			"value_scopes",
		},
		"template-instance-fields.md": {
			"contract-cli contract template instantiate",
			"template_field_list",
			"field_value",
			"create_employee_code",
		},
		"print-file-fields.md": {
			"contract-cli contract print-file",
			"operate_type",
			"file_id",
		},
		"category-fields.md": {
			"contract-cli contract category list",
			"category_resources",
			"abbreviation",
		},
		"share-cooperation-fields.md": {
			"contract-cli contract share get",
			"contract-cli contract cooperation record get",
			"cooperation_record_infos",
		},
		"contract-actions-fields.md": {
			"contract-cli contract upload-file",
			"contract-cli contract submit",
			"contract-cli contract download-file",
			"process_instance_id",
		},
		"openapi-gap-commands.md": {
			"contract-cli contract search-v2",
			"contract-cli contract field update",
			"contract-cli contract sign-url get",
			"contract-cli contract cooperation search",
			"contract-cli contract esign personal-auth-url",
			"contract-cli contract cooperation file download",
		},
	}

	for name, fragments := range referenceFragments {
		name := name
		fragments := fragments
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(skillContent, "references/"+name) {
				t.Fatalf("contract skill must link references/%s", name)
			}
			content := readTextFile(t, filepath.Join(root, "skills", "contract-cli-contract", "references", name))
			for _, fragment := range fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", name, fragment)
				}
			}
		})
	}
}

func TestContractSkillCommandsDoNotSuggestPatchTitleShortcut(t *testing.T) {
	t.Parallel()

	content := readTextFile(t, filepath.Join("..", "..", "skills", "contract-cli-contract", "references", "commands.md"))
	if strings.Contains(content, `contract patch 7023646046559404327 --profile contract --as app --data '{"title":"demo"}'`) {
		t.Fatalf("contract patch commands should not suggest title-only patch payload")
	}
}

func TestContractCreateSkillExplainsCategoryAbbreviationSource(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract")
	skillContent := readTextFile(t, filepath.Join(root, "SKILL.md"))
	createContent := readTextFile(t, filepath.Join(root, "references", "create-contract-fields.md"))
	fieldTreeContent := readTextFile(t, filepath.Join(root, "references", "create-contract-field-tree.md"))
	categoryContent := readTextFile(t, filepath.Join(root, "references", "category-fields.md"))

	assertContainsAll(t, "contract skill", skillContent, []string{
		"创建合同前",
		"category-fields.md",
		"contract_category_abbreviation",
	})
	assertContainsAll(t, "create-contract-fields.md", createContent, []string{
		"contract-cli contract category list",
		"data.contract_category_resource_vo.category_resources[].children[].abbreviation",
		"CATEGORY_ABBREVIATION_FROM_LIST",
		"111717",
	})
	if strings.Contains(createContent, `"contract_category_abbreviation": "PROCUREMENT"`) {
		t.Fatalf("create-contract-fields.md must not present PROCUREMENT as a reusable category abbreviation")
	}
	assertContainsAll(t, "create-contract-field-tree.md", fieldTreeContent, []string{
		"不是固定枚举",
		"不能替代 `contract_category_abbreviation`",
	})
	assertContainsAll(t, "category-fields.md", categoryContent, []string{
		"不是固定枚举",
		"相同的 `--profile` 和 `--as`",
		"末级分类",
	})
}

func TestPaymentAndApprovalSkillFieldReferencesCoverJSONBodyCommands(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	paymentSkillContent := readTextFile(t, filepath.Join(root, "skills", "contract-cli-payment", "SKILL.md"))
	contractSkillContent := readTextFile(t, filepath.Join(root, "skills", "contract-cli-contract", "SKILL.md"))

	referenceChecks := map[string]struct {
		skillContent string
		path         string
		fragments    []string
	}{
		"payment-fields.md": {
			skillContent: paymentSkillContent,
			path:         filepath.Join(root, "skills", "contract-cli-payment", "references", "payment-fields.md"),
			fragments: []string{
				"contract-cli payment create",
				"contract-cli payment update",
				"payment_status_code",
				"finance_system_code",
				"apply_amount",
			},
		},
		"payment-plan-fields.md": {
			skillContent: paymentSkillContent,
			path:         filepath.Join(root, "skills", "contract-cli-payment", "references", "payment-plan-fields.md"),
			fragments: []string{
				"contract-cli payment plan notify",
				"contract-cli payment plan search",
				"payment_lines",
				"must_conditions",
				"operator_type_code",
			},
		},
		"payment-record-fields.md": {
			skillContent: paymentSkillContent,
			path:         filepath.Join(root, "skills", "contract-cli-payment", "references", "payment-record-fields.md"),
			fragments: []string{
				"contract-cli payment record create",
				"contract-cli payment record update",
				"transaction_amount",
				"succeed_amount",
				"department_lark_id",
			},
		},
		"approval-fields.md": {
			skillContent: contractSkillContent,
			path:         filepath.Join(root, "skills", "contract-cli-contract", "references", "approval-fields.md"),
			fragments: []string{
				"contract-cli contract approval start",
				"task_instance_id",
				"command_type",
				"reject_info",
				"reject_return_code",
			},
		},
	}

	for name, check := range referenceChecks {
		name := name
		check := check
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(check.skillContent, "references/"+name) {
				t.Fatalf("skill must link references/%s", name)
			}
			content := readTextFile(t, check.path)
			for _, fragment := range check.fragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("%s missing %q", name, fragment)
				}
			}
		})
	}
}

func readTextFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}

func assertContainsAll(t *testing.T, label string, content string, fragments []string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("%s missing %q", label, fragment)
		}
	}
}
