# CLI 与 MCP 合同搜索对齐：分析与验证记录

日期：2026-09-20，Asia/Shanghai。当前状态：已新增搜索字段发现命令并完成生产只读联调；人员、部门列表命令暂不新增。

## 目标与对齐基线

- 用户要求：本次改造后的 CLI 搜索与 MCP 搜索对齐，开发边界明确，对 Agent 友好。
- CLI 实测版本：正式版 1.8.6，提交 `f3b1549`；开发分支 `codex/release-20260920-lyy`，基线 `9e71b84`，相对发布版本新增的是 CI 配置。
- 环境：用户明确选择生产；CLI 使用独立配置目录 `/Users/lyy/.contract-cli-prod-verification-20260920` 的 `contract` profile，保留原非生产配置。
- MCP 连接：`contract-group-https`，地址 `https://mcp.qfei.cn/mcp-servers/clm-contract`。
- 范围：`search-contracts`、`list-contract-search-filter-fields`、`get-employees`、`get-departments`，以及搜索所需的已有分类查询。均为只读。
- 当前 MCP 连接与此次 CLI 授权是否为同一用户/租户尚未核对；本轮验证请求契约和响应结构，不据此宣称同账号结果集完全一致。
- app V1/V2 搜索有不同语义，保留其兼容性；本次对齐对象为 user MCP 搜索。

## 本次确认并完成的开发范围

用户在验证申请人姓名、部门名称可直接搜索后，将本次新增命令收敛为 `contract search-fields`，对应 MCP `list-contract-search-filter-fields`。此前分析中的人员/部门列表、既有 search 的错误退出、ID 类型映射、请求体数值保真等仍为后续待办，不包含在本次实现内。

- 命令：`contract-cli contract search-fields --keyword "字段展示名" --profile contract --as user --output json`。
- 契约：`GET /open-apis/contract/v1/mcp/contracts/search/filter_fields`，只读、仅 user，省略身份也使用 user；唯一 query 为可选 `keyword`。无语言、分页、caller ID 或 JSON 请求体参数。
- 来源：当前可调用 MCP 的说明与 input schema、当天导出元数据及对应 MCP requestTemplate；字段说明由服务端固定中文。未启用本地较新 CLM 源码的 V2/分页能力。
- 响应：完整保留基础字段、自定义字段、嵌套元数据与数值；新命令使用既有 MCP envelope 校验器，业务失败保留输出并返回错误，不影响原 `contract search` 行为。
- TDD：命令/ToolSpec/Service 缺失先失败，再实现；覆盖中文/特殊字符、空关键词、默认身份、app 拒绝、无请求体、未发布参数拒绝、业务错误、完整元数据/大整数和只读重试。
- 生产联调：独立开发版 `1.8.6-search-fields.local` 按真实展示名查到 1 个自定义文本字段，按返回示例原样保留真实 key/字段并替换测试值，合同搜索业务成功；无匹配字段也业务成功、返回空列表。均使用同一已授权 CLI 配置，未写入生产数据。
- 文档：主 Skill 场景 8、独立字段发现参考、共享路由、命令帮助、README 和测试计划同步；保留 9 场景结构及姓名/部门名称直接搜索规则。
- 检查：CLI 与合同 Service 全包通过；全量仅既有本地 ignored `mcp.yaml` 对齐检查失败，该旧文件也没有本次新增工具。新增契约测试不依赖该文件。全库 lint 仍为既有 42 项，本次修改文件无新增诊断。
- 本地交付：独立开发版在 `bin/contract-cli-search-fields`，查询/共享 Skill 已同步 Codex 与 agents 并备份；已安装正式 CLI 仍为 1.8.6。lint 关闭同类诊断截断后为 48 项，与原 HEAD 基线的文件/诊断逐项相同。

## 生产实测

仅记录协议与结果摘要，不保存 Token、真实人员/部门标识、姓名或合同明细。

