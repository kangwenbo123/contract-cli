# AI 变更记录：合同 skill 字段补全

## 2026-09-15 多合同标题增加序号

- 改动内容：多合同标题前按展示顺序增加从 1 开始的连续序号，保留名称加粗、括号内编号及分隔线。
- 影响范围：合同 Skill 主入口及下载引用；单合同不加序号。
- 关键决策：失败或无文件合同也占序号，保证与展示分组一致。
- 验证结果：静态核对序号规则及 Markdown 格式，diff 检查通过。
- 未处理项或风险项：未重新打包。

## 2026-09-15 去掉合同名称标签

- 改动内容：合同标题仅加粗实际名称，去掉“合同名称：”标签，同行保留括号内合同编号。
- 影响范围：合同 Skill 主入口及下载引用。
- 验证结果：静态核对标题格式及分组衔接，diff 格式检查通过。
- 未处理项或风险项：未重新打包。

## 2026-09-15 合并合同名称与编号展示

- 改动内容：合同信息改为单行“**合同名称：…**（合同编号：…）”，名称加粗，编号括号内展示且不加粗。
- 影响范围：合同 Skill 主入口及下载引用，保留三列表格、总统计和合同间分隔线。
- 验证结果：静态核对单行排版与加粗范围，diff 格式检查通过。
- 未处理项或风险项：未重新打包。

## 2026-09-15 固定下载总统计位置

- 改动内容：结果第一行固定展示本次处理合同数、文件数及成功/失败/跳过数量，置于所有合同分组之前。
- 影响范围：合同 Skill 主入口及下载引用。
- 关键决策：清单不完整时明确写“已发现”，合同查询失败与文件下载失败分开统计。
- 验证结果：静态核对完整/不完整清单及统计口径，diff 格式检查通过。
- 未处理项或风险项：未重新打包。

## 2026-09-15 增加合同分组间距

- 改动内容：相邻合同之间增加 Markdown 水平分隔线，上下留空行；合同信息与表格之间留空行。
- 影响范围：合同 Skill 主入口及下载引用，仅调整展示排版。
- 验证结果：静态核对分隔线位置及 Markdown 空行规则，diff 格式检查通过。
- 未处理项或风险项：未重新打包，未进行宿主回放。

## 2026-09-15 明确合同名称和编号标签

- 改动内容：下载文件表格前分别展示“合同名称：…”和“合同编号：…”，单合同及多合同统一适用。
- 影响范围：合同 Skill 主入口及下载引用，文件表格保持三列。
- 关键决策：缺失值明确标注“未返回”，不将合同 ID 当作合同编号。
- 验证结果：静态核对标签、缺失值及分组规则，diff 格式检查通过。
- 未处理项或风险项：未重新打包。

## 2026-09-15 精简下载展示列

- 改动内容：成功文件表仅保留文件名链接、来源、大小（B），移除状态、文件 ID、备注列。
- 影响范围：合同 Skill 主入口及下载引用，单合同及多合同统一适用。
- 关键决策：失败/跳过项和必要说明置于合同表格下方；ID 留作内部去重追踪，统计不丢失。
- 验证结果：静态核对表头、列数、失败信息与同名文件区分规则，diff 格式检查通过。
- 未处理项或风险项：未重新打包，未进行宿主回放。

## 2026-09-15 多合同下载分组

- 改动内容：多合同下载按合同分子目录、分表展示，总汇总后展示各合同统计。
- 影响范围：合同 Skill 主入口及下载引用，保留单合同目录行为。
- 关键决策：目录保留合同 ID；跨合同分别计数，同名合同不合并；查询失败和空文件合同保留分组，清单不完整时明确标识。
- 验证结果：静态检查同名合同、缺少名称/编号、部分失败、空结果与跨合同相同文件 ID 的规则；diff 格式检查通过。
- 未处理项或风险项：展示示例为模拟数据，未进行多合同宿主回放，未重新打包。

## 2026-09-15 恢复下载汇总表格展示

- 改动内容：重新补充下载结果表格，列为文件名链接、来源、大小（B）、状态、文件 ID、备注，顶部汇总数量和保存目录。
- 影响范围：合同 Skill 主入口及下载引用，不改变下载逻辑和 CLI 输出契约。
- 关键决策：同名不同 ID 分行、多来源合并；失败项保留原因、未知大小显示“—”；不展示临时 URL。
- 验证结果：静态核对成功、失败、同名、多来源及表格转义规则，diff 格式检查通过；未执行宿主 Agent 回放。
- 未处理项或风险项：旧安装包不包含本次修改，更新 Skill 并新建任务后生效。

## 2026-09-15 补充盖章比对范围及节点来源标签

- 改动内容：按字段下载新增盖章节点比对文件；任务附件来源按盖章/审批/归档节点名称展示“审批意见-上传附件”，评论含回复统一展示“评论附件”。
- 影响范围：合同 Skill 主入口和下载引用，未改 CLI 或后端代码。
- 关键决策：核对 OcrComparisonOpenVO、McpTaskInstanceResource、BpmUtil 和节点类型枚举；区分比对文件与意见附件，使用所属任务节点，未知类型不猜测，多来源去重后保留定位。
- 验证结果：Skill 格式及 diff 校验通过；静态核对响应字段和节点类型解析规则，未进行 Agent 会话或线上下载实测。
- 未处理项或风险项：旧包不含本次修改；归档主文件归属信息不足的后端问题仍未修复。

## 2026-09-15 按前端纠正文件选择与展示规则

- 改动内容：重新补充正确的合同附件/其他附件映射、配置展示名、补充/终止协议语义及当前合同归档主文件范围。
- 影响范围：仅合同 Skill 主入口、下载引用及本记录，未修改 CLI、前端或后端代码。
- 关键决策：依据前端 RelatedFiles、ArchiveFiles、getCurrentContractScan 和后端 ContractFilesOpenVO；无法确认扫描件归属时不猜测；审批/评论/自定义字段保留来源，部分失败不宣称下载完成。
- 验证结果：Skill 格式和 diff 校验通过；按前后端取值规则静态复核，未进行线上回放。
- 未处理项或风险项：后端未过滤扫描件列表覆盖过滤结果且缺少 business_id 的问题仍待处理；已有安装包不含本次 Skill 修正。

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
