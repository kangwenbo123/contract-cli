package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMDMSkillAgentMetadataMatchesAppIdentityRoutes(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills")
	tests := []struct {
		name     string
		file     string
		required []string
	}{
		{
			name: "vendor",
			file: filepath.Join(root, "contract-cli-mdm-vendor", "agents", "openai.yaml"),
			required: []string{
				"user or app",
				"vendor candidates",
				"vendor details",
				"create",
				"update",
				"patch",
				"enable",
				"disable",
				"certificate",
			},
		},
		{
			name: "legal",
			file: filepath.Join(root, "contract-cli-mdm-legal", "agents", "openai.yaml"),
			required: []string{
				"user or app",
				"legal entity candidates",
				"legal entity details",
				"create",
				"update",
				"code",
			},
		},
		{
			name: "fields",
			file: filepath.Join(root, "contract-cli-mdm-fields", "agents", "openai.yaml"),
			required: []string{
				"user or app",
				"vendor",
				"legal_entity",
				"vendor_risk",
				"user/MCP",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content := readMDMSkillMetadataText(t, tt.file)
			for _, forbidden := range []string{"user-only", "user-authorized"} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("%s must not keep obsolete identity wording %q: %s", tt.file, forbidden, content)
				}
			}
			for _, required := range tt.required {
				if !strings.Contains(content, required) {
					t.Fatalf("%s missing identity wording %q: %s", tt.file, required, content)
				}
			}
		})
	}
}

func TestVendorMaintenanceSkillDocumentsNewCommandsAndSafetyRules(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-mdm-vendor")
	checks := map[string][]string{
		filepath.Join(root, "SKILL.md"): {
			"contract-cli mdm vendor patch <vendor-id>",
			"contract-cli mdm vendor enable <vendor-id>",
			"contract-cli mdm vendor disable <vendor-id>",
			"本 Skill 只指导 Agent 执行 `contract-cli`",
			"不会改变或增强远程 MCP Server",
			"字段不确定时先查询字段配置",
			"子项 ID 不确定时先查询交易方详情",
			"存在歧义时先询问用户",
			"简单创建",
			"复杂创建",
			"附件替换",
			"部门查询后写入",
			"只允许选择 `status=1` 的启用部门",
			"`status=0` 的停用部门不得用于创建或修改",
			"个人创建不允许传 `vendorAccounts[].bankId`",
			"个人 PATCH 不允许传 `vendorAccounts[].bankId`",
			"未传字段不修改，`null` 请求清空",
			"`UNKNOWN`",
			"不得直接重试",
		},
		filepath.Join(root, "references", "commands.md"): {
			"contract-cli mdm vendor create --profile contract --as user",
			"contract-cli mdm vendor patch 7003410079584092448 --profile contract --as app",
			"contract-cli mdm vendor patch 7003410079584092448 --profile contract --as user",
			"contract-cli mdm vendor enable 7003410079584092448 --profile contract --as user",
			"contract-cli mdm vendor disable 7003410079584092448 --profile contract --as user",
		},
		filepath.Join(root, "references", "vendor-create-parameters.md"): {
			"POST /open-apis/mdm/v1/vendors",
			"POST /open-apis/contract/v1/mcp/vendors",
			"个人身份不允许传 `vendor`、`status`、风险字段和系统字段",
			"个人创建不允许传 `vendorAccounts[].bankId`",
			"`vendorAccounts[].bankId`（string，仅 App 可选，个人禁止）",
			"官方 App 请求体示例",
		},
		filepath.Join(root, "references", "vendor-patch-parameters.md"): {
			"PATCH /open-apis/mdm/v1/vendors/{vendor_id}",
			"PATCH /open-apis/contract/v1/mcp/vendors/{vendor_id}",
			"`_delete: true`",
			"未列出的记录保留",
			"个人 PATCH 不允许传 `vendorAccounts[].bankId`",
		},
		filepath.Join(root, "references", "vendor-status-parameters.md"): {
			"PUT /open-apis/contract/v1/mcp/vendors/{vendor_id}/status",
			"NO_CHANGE",
			"不发起审批",
		},
	}

	for file, fragments := range checks {
		content := readMDMSkillMetadataText(t, file)
		for _, fragment := range fragments {
			if !strings.Contains(content, fragment) {
				t.Fatalf("%s missing %q", file, fragment)
			}
		}
	}
}

func TestMDMSkillReferencesDoNotKeepObsoleteUserOnlyClaims(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("..", "..", "skills", "contract-cli-mdm-vendor", "references", "vendor-query-parameters.md"),
		filepath.Join("..", "..", "skills", "contract-cli-mdm-legal", "references", "entity-query-parameters.md"),
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()

			content := readMDMSkillMetadataText(t, file)
			for _, forbidden := range []string{
				"mdm legal` 和 `mdm fields` 仍保持 user-only",
				"mdm fields` 仍保持 user-only",
			} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("%s must not keep obsolete identity wording %q: %s", file, forbidden, content)
				}
			}
		})
	}
}

func readMDMSkillMetadataText(t *testing.T, file string) string {
	t.Helper()

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", file, err)
	}
	return string(content)
}
