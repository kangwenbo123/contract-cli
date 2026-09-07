# AI 变更记录：CLI MCP 审批支持

## 2026-09-07（代码复审修复）

- 改动内容：拆分审批与文件下载职责，补齐覆盖成功测试，并为不支持硬链接的文件系统增加独占创建降级。
- 影响范围：`internal/cli` 的 user 文件下载实现、测试、帮助文档及合同 Skill。
- 验证结果：目标红绿回归、`go test ./...`、`go test -race ./...`、`go vet ./...`、`make test`、Linux/Windows amd64 构建、gofmt 与 `git diff --check` 均通过。
- 未处理项或风险项：跨平台不再承诺统一原子替换；进程被强制终止时的文件系统持久化语义仍由操作系统决定。

## 2026-09-07

- 改动内容：新增 6 个 User MCP 审批接口的开发、联调和 CLI 使用说明。
- 影响范围：`docs/user-mcp-approval-interfaces.md`、README 文档入口及文档契约测试。
- 验证结果：文档契约红绿测试、`make test`、`git diff --check`、接口字段自检和本地 qfei-code-review 通过。
- 未处理项或风险项：不修改接口实现、CLI 行为、飞书原始接口文档、版本或依赖。

## 2026-09-05

- 改动内容：为六个 user MCP 审批接口补齐评论、任务、详情和文件下载命令，保留原 app 行为。
- 影响范围：`internal/cli`、`internal/openplatform`、合同 Skill、README、命令文档与测试计划。
- 关键决策：调用人只取个人 Token；`--mention-id-type` 仅解释被 @ 用户；不发送 `X-MCP-Response-Profile`。
- 关键决策：任务列表按读操作允许一次临时网络重试；评论创建和任务处理禁止自动重试。
- 关键决策：user 下载立即消费 300 秒预签名地址，不输出/记录 URL；校验大小后原子提交。
- 验证结果：`go test ./...`、`go vet ./...`、`go test -race ./...`、`make test`、Linux/Windows amd64 构建与本地代码审查通过。
- 测试修正：移除一个复用同一 `App` 的并行子测试标记，消除存量 `updateNotice` 数据竞争，不改生产逻辑。
- 未处理项或风险项：本期不修改后端、Swagger、飞书接口文档、Higress 配置、版本号或依赖。
