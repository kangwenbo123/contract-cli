package cli

func addDirectoryHelp(registry map[string]helpTopic) {
	registry["employee"] = helpTopic{Name: "employee", Summary: "人员查询。", Usage: []string{"contract-cli employee <subcommand> [flags]"}, Commands: []helpCommand{{"contract-cli employee list [flags]", "按姓名或部门查询人员"}}}
	registry["department"] = helpTopic{Name: "department", Summary: "部门查询。", Usage: []string{"contract-cli department <subcommand> [flags]"}, Commands: []helpCommand{{"contract-cli department list [flags]", "查询部门候选"}}}
	registry["employee list"] = directoryListHelp("employee", "查询人员，返回 user_id、姓名、部门和状态。", []string{
		"--name 与 --department-id 必须二选一；姓名是关键词，不保证唯一人员。",
		"对应 get-employees；固定 user_id_type=user_id，原样保留所有候选及状态，不自动选择人员。",
		"status：1 正常/在职，0 删除/离职，2 停用，-1 未知；历史查询不能自动丢弃非在职人员。",
	})
	registry["department list"] = directoryListHelp("department", "查询部门，返回 department_id、名称、层级和状态。", []string{
		"--name 与 --department-id 互斥；均省略时查询部门列表的一页。",
		"对应 get-departments；department_id 是开放部门 ID（open_department_id），不是内部数字 ID。",
	})
}

func directoryListHelp(resource, summary string, notes []string) helpTopic {
	return helpTopic{
		Name: resource + " list", Summary: summary, Usage: []string{"contract-cli " + resource + " list [flags]"},
		Flags: []helpFlag{
			{"--profile <profile>", "使用已授权的 contract profile"}, {"--as <identity>", "仅支持 user，省略也使用 user"},
			{"--output <format>", "json / yaml / table"}, {"--raw", "原样输出响应"},
			{"--name <keyword>", "名称关键词"}, {"--department-id <id>", "单个开放部门 ID；映射 department_collection，原样使用 department list 返回值"},
			{"--page-size <n>", "每页 1–200 条，默认 10"}, {"--page-token <token>", "原样透传上一页返回的 token"},
		},
		Examples: []string{"contract-cli " + resource + " list --name 测试 --profile contract --as user --output json"},
		Notes:    append([]string{"只读 GET /open-apis/contract/v1/mcp/" + resource + "s；仅支持 user 身份。", "每次只查一页；仅 has_more=true 时保留全部条件继续翻页。业务失败返回非零退出码。"}, notes...),
	}
}