| 场景 | 结果 | 结论 |
| --- | --- | --- |
| CLI 用户授权 | 成功，`auth status` 显示 prod / user / authorized | 正式版生产只读验证可执行 |
| CLI / MCP 首页，page_size=1 | 两侧 code=0；data 为 items / has_more / page_token | 当前响应结构一致 |
| 特制无命中编号 | 两侧 code=0、items=[]、has_more=false；page_token 仍非空 | 是否继续分页以 has_more 为准，不能仅看 token |
| CLI 非法 search_tab_code=2 | HTTP 200，code=111706，success=false，但进程退出码为 0 | 已复现 CLI 业务错误没有转为失败退出的问题 |
| CLI 非法 sort_type | HTTP 200，code=110000，success=false，但进程退出码为 0 | 同上；不得将业务失败解释成没有合同 |
| MCP 字段发现，keyword=合同编号 | 返回 CUSTOM_FIELD / CONTRACT_FORM_FIELDS_TEXT / TEXT，要求唯一 key，写入 filter_units | 同名自定义字段与系统编号不可混淆 |
| 用上述元数据构造自定义文本条件 | CLI 和 MCP 均接受请求；相同特制无命中值均返回空结果 | CLI 已有 JSON 输入能够传递该类条件；缺少发现入口 |
| MCP 部门列表 | 返回 department_id 为 od- 前缀字符串；列表含 status=0 候选 | 外部调用不能照内部 DTO 文档传内部数字 ID |
| status=0 部门 ID 原样传人员查询 | 返回 code=110000，参数有误 | 不能盲选目录第一项；此样例不足以断言所有部门链路都失败 |
| 从已观察合同的部门名称查候选 | 命中一个同名 status=1 部门 | 得到可用于进一步验证的正常候选 |
| 该正常部门 ID 原样传人员查询 | code=0，返回人员列表 | 正常候选的部门→人员链路已实测跑通 |
| MCP 人员列表 | 能返回人员；样例包含 status=0、数字字符串 user_id | 保留 ID 原值，不按 ou_ 前缀武断识别或改写 ID；状态用于解释和消歧 |
| CLI search-fields / employee / department 入口探测 | 都返回 unknown contract subcommand | 1.8.6 未封装这些入口；拟议命名尚未实现 |

当前生产 CLI 和 MCP 响应都没有 `data.pagination` 和 `contract_status_display_name` 等扩展字段。MCP 元数据及 CLI 文档有相关声明，但实际样例未提供；不能强制要求这些字段或补造统计值。这是文档/服务版本差异，不是 CLI 单独丢字段。

### 页面“合同文本”搜索对照（2026-09-20 16:39–16:41）

用户截图中关键词“里斯”返回 19 个结果，请求含 5 个文本字段。用同一生产 CLI profile、`search_tab_code=0`、`page_size=50` 控制变量验证：

| 条件 | 返回条目数 | 分页 |
| --- | --- | --- |
| 仅 `CONTRACT_TEXT_FIELD`，显式 `union_type=MUST` | 0 | has_more=false |
| 仅 `CONTRACT_TEXT_FIELD`，显式 `union_type=SHOULD` | 18 | has_more=false |
| 仅 `CONTRACT_TEXT_FIELD`，省略 union_type | 18 | has_more=false |
| 页面所示 5 类文本字段，每项 `union_type=SHOULD` | 19 | has_more=false |

5 类字段为 `CONTRACT_SCAN_TEXT`、`CONTRACT_TEXT_FIELD`、`CONTRACT_ATTACHMENT_TEXT`、`CONTRACT_ARCHIVE_ATTACHMENT`、`CONTRACT_CAUSE_FIELD`，每项 `search_value` 均为同一关键词。所有请求均 code=0、success=true；CLI 返回包含截图可见的合同编号。数量与截图一致，但未获取页面全量 ID 集合，不能声称已逐项比对全部结果和排序。

- 之前“未查到匹配合同”的答复必须纠正：该结果仅对应显式 MUST 请求，不能代表页面搜索或默认正文关键词搜索。
- 两个独立差异：MUST 与默认 SHOULD 的行为不同；页面“合同文本”覆盖 5 类文本，单独正文字段覆盖范围更小。不能只修字段数量而遗漏条件组合问题。
- 本地后端 `GroupQuerySchemaBuilder.buildSearchSource` 在构造关键词条件后无条件设置 `minimumShouldMatch(1)`，为 MUST-only 请求出现空结果提供源码线索；该源码不是生产部署版本证明，具体线上实现仍需按部署版本确认。
- CLI 1.8.6 已能透传正确请求。Agent/Skill 必须增加页面“合同文本”查询配方、正文限定配方、默认 SHOULD 说明及回归验收。不能将业务上“必须包含”机械映射为所有关键词单元 MUST，也不能为绕开空结果擅自把用户多个 AND 条件改成 OR。
- 搜索参考当前把 `CONTRACT_CAUSE_FIELD` 写为“原因关键词”，与合同附件语义不一致；本次 Skill 更新须一并核对正文、合同附件、其他附件、扫描件、归档附件的字段映射。

### 申请人姓名与部门名称搜索对照（2026-09-20 17:10–17:11）

