# 搜索条件组合

本页记录 user 搜索已验证的组合路径。验证使用同一授权身份和完整结果集合；不将单项可用推断为任意组合可用。

需求人与指定交易方的两个 filter/MUST 组合见 [完整配方](trading-party.md)：复用需求人外部 user_id，交易方按已确认主体 ID 走 `CONTRACT_TRADING_PARTY_ID`，保留两个条件并从第一页开始。不要把页面同名字段的 ID 数组或 label 命中当作精确主体筛选。

## 合同名称和正文同时满足

**用户怎么说**：“查名称含采购、正文包含里斯的合同。”

**应该怎么查**：名称放 `combine_condition.contract_name`，正文放 `condition_units.CONTRACT_TEXT_FIELD` 并使用 SHOULD。不要把名称和正文都放进 SHOULD 数组，那会成为任一关键词命中。名称为模糊匹配，不承诺唯一合同。

**示例 JSON**：将示例关键词换成用户实际要求。

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "combine_condition": {"contract_name": "采购"},
  "condition_units": [
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "里斯", "union_type": "SHOULD"}
  ]
}
```

2026-09-20 生产验证：正文完整结果 18 条；使用其中真实合同名称，名称单项和组合各返回 1 条，组合 ID 集合等于两者交集；仅把名称换成无匹配名称后返回 0 条。这里的“采购”为脱敏示例，未把示例值的命中数量当作实测。

## 合同名称和状态同时满足

**用户怎么说**：“查名称含采购、正在审批的合同。”

**应该怎么查**：名称和状态分别写 `combine_condition.contract_name`、`combine_condition.contract_status_in`，保留用户要求的页签。名称仍是模糊匹配。

**示例 JSON**：将名称、状态和页签替换为实际条件；本例范围是全部可见合同。

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "combine_condition": {"contract_name": "采购", "contract_status_in": "3"}
}
```

2026-09-20 生产验证：真实合同名称加其实际状态返回 1 条；相同名称换成另一有效状态返回 0 条，两组均完整。验证范围为全部可见页签；不能宣称“采购”、审批中或“我的合同”示例组合已有阳性样本。申请日期与状态组合的独立对照见 [申请日期说明](submission-date.md)。

## 申请人和正文同时满足

**用户怎么说**：“查张三申请、正文包含里斯的合同。”

**应该怎么查**：申请人用 `filter_units.CONTRACT_SUBMIT_ID` 的真实 `user_id` 字符串数组与 MUST；正文用 `condition_units.CONTRACT_TEXT_FIELD` 与 SHOULD。姓名关键词和正文不能直接并列后声称 AND；服务端可能重写姓名的 union_type 或移除无候选的姓名条件。

真实 ID 可复用同租户、同身份已确认的 `submitter_user_id`；没有可靠 ID 时，先按 [人员 Skill](../../contract-cli-employee/SKILL.md) 调用 `employee list --name` 取得候选。仍须确认其业务身份，不能从重名结果随便选一个；没有可靠 ID 时保留该条件并补齐信息。单独按姓名关键词查仍使用主 Skill 场景 6，不要求先取得 ID。

获取 ID 由 Agent 优先使用已有数据完成，不等于让用户查询技术 ID。成对读取 `submitter_user_name` 与非空 `submitter_user_id`，按真实 ID 去重；只看一个展示姓名或拿第一条结果都不足以确认人员。缺少可靠候选时，使用独立人员查询，不依赖按姓名搜合同反推；复用已有合同数据时须注意合同搜索不是完整人员目录，组展开可能带出其他申请人的合同；不能把未查完的候选称为完整，也不能用姓名无命中证明没有历史合同。用户明确指某个人、仍有同名或文字与截图不同等歧义时，先问业务身份。

MCP 的 `CONTRACT_SUBMIT_ID` 接收外部 `user_id` 字符串。页面抓包中的内部 employeeId 不能照搬，也不能把姓名填进 ID 数组。`filter_units.CONTRACT_SUBMIT_NAME` 不在筛选白名单，`combine_condition` 没有申请人姓名参数；不要借用需求人字段代替申请人。

