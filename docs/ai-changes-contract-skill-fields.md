# AI 变更记录：合同 skill 字段补全

## 2026-09-15 按字段下载文件规则

- 改动内容：明确归档文件合集及六类文件字段映射，更新合同 Skill 导航和下载引用。
- 影响范围：仅 Skill 与本变更记录，未修改 CLI 或接口代码。
- 关键决策：按字段筛选后去重，多文件全部收集；自定义字段精确定位，同名歧义先确认；字段缺失不替换为其他分类，指定合同字段不额外收集流程附件。
- 验证结果：静态核对字段映射、多文件、归档合集、自定义字段歧义与缺失处理；Skill 格式校验和 diff 检查通过。
- 未处理项或风险项：文件选择仍由 Agent 按 Skill 编排，未进行线上下载实测；已有安装包不含本次修改。

## 2026-05-26 合同命令字段参考补齐

变更摘要：以 `contract create` 的字段文档结构为模板，为 `contract-cli-contract` 补齐搜索、详情响应、模板、模板实例、更新、打印文件、分类、分享协商和轻量动作命令的字段参考。

涉及文件/模块：

- `skills/contract-cli-contract/SKILL.md`
- `skills/contract-cli-contract/agents/openai.yaml`
- `skills/contract-cli-contract/references/*.md`
- `internal/cli/contract_skill_reference_test.go`

关键逻辑/决策：

- 只补 skill 文档和文档契约测试，不改变 CLI 命令行为、调用链路或接口参数解析。
- `contract search` 补充 `combine_condition`、`logic_search`、状态枚举、收支类型和分页字段，便于直接组装查询 JSON。
- `contract patch` 按官方“更新合同”文档收敛为文件/归档字段，不再引导使用 `title` 或 `contract_name` 这类未确认 payload。
- `template list/get` 与 `template instantiate` 分别补模板字段读取和 `template_field_list.field_value` JSON 字符串格式，支撑从模板详情到实例创建的闭环。
- `sync-user-groups`、`text`、`enum list` 暂未补到 `contract create` 级别字段完备度，因为本次未找到对应完整字段页。

验证策略：

- 新增文档契约测试，要求 `SKILL.md` 链接所有新增 reference，并要求关键字段片段存在。
- 本地执行 `go test ./...`、`contract-cli skills list` 和 `contract-cli skills install --target <临时目录>`。
- dev 环境命令验收单独记录每条命令的通过、失败或 fixture 缺失状态，不输出 token、app secret 或数据库密码。

## 2026-05-26 dev user/app 验收补充

环境结果：

- `contract-dev` user OAuth 已授权。
- `contract-dev` app credentials 已授权，tenant access token 换取成功。
- dev 只读库 `clm_dev` 可查询 fixture；未直接修改数据库。

验收结果：

- user 通过：`contract category list`、`contract enum list`、`contract template list`、`contract template get`、`contract template instantiate`、`contract search`、`contract get`、`contract text`。
- app 通过：`contract upload-file`、`contract category list`、`contract template list`、`contract template get`、`contract template instantiate`、`contract search`、`contract submit`、`contract patch`、`contract print-file`、`contract download-file`、`contract delete`、`contract share get`、`contract cooperation link get`、`contract cooperation record get`。
- `contract upload-file --as user` 在 dev 后端返回 403，错误为 OAuth user token 不允许访问文件上传接口；同一命令用 app 身份验证通过。
- `contract resubmit` 未安全闭环：专用 smoke 合同不能直接创建为已拒绝或已撤回状态，后端返回“合同状态参数非法”；不对普通业务合同执行重提。

专用测试数据：

- 创建 smoke 前缀：`contract-cli-skill-field-smoke-20260526114343`、`contract-cli-skill-field-smoke-20260526114522`。
- 写操作均作用于 smoke 前缀合同或 smoke 上传文件；分享、协商查询使用只读 fixture。