正式 CLI 1.8.6、同一生产 `contract` profile / user、`search_tab_code=0`、`page_size=20`，单项 `condition_units`、显式 SHOULD。姓名与部门取自此前已验证合同，不把真实名称或人员标识写入仓库。

| 条件 | 本次返回 | 分页与核验 |
| --- | --- | --- |
| `CONTRACT_SUBMIT_NAME`，真实完整姓名 | 20 条 | has_more=true；本页申请人均为该姓名，包含已知合同 |
| `CONTRACT_SUBMIT_NAME`，上述姓名首字 | 20 条 | has_more=true；本页匹配两位不同姓名申请人，包含已知合同 |
| `CONTRACT_SUBMIT_NAME`，特制无匹配关键词 | 0 条 | has_more=false |
| `CONTRACT_DEPARTMENT_NAME`，已知合同的部门名称 | 20 条 | has_more=true；本页部门名称均符合请求 |

所有请求均 code=0、success=true。20 条是本次页面条目数，未遍历后续页，不是总量。姓名查询由服务端解析候选，不依赖 CLI 人员列表命令；完整姓名也可能重名，不能与唯一人员 ID 筛选等同。未测试停用人员、姓名与其他关键词的 AND 组合或所有结果的身份唯一性。主 Skill 场景 6 改用姓名关键词完整示例，将真实 ID 前置要求限定为精确人员筛选或“我申请的”。

## 初期识别的完整对齐待办（按当前范围分期）

1. 搜索字段发现已新增，复用现有 user OAuth、HTTP 客户端、只读重试和输出机制；人员列表/查询、部门列表/查询本次暂不新增，名称关键词使用既有搜索字段。
2. 更新 user 搜索业务错误处理：保留 JSON 错误响应，非零业务 code / success=false 必须产生失败退出。不要顺带改变 app 搜索契约。
3. 修正搜索 `user_id_type` 映射：MCP 支持默认 user_id、显式 union_id；CLI 当前 ToolSpec 固定 user_id，公共 flag 无法覆盖。需要对默认与显式 query 分别测试。
4. 保证搜索条件 JSON 类型与数值保真。当前 `resolveJSONObjectBody` 通过 `json.Unmarshal` 解码到 any，数值进入 float64；需用回归测试确认并修复大整数重编码损失，检查共享调用方影响。
5. 同步主 Skill、共享 Skill、Agent 提示、命令帮助和搜索参考文档；移除已补齐的能力缺口说明，修正文档中失实的参数与响应承诺。
6. 建立搜索参数及前置工具的契约测试，并用固定夹具覆盖端到端编排；核心测试不能依赖未纳入仓库、缺失即跳过的 mcp.yaml。
7. 补齐页面“合同文本”五字段查询配方和默认 SHOULD 规则；将上述 0/18/19 的生产对照纳入验收，追踪 MUST-only 服务端语义问题，不能仅验证 HTTP 成功或命令存在。

### ID 与值来源

- 生产外部入口的部门候选 ID 原样使用；不要在 CLI 二次转换内部 ID。
- 人员查询固定 user_id 的默认契约，搜索显式 union_id 是单独的可选契约；不能将一种 ID 的候选无条件送入另一种 ID 类型的请求。
- filter_units / condition_units 中 search_value 按字段元数据约定传递。通用网关转换并不意味着这些任意值会自动被转换。
- 分类可复用已有 category list；number / abbreviation / id 根据 request_paths 和值说明选择，不得混用。
- 枚举使用 value_scopes.value，保留其实际 JSON 类型；自定义币种等不能仅凭 label 推断 code。
- 对状态为删除、停用或未知的候选不能默认选取；用户查询历史人员时也不能擅自删掉该条件或静默替换成在职人员。

## Agent 工作流与兼容要求

1. 已知字段名称时按名称发现，不把用户的筛选值误当 keyword；仅明确查看目录时获取目录。
2. 根据字段名称、类型、唯一 key 以及候选完整性消歧；不得默认使用首项。当前页只有一项不等于候选唯一。
3. 读取元数据中的 request_location / request_paths / search_value_type / value_scopes / examples；紧凑目录缺少执行信息时，再查询字段详情。
4. 人员/部门从对应查询获得同环境、同身份的真实候选；需要时结合部门和状态消歧。
5. 替换示例占位值，正确合并多个条件。不得覆盖已有条件、擅自把 AND 改为 OR，或因字段解析失败而删除条件扩大查询。
6. 使用相同 profile 进行 user 搜索；区分成功空结果、参数错误、权限错误、请求失败。
7. 按 has_more / page_token 翻页，token 原样传递；没有返回总数时只描述已返回数量。将来返回 CONTRACT_GROUP 统计时区分合同组数与合同条目数。