**示例 JSON**：以下是模板，替换已确认的真实 ID 后才能执行。

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "condition_units": [
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "里斯", "union_type": "SHOULD"}
  ],
  "filter_units": [
    {"search_field": "CONTRACT_SUBMIT_ID", "search_value": ["REPLACE_WITH_CONFIRMED_USER_ID"], "union_type": "MUST"}
  ]
}
```

2026-09-20 生产验证：完整正文结果 18 条，按其中真实申请人 ID 的预期子集为 1 条，组合结果 ID 集合完全一致；相同 ID 加无匹配正文返回 0 条。该验证不证明姓名或部门名称关键词的 AND 可用。

## 申请人姓名与全部关键词：直接并列不能保证 AND

用户说“全部关键词搜毛鹏，同时限定某申请人”时，需要 `(十二字段中任一字段命中毛鹏) AND 申请人条件`。不能把申请人姓名直接追加到十二项中，即使写成 MUST：本地源码会把姓名解析后的单元改为 SHOULD，无候选时还可能删除姓名条件。

2026-09-20 生产反例：保留已完整查得的十二字段“毛鹏”，追加一个无匹配申请人姓名及 MUST，仍返回相同的 30 条、27 组，has_more=false、合同 ID 集合完全一致。预期应为 0 条，因此该写法不可靠；这不是“无匹配姓名也满足条件”，也不能将结果交付成指定申请人的合同。

随后用户确认以截图中的“范学冬（Peter）”为准。十二字段毛鹏追加该姓名 MUST，首批已返回 51 条、50 组、has_more=true，其中 16 条申请人不是目标，34 条不在完整毛鹏基线中；没有为错误组合继续翻页，不把首批数量当总数。这确认直接姓名写法还会扩大结果。

**用户怎么说**：“全部关键词搜毛鹏，并且是范学冬（Peter）申请的合同。”

**应该怎么查**：按人员 Skill 查询姓名候选取得 `user_id`，或从已取得的同身份合同结果成对确认姓名及真实外部 `submitter_user_id`；本次完整关键词基线中目标人员对应唯一 ID。保留十二项关键词，追加 `filter_units.CONTRACT_SUBMIT_ID/MUST`。ID 由 Agent 自动取得，用户不需要手工提供；重名或候选不足时再确认业务身份。

**示例 JSON**：以下为待替换真实 ID 的完整请求模板，不能原样执行。不要用姓名或页面内部 employeeId 替换占位值。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "condition_units": [
    {"search_field": "CONTRACT_NAME_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_NUMBER_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_OWNER_NAME", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_REMARK_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_TRADING_PARTY", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "ARCHIVE_NUMBER_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_FORM_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_CAUSE_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_SCAN_TEXT", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_ARCHIVE_ATTACHMENT", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_ATTACHMENT_TEXT", "search_value": "毛鹏", "union_type": "SHOULD"}
  ],
  "filter_units": [
    {"search_field": "CONTRACT_SUBMIT_ID", "search_value": ["REPLACE_WITH_CONFIRMED_USER_ID"], "union_type": "MUST"}
  ]
}
```

2026-09-20 生产正例：使用目标真实外部 ID 后返回完整 **5 条、3 组**，ID 集合恰好等于完整毛鹏结果中该申请人的子集；与截图 3 组计数及已匹配的页面样本对应。保持申请人 ID，将十二处关键词全部替换为无匹配词后返回 0 条、has_more=false。此项现已验证，不能扩展为任意词或任意组合均已验证；含分号的关键词仍遵守 [全部关键词边界](all-keyword-search.md#能力边界)。

以上限制针对姓名与其他关键词并列。用户明确要求姓名关键词时，**单个申请人姓名加独立状态**是另一种已验证结构，可直接用姓名，见 [姓名与状态配方](applicant-status.md)；指定具体申请人则遵循总纲，优先候选消歧与 ID 筛选。

## 分号批量编号和正文同时满足

**用户怎么说**：“在这两个编号的合同里，查正文包含里斯的。”

**应该怎么查**：将用户明确给出的真实编号以英文分号放入顶层 `contract_number`，正文仍为 SHOULD。不要把分号批量编号放 `condition_units.CONTRACT_NUMBER_FIELD`；该路径已实测出现正文限制失效。此配方仅适用 user 搜索，不套用于 app V2。

**示例 JSON**：两个编号均为占位值，替换后才能执行。

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "contract_number": "REPLACE_WITH_FIRST_NUMBER;REPLACE_WITH_SECOND_NUMBER",
  "condition_units": [
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "里斯", "union_type": "SHOULD"}
  ]
}
```

2026-09-20 生产验证：两个真实编号的关键词批量单查返回 2 条，叠加无匹配正文 MUST 后仍返回相同 2 条；改用顶层批量编号，正文无匹配时 0 条，正文“里斯”时 2 条且等于编号、正文完整结果集交集。不要从这一个组合推断任意批量附加条件均已验证。

## 多个正文关键词同时满足：当前无可靠配方

用户说“正文同时包含甲和乙”时，业务意图已经明确，不必反复确认。当前不能用重复的正文单元、一项 SHOULD 加一项 MUST 实现：2026-09-20 生产对照中单项正文“里斯”完整返回 18 条，而两个相同词的正文 SHOULD + MUST 返回 0 条，未保持预期等价关系。

不要把两个词拼接成短语、两个 SHOULD 或全部 MUST 来代替独立词 AND。页面五类来源的多词 AND 同样没有已验证的分组表达。告知用户当前无法可靠表达该组合，保留原需求；只有用户明确选择改变条件时才按新条件查询。该限制需要继续核实服务端能力，不能通过用户确认或 Skill 文案宣称已经支持。

## 严格边界及其他无法可靠表达的条件

用户要求“金额严格大于 5000”时，币种不明确可用业务语言询问；严格大于的含义已经明确，不重复追问“是否要大于”。已验证的 `[5000,null]` 包含 5000，不能代替严格比较，也不擅自把下界加 0.01。由 Agent 核实字段精度和服务端严格比较能力；缺少依据时说明当前范围查询包含边界，暂不能可靠执行该原始条件。

对于排除、多组 AND 或其他未有验证配方的条件，也先核实契约，不把请求成功或空结果当作语义正确。保留原条件，只在用户明确改变需求后采用新的条件；不能自动删条件、改 OR、截取首批后本地过滤并声称完整查询。
