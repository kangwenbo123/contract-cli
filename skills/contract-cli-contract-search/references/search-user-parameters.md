# contract search User MCP Parameters

本页专用于 user 身份下的 MCP 合同搜索。

- 命令：`contract-cli contract search --profile contract --as user`
- 接口：`POST /open-apis/contract/v1/mcp/contracts/search`
- 身份：仅本页语义使用 `user` OAuth 身份
- 参数基线：MCP 工具 `search-contracts` 描述；CLM `MCPContractOpenPlatformController` 与 `McpContractSearchRequestDTO`
- 权限：服务端自动注入当前登录用户的可见范围，不需要在 body 中传 `user_id` 或 `permission`

## 目录

- [CLI 参数映射](#cli-参数映射)
- [请求体字段](#请求体字段)
- [组合条件](#组合条件)
- [关键词条件 condition_units](#关键词条件-condition_units)
- [页面全部关键词搜索](all-keyword-search.md)
- [申请人姓名与合同状态](applicant-status.md)
- [合同需求人](demand-person.md)
- [需求人与交易方](trading-party.md)
- [合同文本搜索配方](text-search.md)
- [结构化筛选 filter_units](#结构化筛选-filter_units)
- [范围、日期与部门角色](#范围日期与部门角色)
- [合同申请日期实测](submission-date.md)
- [排序与分页](#排序与分页)
- [枚举与约束](#枚举与约束)
- [动态字段发现](#动态字段发现)
- [示例](#示例)
- [响应摘要](#响应摘要)

## CLI 参数映射

| 参数 | 请求位置 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- | --- |
| `--as user` | 本地身份 | enum | 建议显式传 | 使用当前 OAuth 用户，并固定路由到 MCP 搜索。 |
| `--contract-number` | `$body.contract_number` | string | 可选 | 覆盖请求体同名字段。 |
| `--page-size` | `$body.page_size` | integer | 可选 | 覆盖请求体同名字段。 |
| `--page-token` | `$body.page_token` | string | 可选 | 覆盖请求体同名字段。 |
| `--input-file` | `$body` | JSON file | 与 `--data` 互斥 | 复杂条件推荐使用。 |
| `--data` | `$body` | JSON string | 与 `--input-file` 互斥 | 适合简单查询。 |
| `--user-id-type` | `$query.user_id_type` | string | 可省略 | 当前 CLI user 搜索固定 `user_id`；本次开发版仅接受省略或显式 `user_id`，其他值在发送搜索请求前报错。旧 CLI 1.8.6 会静默忽略覆盖值，不能据此声称已切换 ID 类型。 |
| `--user-id` | `$query.user_id` | string | 不需要 | MCP 根据 OAuth 登录态识别用户；不要用它代替 user 登录。 |
| `--profile` | 本地上下文 | string | 可选 | 不传时使用当前 profile。 |
| `--output` | CLI 输出 | enum | 可选 | `json`、`yaml`、`table`，默认 `json`。 |
| `--raw` | CLI 输出 | boolean | 可选 | 原样输出服务端响应。 |

除前三个 body flag 外，其余搜索字段都通过 `--input-file` 或 `--data` 放入 JSON body。

本次开发版保留输入 JSON 数值的原始表示；生成请求时也应避免先转成会丢精度的浮点数。字符串人员/部门 ID 仍保持字符串，不改成数值。超大整数、高精度小数与科学计数的保真已通过本地 HTTP 请求链路测试，生产只验证普通金额修复前后完整结果集一致；这不代表服务端支持任意大数或精度。旧 CLI 1.8.6 的解析不能保证此能力。

## 请求体字段

顶层参数都不是 schema 必填项；空对象表示在当前用户可见范围内按默认页签、排序和分页查询。

| JSON 路径 | 类型 | 必填性 | 默认值/约束 | 说明 |
| --- | --- | --- | --- | --- |
| `contract_number` | string | 可选 | 非空关键词；分号可分隔批量编号 | 顶层编号与 `condition_units` 编号的组合路径不同，不能概括为向同一个 SHOULD 数组追加条件；见 [组合规则](search-combinations.md)。 |
| `combine_condition` | object | 可选 | 基础字段组合模型 | 按元数据和已验证组合选用；申请日期、名称与正文/状态的组合使用此区域，不统一改写为搜索单元。 |
| `condition_units` | array<object> | 可选 | 最多 50 项 | PC 兼容关键词条件。 |
| `filter_units` | array<object> | 可选 | 最多 50 项 | PC 兼容结构化筛选条件。 |
| `sort_type` | string | 可选 | 默认 `SUBMITTED_TIME` | MCP 排序字段。 |
| `order` | string | 可选 | 默认 `DESC` | `ASC` 或 `DESC`。 |
| `sort` | string | 可选 | 旧兼容字段 | 仅 `asc` 表示升序；否则按提交时间降序。优先使用 `sort_type/order`。 |
| `search_tab_code` | integer | 可选 | 默认 `0`；枚举 `0/1` | `0` 全部合同；`1` 我的合同；暂不支持草稿页签。 |
| `page_size` | integer | 可选 | 默认 `10`；建议不超过 `50` | 合同编号分号批量模式最大 `50`。 |
| `page_token` | string | 可选 | 首次不传 | 翻页时原样回传响应中的 token。 |
| `lang` | string | 可选 | 后端默认语言 | 例如 `zh-CN`、`en-US`。 |

`user_id_type` 位于 query，不属于 JSON body。

## 组合条件

`combine_condition` 用于基础字段，MCP 搜索仍使用其中的日期与组合条件。MCP 会自行填充权限，因此本模型不列出 `user_id` 和 `permission`。

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `combine_condition.archive_number` | string | 可选 | 归档编号，精确匹配。 |
| `combine_condition.contract_category_abbreviation` | string | 可选 | 合同分类缩写，精确匹配。 |
| `combine_condition.contract_category_name` | string | 可选 | 合同分类名称，精确匹配。 |
| `combine_condition.contract_name` | string | 可选 | 合同名称，支持模糊匹配。 |
| `combine_condition.contract_number` | string | 可选 | 旧组合条件编号语义；新搜索按场景使用顶层编号，不能将各编号路径视为可互换；见 [组合规则](search-combinations.md)。 |
| `combine_condition.contract_status_in` | string | 可选 | 状态 code，英文逗号分隔。 |
| `combine_condition.pay_type` | integer | 可选 | `1` 收入、`2` 支出、`3` 收入支出、`4` 无金额。 |
| `combine_condition.currency_in` | array<string> | 可选 | 币种英文码，例如 `["CNY","USD"]`。 |
| `combine_condition.demand_employee_ids` | array<string> | 历史兼容项，不推荐 | MCP 旧描述与 CLM/网关的 ID 转换契约存在冲突，不能凭名称认定为内部或外部 ID。新查询使用字段元数据指定的 `filter_units.CONTRACT_DEMAND_PERSON`，见 [需求人说明](demand-person.md)。 |
| `combine_condition.saas_demand_department_id_in` | array<string> | 可选 | 需求人部门 ID 集合。 |
| `combine_condition.use_template` | integer | 可选 | `1` 使用模板；`0` 不使用模板。 |
| `combine_condition.create_time_start/end` | string | 可选 | 创建时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.update_time_start/end` | string | 可选 | 更新时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.submited_time_start/end` | string | 可选 | 页面“合同申请日期”，即系统提交时间；保留 `submited` 拼写，见 [实测及边界](submission-date.md)。 |
| `combine_condition.archived_time_start/end` | string | 可选 | 归档时间范围。 |
| `combine_condition.contract_sign_date_start/end` | string | 可选 | 签订时间 `signed_time` 范围。 |
| `combine_condition.contract_signed_date_start/end` | string | 可选 | 签约日期 `signed_date` 范围。 |

所有时间字符串使用 `YYYY-MM-DD HH:mm:ss`。

## 关键词条件 `condition_units`

单项结构：

| 字段 | 类型 | schema 必填性 | 执行约束 |
| --- | --- | --- | --- |
| `search_field` | string | 未标必填 | 有效条件必须提供；可传枚举名或 PC 字段名。 |
| `search_value` | string/number/boolean/array/object | 未标必填 | 有效条件必须提供；绝大多数关键词字段传 string。 |
| `union_type` | string | 可选 | 接口接受 `MUST`、`SHOULD`、`MUST_NOT`；关键词默认 `SHOULD`。文本场景使用默认值或显式 SHOULD，注意下述已验证限制。 |
| `filter_unique_key` | string | 可选 | 通常不传；限定具体自定义字段时传 `attribute_key`。 |

可用关键词字段：

| 枚举名 | PC 字段名 | 搜索值与说明 |
| --- | --- | --- |
| `CONTRACT_NAME_FIELD` | `contractName` | 合同名称关键词 string。 |
| `CONTRACT_NUMBER_FIELD` | `contractNumber` | 合同编号 string；包含 `;` 或 `；` 时进入批量编号模式，`page_size<=50`。该位置的批量编号已复现其他关键词条件被忽略，组合查询见 [组合规则](search-combinations.md)。 |
| `CONTRACT_CREATOR_FIELD` | `contractCreator` | 创建人姓名关键词。 |
| `CONTRACT_OWNER_NAME` | `ownerEmployeeName` | 归属人姓名；后端先解析 employeeId。 |
| `CONTRACT_SUBMIT_NAME` | `submitterEmployeeName` | 申请人姓名关键词 string；后端先解析 employeeId，无需 CLI 预先取得人员 ID。完整姓名也不保证唯一人员。 |
| `CONTRACT_DEPARTMENT_NAME` | `departmentName` | 页面“所属部门”的名称关键词；后端先解析部门 ID。不要与需求部门、自定义部门或 `DEPARTMENT_AUTHORITY` 的 ID 筛选默认等同，见下文角色说明。 |
| `CONTRACT_TRADING_PARTY` | `contractTradingParty` | 交易方关键词。 |
| `CONTRACT_TRADING_PARTY_NAME_PRECISE` | `contractTradingPartyNamePrecise` | 本地实现为交易方名称包含匹配；PRECISE 不保证名称相等或唯一主体，精确主体使用 filter 的 `_ID` 路径。 |
| `CONTRACT_LEGAL_ENTITY_NAME_PRECISE` | `contractLegalEntityNamePrecise` | 我方主体名称；后端先解析主体 ID。 |
| `CONTRACT_REMARK_FIELD` | `contractRemark` | 备注关键词。 |
| `CONTRACT_CAUSE_FIELD` | `contractCause` | 合同附件文本。 |
| `CONTRACT_FORM_FIELD` | `dynamicForm` | 动态表单关键词。 |
| `CONTRACT_FORM_FIELDS_TEXT` | `contractFormFieldsText` | 自定义文本字段关键词。 |
| `ARCHIVE_NUMBER_FIELD` | `archiveNumber` | 归档编号关键词。 |
| `CONTRACT_TEXT_FIELD` | `contractText` | 合同正文文本。 |
| `CONTRACT_SCAN_TEXT` | `contractScan` | 扫描件文本。 |
| `CONTRACT_ARCHIVE_ATTACHMENT` | `contractArchiveAttachment` | 归档附件文本。 |
| `CONTRACT_ATTACHMENT_TEXT` | `contractAttachment` | 其他附件文本。 |

页面“合同文本”是五类文本字段的同词 OR 组合，不等于单独的 `CONTRACT_TEXT_FIELD`。完整请求见 [合同文本搜索配方](text-search.md)，其中同时提供页面范围与仅正文两种方式。

页面“全部”关键词另用 [十二字段同词 OR 配方](all-keyword-search.md)，不等于五类文本或单独申请人。截图中的 `CONTRACT_OWNER_NAME` 是归属人，不能替换成 `CONTRACT_SUBMIT_NAME`。`CONTRACT_FORM_FIELD` 是动态表单关键词，不需要编造一个自定义字段 key，也不替代针对具体自定义字段的结构化筛选。

MCP DTO 未声明或透传页面 `allConditionUnits`，也没有对应的 `all_condition_units` 参数；不要加入请求。源码中该标志影响人员候选与表单人员扩展，并非 AND/OR 开关。当前生产“毛鹏”样本的十二字段请求已与页面计数、可见样本对应，但不能据此宣称所有人员候选及页面行为完全一致。

用户明确要求申请人姓名关键词时，可直接使用 `CONTRACT_SUBMIT_NAME` + 字符串关键词 + SHOULD，完整请求见 [主 Skill 场景 6](../SKILL.md#场景-6按申请人姓名关键词搜索)。2026-09-20 生产只读实测已验证完整姓名、部分姓名和无匹配关键词；部分姓名可命中多位申请人。指定某个人时优先确认真实 `user_id` 并使用 `filter_units.CONTRACT_SUBMIT_ID`，不要声称姓名查询与 ID 筛选始终等价，尤其不能由姓名无命中推断停用人员没有历史合同。

用户明确要求姓名关键词时，单个申请人姓名 SHOULD 与 `combine_condition.contract_status_in` 可组合查询，范学冬＋变更中已与真实 ID 路径做完整结果对照；指定人员仍优先 ID，详情见 [姓名与状态](applicant-status.md)。姓名与十二字段关键词并列则不同：直接姓名 MUST 已实测扩大或丢失姓名限制，按 [组合规则](search-combinations.md#申请人姓名与全部关键词直接并列不能保证-and) 使用真实 ID。

生产正文查询已复现：同词单项 `MUST` 返回空，默认/显式 `SHOULD` 能命中；不要把接受参数等同于预期搜索语义已验证。本文普通关键词示例采用 SHOULD，复杂多个 AND/排除条件需单独确认，不能擅自改成 OR。

`CONTRACT_CURRENCY` 不在 `condition_units` 白名单，币种必须放在 `filter_units`。

部分工具描述建议将编号放入 `filter_units.CONTRACT_NUMBER_FIELD + MUST`，但 filter 白名单和当前 CLM 策略未列出该字段，不能依赖此冲突写法。单独编号查询可用顶层 `contract_number`；多条件请求按 [组合规则](search-combinations.md) 选择已验证路径。2026-09-20 生产对照：两个真实编号放在 `condition_units.CONTRACT_NUMBER_FIELD` 时，追加无匹配正文 MUST 仍返回相同两条；同样批号放在顶层 `contract_number` 时，无匹配正文 SHOULD 返回 0 条，有匹配正文则返回两条且等于对照结果集交集。这仅验证该路径与样本，不代表任意编号、任意关键词组合都已验证。

## 结构化筛选 `filter_units`

当前 `contract-cli` 不发送 `X-MCP-Response-Profile`，服务端按默认 `clm-web-status-v1.0`（legacy）处理。下表按 MCP 契约与本地源码整理取值形状，不表示每项均已完成生产命中验证；实测范围在对应参考中单独注明。其中标记为单值或数组的字段，Agent 优先使用字段发现接口返回的数组形态，兼容更严格的 MCP profile。

单项结构与 `condition_units` 相同，但默认 `union_type=MUST`。自定义字段必须同时传 `filter_unique_key=attribute_key`。范围值固定包含两个位置；当前 legacy 路径允许用 `null` 表示单边范围，但不能两端同时为 `null`，两端都有值时应满足 `start<=end`。

页面抓包中的 `filter_units.CONTRACT_STATUS` 不在 MCP 白名单，已实测业务失败；按生产元数据将状态写入 `combine_condition.contract_status_in`，类型 `STATUS_CODE_CSV`。例如变更中是字符串 `"10"`，已变更是 `"11"`，不传 `[10]` 或中文 label。

固定字段：

| 枚举名 / PC 字段名 | `search_value` JSON 类型 | 约束 |
| --- | --- | --- |
| `CONTRACT_SUBMIT_ID` / `submitterEmployeeId` | `string` / `array<string>` | 飞书 `user_id`；legacy 支持单值或非空数组，Agent 优先传数组。服务端再解析为内部 employeeId。 |
| `DEPARTMENT_AUTHORITY` / `departmentAuthority` | `string` / `array<string>` | 字段元数据显示为“申请部门”，值为部门 OPENID；不是部门名称。Agent 优先传非空数组，不视为名称字段的等价精确版本。 |
| `CONTRACT_DEMAND_PERSON` / `demandEmployeeId` | `string` / `array<string>` | 需求人的飞书 `user_id`；Agent 优先传非空数组。姓名与 PC 内部 employeeId 不能代替；优先由 employee list 查询姓名候选，已有合同详情也可交叉核验，见 [需求人配方](demand-person.md)。 |
| `CONTRACT_DEMAND_PERSON_DEPARTMENT` / `demandDepartmentIds` | `string` / `array<string>` | 需求部门 OPENID；Agent 优先传非空数组。 |
| `CONTRACT_AMOUNT` / `contractAmount` | `array` | 恰好两个元素 `[start,end]`；包含两端，元素为 number 或 `null`，legacy 单边可用 `null`，至少一端必须是有限 number。金额与实际币种配合，详见下文边界验证。 |
| `CONTRACT_CURRENCY` / `contractCurrency` | `string` / `integer` / `array` | 支持英文币种码或数字 code；数组元素为 string 或 integer，且不得为空。无效币种在 legacy 路径返回空结果。 |
| `CONTRACT_LEGAL_ENTITY_NAME_PRECISE` / `contractLegalEntityNamePrecise` | `string` | 当前公开 filter 契约为我方主体名称 TEXT；无 label 分支按名称包含查询，不能称为主体 ID 精确筛选，也不把 `mdm legal` 返回 ID 填入此字符串。 |
| `CONTRACT_TRADING_PARTY` / `contractTradingParty` | `string` | 交易方名称关键词；不是 ID 数组。带 label 的 legacy 分支按名称集合查询，也不代表按 ID 精确筛选，见 [交易方配方](trading-party.md)。 |
| `CONTRACT_TRADING_PARTY_ID` | `array<string>` | 已确认的交易方主体 ID 非空数组；可从 `mdm vendor list` 候选 `id` 或合同 `counter_party_list[].counter_party_id` 取得，不用编码或我方主体 ID。与需求人组合已实测，见 [交易方配方](trading-party.md)。 |
| `CONTRACT_ASSOCIATION_PARTY` / `contractAssociationParty` | `integer` | 关联交易枚举 code；使用字段元数据返回的 `value_scopes[].value`。 |
| `CONTRACT_ANTI_DATED` / `antiDated` | `integer` | JSON integer `0` 或 `1`。 |
| `CONTRACT_ANTI_DATED_MULTI` / `antiDatedMulti` | `array<integer>` | 非空倒签类型枚举数组；值取字段元数据。 |
| `CONTRACT_SCAN_EXISTS` / `contractScanExists` | `integer` | JSON integer `0` 或 `1`。 |
| `CONTRACT_OWNER_EQUALS_SUBMITTER` / `ownerEqualsSubmitter` | `integer` | JSON integer `0` 或 `1`。 |
| `CONTRACT_SHARE_TO_ME` / `contractShareToMe` | `integer` | JSON integer `0` 或 `1`。 |
| `CONTRACT_STANDARD_FRAMEWORK_AGREEMENT` / `contractFrameworkParentFlag` | `integer` | JSON integer `0` 或 `1`。 |
| `CONTRACT_SIGNED_DATE` / `signedDate` | `array` | 恰好两个毫秒时间戳或 `null`：`[start,end]`；legacy 单边可用 `null`。 |
| `CONTRACT_FIRST_SIGNER` / `contractFirstSigner` | `integer` | `0` 未指定、`1` 对方先签、`2` 我方先签。 |
| `CONTRACT_SIGN_PLATFORM_TYPE` / `contractSignPlatformType` | `integer` / `array<integer>` | `0` 无、`1` 纸质签、`2` 电子牵、`3` DocuSign、`4` e签宝、`5` 法大大；以租户字段元数据中启用的值为准。 |
| `CONTRACT_SEAL_NUMBER` / `contractSealNumber` | `integer` / `array<integer>` | 盖章份数，不是印章编号；Agent 优先传非空数组，例如 `[2]`。 |
| `CONTRACT_TEMPLATE_MULTI_CONTRACT` / `contractTemplateMultiContract` | `integer` | JSON integer `0` 或 `1`。 |
| `CONTRACT_EFFECTIVE_STATUS` / `contractEffectiveStatus` | `array<integer>` | 非空数组：`0` 未生效、`1` 生效中、`2` 已失效。 |
| `CONTRACT_NODE_OCR_USAGE` / `contractNodeOcrUsage` | `integer` | `0` 全部使用或不涉及 OCR、`1` 部分未使用 OCR。 |
| `CONTRACT_OWNER_AUTHORIZED_EMPLOYEES` / `ownerAuthorizedEmployeeId` | `string` / `array<string>` | 授权人的飞书 `user_id`；Agent 优先传非空数组。 |
| `CONTRACT_PRE_AUTHORIZED_EMPLOYEES` / `preAuthorizedEmployeeId` | `string` / `array<string>` | 预授权人的飞书 `user_id`；Agent 优先传非空数组。 |

上述 `0/1` 字段不传 JSON boolean `true/false`；当前 legacy processor 按 `Integer` 消费这些值。

自定义字段：

| 枚举名 / PC 字段名 | `search_value` JSON 类型 | 额外约束 |
| --- | --- | --- |
| `CONTRACT_FORM_FIELDS_TEXT` / `contractFormFieldsText` | `string` | 非空文本；必传 `filter_unique_key`。 |
| `CONTRACT_FORM_FIELDS_DOUBLE` / `contractFormFieldsDouble` | `array` | 恰好两个 number 或 `null`；legacy 单边可为 `null`；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_DATE` / `contractFormFieldsDate` | `array` | 恰好两个毫秒时间戳或 `null`，不传日期字符串；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_OPTION` / `contractFormFieldsOption` | `string` / `array<string>` / `array<integer>` | 类型取决于元数据：`OPTION_LABEL_ARRAY` 使用普通选项的字符串 value 数组；`CURRENCY_ID_ARRAY` 使用自定义币种的整数 code 数组。Agent 优先用实际非空数组；必传唯一 key，不能只凭本枚举推断类型。 |
| `CONTRACT_FORM_FIELDS_OPTION_ID` / `contractFormFieldsOptionId` | `string` / `integer` / `array` | legacy 选项原始 value/id；数组元素为 string 或 integer。新请求不要手选本字段，优先使用字段元数据返回的枚举、语义类型及 value 数组；普通选项通常为 `CONTRACT_FORM_FIELDS_OPTION + OPTION_LABEL_ARRAY`；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT` / `contractFormFieldsEmployeeDepartment` | `string` / `array<string>` | legacy 人员/部门显示名搜索；新请求不要手选本字段，使用下一行的外部 ID 形式；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT_ID` / `contractFormFieldsEmployeeDepartmentId` | `string` / `array<string>` | 人员字段传飞书 `user_id`，部门字段传部门 OPENID；Agent 优先传非空数组；必传唯一 key。 |
| `CONTRACT_FORM_FIELDS_CURRENCY` / `contractFormFieldsCurrency` | `integer` / `array<integer>` | 保留的 legacy 名称；新请求不要手选本字段，按元数据使用 `CONTRACT_FORM_FIELDS_OPTION + CURRENCY_ID_ARRAY` 及整数 code 数组；必传唯一 key。 |
| `CONTRACT_HYPERLINK` / `contractHyperlink` | `object` | legacy 至少读取非空 string `url` / `title` 之一；Agent 应同时传两者且不附加其他键；必传唯一 key。 |

完整的语义类型到 JSON 形状、空候选处理、元数据响应到请求示例，见 [搜索字段发现](search-filter-fields.md)。实际字段的 `search_value_type`、值说明与 `value_scopes[].value` 决定取值；`examples[].request_fragment` 提供结构，不能把示例业务值直接当成真实候选。2026-09-20 生产元数据已确认自定义币种使用整数 value；选定样本的整数及 label 查询均为 0 条，不能宣称命中对照通过或两者等价。

## 范围、日期与部门角色

固定金额 `CONTRACT_AMOUNT` 的范围包含端点：`[a,b]` 表示 `a <= 金额 <= b`，`[a,null]` 表示至少 a，`[null,b]` 表示不超过 b，`[a,a]` 表示等于 a。必须同时保留用户要求的币种，不把不同币种的数字直接比较。用户要求“严格大于 a”时，不能直接使用 `[a,null]`，也不能在未确认金额精度与接口能力时擅自改为 `[a+0.01,null]`；按 [组合规则](search-combinations.md) 处理无法直接表达的条件。

2026-09-20 生产只读对照以 18 条已查完的正文结果为基线，选取其中真实正金额 a 和币种：同正文及币种下，`[a,a]` 返回 5 条、`[a+1,a+1]` 返回 0 条、`[a,null]` 返回 16 条、`[null,a]` 返回 6 条，四组均已查完，合同 ID 集合与基线按实际金额/币种筛出的结果完全一致。这验证了固定金额的闭区间和单边 null；自定义数字处理器虽有相同的本地 gte/lte 逻辑，尚不能据此宣称所有自定义数值字段已完成生产验证。

日期按业务角色区分：提交时间、归档时间、签订时间、签约日期及自定义日期不能相互替代。`DATETIME_STRING_RANGE` 按 `request_paths` 写两个 `YYYY-MM-DD HH:mm:ss` 字符串属性；`MILLIS_RANGE_ARRAY` 则传两个毫秒时间戳。复用已确认业务时区，将“本月”等相对时间换算为具体范围后说明；不要从客户端时区推断服务端时区。

页面“合同申请日期”已用实际数据验证，对应 `combine_condition.submited_time_start/end`，不放入 `filter_units`。页面抓包中的 `CONTRACT_SUBMITTED_TIME/submittedTime` 与 MCP 元数据的 `CONTRACT_SUBMITED_TIME/submitedTime` 拼写不同，不能直接复制；请求按元数据给出的 combine 路径构造。日期与状态组合、毫秒边界和计数范围见 [申请日期实测](submission-date.md)，不将此项视为自定义日期验证。

本地自定义日期处理器会把输入毫秒格式化为 `yyyy-MM-dd` 后执行包含端点的范围查询，且格式化使用服务进程默认时区。这是源码线索，当前尚缺真实非空字段样本证明生产日内边界；不能仅因输入是毫秒就承诺小时级过滤，也不能将此粒度推及所有日期字段。生产元数据已确认自定义日期的输入类型为 `MILLIS_RANGE_ARRAY`。

部门需要同时确认业务角色和值来源：

| 查询方式 | 已确认信息 | 不应作出的推断 |
| --- | --- | --- |
| `condition_units.CONTRACT_DEPARTMENT_NAME` | 页面名称关键词；本地处理逻辑按部门名解析，再查 `contracts.baseInfo.departmentId` | 名字命中不证明所有部门角色等价，也不保证唯一部门 |
| `filter_units.DEPARTMENT_AUTHORITY` | 生产元数据展示“申请部门”，要求外部部门 OPENID；本地索引路径为 `contracts.authority.departmentIds` | 不能默认作为上一行名称查询的精确 ID 版本 |
| `filter_units.CONTRACT_DEMAND_PERSON_DEPARTMENT` | 合同需求人部门，使用元数据要求的外部部门 OPENID | 不用申请部门替代需求部门 |
| 自定义部门字段 | 真实唯一 key + `CONTRACT_FORM_FIELDS_EMPLOYEE_DEPARTMENT_ID`，类型为 `DEPARTMENT_OPEN_ID_ARRAY` | 同名自定义“所属部门”不等于系统名称字段 |

上表区分了生产元数据与本地索引依据；尚无足够的生产角色对照样本证明前两种查询等价。需要精确部门角色而上下文无法区分时，用业务含义澄清，不让用户选择内部枚举。`value_scopes` 为空时也不改用其他部门角色；先检查已有可靠外部 ID，再补齐必要来源。

## 排序与分页

`sort_type` 枚举：

| 值 | 含义 | 值 | 含义 |
| --- | --- | --- | --- |
| `SUBMITTED_TIME` | 申请时间 | `SIGNED_TIME` | 签订时间 |
| `CREATE_TIME` | 创建时间 | `ARCHIVED_TIME` | 归档时间 |
| `CONTRACT_NUMBER` | 合同编号 | `START_DATE` | 合同开始日期 |
| `CONTRACT_STATUS` | 合同状态 | `END_DATE` | 合同结束日期 |
| `AMOUNT` | 合同金额 | `ARCHIVE_NUMBER` | 归档编号 |

- `order` 只允许 `ASC`、`DESC`，默认 `DESC`。
- `sort` 是旧兼容字段；仅值 `asc` 表示升序，且排序字段固定为提交时间。
- 首次查询不传 `page_token`；后续直接使用响应 token，不自行计算。
- `page_size` 不保证等于返回合同条目数，组展开后可能更多。生产申请日期样本按 `SUBMITTED_TIME/DESC` 请求后，展平条目的 `submitted_time` 并非全局倒序；不能把第一页称为全部条目中精确的“最新 N 条”。按原响应分页；确需条目全局排序时须先完整取回再排序，并说明这是本地展示排序。

## 枚举与约束

合同状态 code：

| code | 含义 | code | 含义 |
| --- | --- | --- | --- |
| `0` | 正编辑/草稿 | `9` | 已归档 |
| `1` | 已作废 | `10` | 变更中 |
| `2` | 已撤回 | `11` | 已变更 |
| `3` | 审批中 | `12` | 我方已签约，归入签订中 |
| `4` | 已拒绝 | `13` | 对方已签约，归入签订中 |
| `5` | 审批已通过 | `16` | 终止中 |
| `6` | 签订中 | `17` | 已终止 |
| `7` | 已签订 | `19` | 审批流程被干预中止，PC 按已作废展示 |
| `8` | 归档中 | `23` | 已完成 |

PC 聚合状态：已作废=`1,19`，签订中=`6,12,13`，其余使用对应单个 code。

其他硬约束：

- `condition_units` 和 `filter_units` 各最多 50 项。
- 历史 `combine_condition.demand_employee_ids` 的本地处理只允许一个元素，但字段文案/ID 转换存在冲突，严格 MCP 白名单也未包含它；不能由此推荐该路径。需求人新查询按元数据走 `filter_units.CONTRACT_DEMAND_PERSON`。
- 自定义 filter 字段必须传 `filter_unique_key`。
- `DEPARTMENT_AUTHORITY` 传部门 ID 数组；按部门名称搜索使用 `condition_units.CONTRACT_DEPARTMENT_NAME`。
- 日期型 filter 传毫秒时间戳范围，不传 `YYYY-MM-DD` 字符串。
- `0/1` filter 传 JSON integer，不传 boolean；范围数组固定两个位置，单边范围只在当前 legacy profile 使用 `null`。
- 固定/自定义枚举优先使用字段元数据的 `value_scopes[].value`；不要凭展示文案反推 code。
- 不传 `search_tab_code` 时默认 `0`；非法页签、排序字段或排序方向返回参数错误。

## 动态字段发现

固定 `condition_units.CONTRACT_SUBMIT_NAME` / `CONTRACT_DEPARTMENT_NAME` 在用户明确要求姓名或部门名称关键词时无需字段发现或候选查询，可直接按主 Skill 场景 6、7 执行。指定具体申请人或部门时，遵循主总纲优先候选消歧与该角色已支持的 ID 筛选。

以下流程适用于需要元数据的筛选字段：按非固定字段、自定义字段、选项、人员或部门筛选字段搜索前，使用对应 MCP `list-contract-search-filter-fields` 的命令：

```bash
contract-cli contract search-fields --keyword "项目区域" --profile contract --as user --output json
```

已知字段展示名时传 `--keyword`，按展示名字面量模糊匹配；只有用户明确要求完整字段清单时才省略。接口一次返回匹配字段，固定中文说明，不需要分页。

- 根据返回的 `request_location`、`request_paths` 放置条件，可能是 `combine_condition`、`condition_units` 或 `filter_units`；不能仅凭字段类型猜位置。
- 自定义字段同时使用真实 `search_field` 与 `filter_unique_key`，按 `search_value_type`、`value_description`、`usage_hint`、`value_scopes` 和 `examples` 构造值，保持原始 JSON 类型。
- 发现与搜索使用相同配置、租户和用户身份。多个同名结果需确认目标字段；空字段结果不等于没有符合条件的合同，不猜 key、不删除原条件。
- 本次开发版新增 `contract search-fields`，旧 CLI 1.8.6 不支持；顶层人员与部门查询已新增：见 [人员 Skill](../../contract-cli-employee/SKILL.md) 的 `employee list` 和 [部门 Skill](../../contract-cli-department/SKILL.md) 的 `department list`。要求 ID 的字段先查询真实候选，不把名称关键词当作 ID。

命令参数、响应解读及后续搜索步骤见 [搜索字段发现](search-filter-fields.md)。

## 示例

按合同编号关键词搜索：

```bash
contract-cli contract search --profile contract --as user --data '{"contract_number":"CT2026","page_size":10}'
```

关键词与结构化条件组合，保存为 `search-user.json`：

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "sort_type": "SUBMITTED_TIME",
  "order": "DESC",
  "condition_units": [
    {
      "search_field": "CONTRACT_NAME_FIELD",
      "search_value": "采购",
      "union_type": "SHOULD"
    }
  ],
  "filter_units": [
    {
      "search_field": "CONTRACT_AMOUNT",
      "search_value": [1000, 5000],
      "union_type": "MUST"
    },
    {
      "search_field": "CONTRACT_CURRENCY",
      "search_value": ["CNY", "USD"],
      "union_type": "MUST"
    }
  ]
}
```

```bash
contract-cli contract search --profile contract --as user --input-file search-user.json
```

按自定义日期字段筛选：先用真实元数据替换下例的 `custom_sign_date`，它只是占位 key，不可原样作为租户字段使用。

```json
{
  "filter_units": [
    {
      "search_field": "CONTRACT_FORM_FIELDS_DATE",
      "filter_unique_key": "custom_sign_date",
      "search_value": [1704067200000, 1706659200000],
      "union_type": "MUST"
    }
  ]
}
```

## 响应摘要

- 先检查业务 `code` / `success`；本次开发版 user 搜索遇非零 code、`success=false` 或缺失 code 返回失败退出码并保留响应。旧 CLI 1.8.6 业务失败仍可能返回退出码 0，不能将错误响应解释为空结果。
- 当前生产默认常用路径：`data.items[]`、`data.has_more`、`data.page_token`。`has_more=false` 时 token 仍可能非空，以 has_more 决定是否翻页。
- 扩展响应可能带 `data.pagination`；若 `pagination.unit=CONTRACT_GROUP`，其中 `returned_units/total_units` 是合同组数，不能与 `returned_items` 混用。未返回统计时不推算总数；完整分页后可分别按 `contract_id` 和有效 `group_id` 去重统计已取回条目与组，组 ID 缺失时不报完整组数。实测 103 组可对应 110 条合同，见 [计数说明](submission-date.md#结果计数与排序)。
- `contract_status_display_name`、`contract_status_display_name_source`、`contract_status_display_policy_version` 等扩展字段在当前生产默认样例中未提供，不应视为必返。
- 完整合同字段见 [contract-response-fields.md](../../contract-cli-contract/references/contract-response-fields.md)。
- 返回的是组展开后的条目，不能从某组命中推断每条都满足人员及状态。姓名＋变更中样本含 2 条审批中成员，见 [组与条目](applicant-status.md#结果按合同组解释)；读取每条实际状态，不能统一改写成请求中的状态。