本地 CLM 源码包含分页和 V2 字段匹配状态，但当前 MCP 工具参数仅暴露 keyword；新增字段分页参数及 V2 profile 的生产支持尚需验证。首版不自动切换响应 profile，不把本地源码的新契约当作生产既成事实。

MCP schema 未暴露 union_type，而部分参数描述建议使用 MUST；现有 CLI 文档还保留一段与当天 MCP 描述不一致的编号条件说明。应核对已部署行为，统一条件组合的可用契约，不能机械复制互相冲突的说明。

## 验收

- TDD：先复现参数映射、业务错误退出与数值精度差异，再做最小修改和回归。
- 单元/契约：中文与特殊字符编码、分页边界、user-only、默认/显式 ID 类型、响应字段及类型保真、完整错误响应。
- 集成：字段发现→文本/选项/金额/日期搜索；部门查询→人员查询→人员字段搜索；多条件合并、重名、无候选、失效候选、分页未完、零结果。
- 真实对照：确认同一租户与用户，用等价查询分别执行 CLI 和 MCP，对比合同 ID 集合、排序、分页和错误。当前跨连接样例不替代此项。
- Agent 验收：从用户业务表述出发，检查调用路径、字段选择、值来源和失败解释；记录调用次数/返回量，不能仅以接口 HTTP 200 判通过。
- Go 代码交付前运行 `go test ./...`、`golangci-lint run --enable=exhaustive ./...`。
- CLI 与 Skill 同版安装后，在新任务验证发现与编排流程；默认 skills install 会跳过已存在目录，更新需覆盖安装并先备份。

## 已完成的本地准备

- CLI 已安装 1.8.6，并调整登录 shell PATH 优先使用用户目录版本；旧系统版本仍保留。
- 11 个随包 Skill 已同步覆盖安装到 ~/.codex/skills 与 ~/.agents/skills，两个目录内容一致、版本均为 1.8.6；原内容已备份。
- 现有搜索、MCP ToolSpec、MCP 响应相关定向测试通过；未据此宣称新增能力已完成。
- 本次没有修改业务代码、创建合同或更改生产数据。

## 独立查询 Skill（2026-09-20 后续落实）

- 按用户要求新增 `skills/contract-cli-contract-search`，统一承接三种身份/接口的搜索文档，并增加 `references/text-search.md`。
- 页面五类文本与仅正文各有可执行 JSON 配方；正文 MUST 空结果、错误 envelope、分页终止和缺少字段元数据的处理都在主 Skill 可见。
- 原合同和 shared Skill 改为路由新目录；旧四份搜索参考仅保留链接入口，避免重复维护。响应参考修正可选扩展字段说明。
- 此次仅拆分 Skill 与相应文档/测试，字段发现、人员/部门新命令和业务错误退出修复仍待开发。
- 验证：`go test ./...` 通过；新 Skill 的 quick_validate、实际内嵌安装及引用链接、文本 JSON 配方、7 个独立 Agent 文档场景通过。`golangci-lint run --enable=exhaustive ./...` 仍报基线已有 42 项；关闭诊断截断后双方各 48 项逐项相同，本次无新增，但全库 lint 尚未通过。
- 本机三个相关 Skill 已从工作树同步到 `~/.codex/skills` 与 `~/.agents/skills`，安装版本为 `1.8.6-search-skill.local`；正式 CLI 二进制仍为 1.8.6。未发布新 npm/GitHub 版本，旧二进制的 `skills install --force` 仍会覆盖回旧内嵌文档。
- 后续主文件增加 8 类 user 场景提示与 3 个短 JSON；明确替换示例值、页签与申请人区别、金额/状态同时满足、日期角色和保持翻页条件。新增回归测试及 4 个独立 Agent 模拟通过，未据此声称生产组合查询已验证。
- 切回 `/Users/lyy/contract-cli` 后复查：CLI 全包测试通过；全量测试仅 `TestContractMCPToolSpecsStayAlignedWithMCPYAML` 因该目录已有且被 Git 忽略的 `mcp.yaml` 缺少 `list-process-comments` 失败，相关 Go 测试/ToolSpec 均未修改。之前独立工作树没有该本地文件，该检查被跳过；两次“全量”结果的差异来自此本地样本。
- 按用户进一步反馈，将主文件场景表与分散的 3 个示例重排为 9 个场景（人员/部门拆开），每节依次给出“1. 用户怎么说 / 2. 应该怎么查 / 3. 示例 JSON”。人员、选项字段和续页的占位模板都紧邻标注前置条件；续页测试校验除 token 外与原请求完全一致。
