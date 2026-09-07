package cli_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestContractSkillDocumentsMCPApprovalWorkflow(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "skills", "contract-cli-contract")
	skillContent := readTextFile(t, filepath.Join(root, "SKILL.md"))
	for _, reference := range []string{"approval-mcp-fields.md", "commands.md"} {
		if !strings.Contains(skillContent, "references/"+reference) {
			t.Fatalf("contract skill must link references/%s", reference)
		}
	}

	fieldReference := readTextFile(t, filepath.Join(root, "references", "approval-mcp-fields.md"))
	assertContainsAll(t, "approval-mcp-fields.md", fieldReference, []string{
		"contract approval get",
		"contract approval comment list",
		"contract approval comment create",
		"contract approval task list",
		"contract approval task approve",
		"contract approval task reject",
		"contract download-file",
		"--mention-id-type",
		"task_type_code",
		"approveAttachment",
		"110507",
		"300 秒",
		"非幂等",
		"执行结果不确定",
		"X-MCP-Response-Profile",
	})

	commands := readTextFile(t, filepath.Join(root, "references", "commands.md"))
	assertContainsAll(t, "commands.md", commands, []string{
		"contract-cli contract approval get <process-instance-id> --profile contract --as user",
		"contract-cli contract approval comment list <process-instance-id> --profile contract --as user",
		"contract-cli contract approval comment create <process-instance-id> --profile contract --as user",
		"contract-cli contract approval task list --profile contract --as user",
		"contract-cli contract approval task approve <task-instance-id> --profile contract --as user",
		"contract-cli contract approval task reject <task-instance-id> --profile contract --as user",
		"contract-cli contract download-file <file-id> --contract <contract-id> --profile contract --as user",
	})
}

func TestUserMCPApprovalInterfaceDocumentCoversSixAPIs(t *testing.T) {
	t.Parallel()

	document := readTextFile(t, filepath.Join("..", "..", "docs", "user-mcp-approval-interfaces.md"))
	assertContainsAll(t, "user-mcp-approval-interfaces.md", document, []string{
		"GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments",
		"GET /open-apis/contract/v1/mcp/contracts/{contract_id}/files/{file_id}/download",
		"GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}",
		"POST /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments",
		"POST /open-apis/contract/v1/mcp/tasks",
		"POST /open-apis/contract/v1/mcp/tasks/{task_instance_id}/approval",
		"Bearer <user_access_token>",
		"X-MCP-Response-Profile",
		"task_status",
		"user_id_type",
		"task_type_code",
		"approveAttachment",
		"expires_in_seconds",
		"执行结果不确定",
	})
}
