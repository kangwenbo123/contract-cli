# 搜索总纲与接口缺口复核

日期：2026-09-20。依据：当前开发分支命令、MCP 契约与本地 CLM 源码、[逐项生产验证](contract-search-validation.md)。总纲分析复用原证据；另按用户追加要求，以当前生产 user 授权实测我方主体列表和详情各一次，不将源码分支当作已部署并实测的通用合同搜索能力。

## 总纲的最终顺序

1. **先确认授权与身份。** 复用有效授权，未知状态才查 `auth status`；user/app 分别选择接口与参考文档。授权状态不等于当前业务用户身份，不能从 Token 或 profile 名猜本人 user_id。app 不具备 user MCP 的 filter/condition/search-fields 能力。
2. **识别角色与字段，按语义分流。** 明确人员、部门、交易方或我方主体时，先候选消歧，优先该角色已经支持的精确筛选；明确固定字段时保留字段；疑似或明确自定义字段先发现元数据；只有未指定字段的纯关键词查询才默认全部关键词。
3. **取得真实值。** 人员/部门/主体 ID 由 Agent 查询，不默认要求用户提供；自定义字段按 field_type、业务含义、位置、key、类型及候选共同解析。上下文与元数据已经清楚时不重复询问；同名字段、同名对象或角色仍有歧义时只问业务问题。
4. **组合全部条件。** user 通常以 filter/MUST 限定对象，但状态和申请日期仍走 combine；自定义字段按实际元数据位置。追加条件保留旧值、改变条件清除分页；不能把 AND 改成 OR，或将精确需求降级成名称关键词。
5. **核对响应与计数。** 区分错误与零结果、分页与总量、合同组与成员条目。需要同一份合同满足全部条件时核对条目，不能将组层面独立条件匹配等同于条目级 AND。

用户提出的方向成立，但第二、四点需要上述边界。主文件保留九个三段式示例作为按需示例库；申请人姓名场景改为用户明确要求“姓名关键词”时使用，避免它覆盖“明确对象优先精确”的总原则。

## 基础接口入口：当前开发版已具备

| 步骤 | 已有 CLI | 身份与说明 |
| --- | --- | --- |
| 授权状态及官方授权 | `auth status`、现有 auth 流程 | 分 user/app，复用原配置；status 不刷新 Token，也不提供可靠业务 user_id |
| 人员候选 | `employee list --name` / `--department-id` | user-only，返回 user_id，已有独立 Skill 与生产验证 |
| 部门候选 | `department list --name` / `--department-id` | user-only，开放部门 ID；按角色进入合同筛选 |
| 交易方候选与详情 | `mdm vendor list/get` | user/app；user 名称候选已与精确交易方搜索做实测闭环 |
| 我方主体候选与详情 | `mdm legal list/get` | user/app；当前生产 user 列表/详情均 code=0、success=true，完整一个候选且 ID 与已知合同一致；不代表合同搜索已有通用精确 ID 契约 |
| 搜索字段发现 | `contract search-fields` | user-only，基础与自定义字段；按展示名 keyword 定向发现 |
| 合同搜索 | `contract search`、`contract search-v2` | user MCP、app V1、app V2 分开说明，不能混用请求体 |

因此无需再新增一套“人员、部门、交易方、我方列表”接口。这里指当前开发分支及包含相应命令的开发二进制，不表示正式 CLI 1.8.6 已拥有后来新增的能力。

## 仍需补齐：优先扩展现有接口契约

| 优先级 / 条件 | 缺口 | 建议落点与验收 |
| --- | --- | --- |
| P0：落实我方主体精确筛选 | 候选 ID 已有，但公开 filter 仅 `CONTRACT_LEGAL_ENTITY_NAME_PRECISE/TEXT`。本地有依赖 label 的 legacy ID 分支及特殊租户编码分支，严格契约又不接受 label；当前无已验证通用配方 | 扩展现有 MCP search 的正式精确主体条件与 search-fields 元数据，统一候选 ID 类型；用正确主体、另一真实主体和同名主体做对照。无需新建我方列表命令 |
| P0：字段发现能完整指导 Agent | 交易方 `_ID` 搜索已支持且实测，但按“交易方”发现只返回名称 TEXT；Agent 仍依赖 Skill 的额外固定规则 | 在现有 search-fields 返回精确 ID 字段、字符串数组类型、候选来源、角色及正确片段，禁止把名称示例冒充 ID 示例 |
| P1：页面全部关键词完全对齐 | MCP 未声明或传递 PC 的 allConditionUnits 等价语义，影响人员候选和表单人员扩展；毛鹏样本对应不能证明全场景等价 | 由现有 search 明确页面全部关键词的语义，必要时新增正式模式/参数；不能只透传未知字段后用 code=0 认定支持 |
| P1：可靠表达复杂 AND | 多词正文 AND 已有失败反例；部分人员姓名并列会扩大/忽略条件；不同 filter 还可能由同组不同合同满足 | 在 search 明确分组布尔条件与同一合同匹配语义，或提供可验证的匹配成员信息；以正反结果集合验收，不靠 CLI 拼接更多 MUST |
| 按产品范围决定：归属人/创建人也要精确限定 | 有名称关键词，普通精确 ID filter 白名单尚无对应角色 | 若“人员精确”涵盖这些角色，在现有 search 补正式角色字段及元数据；不能借用申请人/需求人/授权人员字段 |
| 按产品范围决定：“我申请的”无需身份澄清 | auth status 不返回可靠的当前业务 user_id；人员候选查询不能证明哪个是授权本人 | 先确认可复用的当前身份接口；若无，新增最小当前身份查询或扩展授权状态的业务身份契约。未发现现成已发布 MCP whoami，不能承诺只加 CLI 包装即可 |

app 如果也要求与 user/MCP 完全同等的搜索能力，需要独立的 app 能力扩展设计；这不属于缺少某个候选列表命令，不能以切换身份代替授权和接口设计。

## 证据边界与源码入口

- CLI 路由及候选：`internal/openplatform/contract/service.go`、`directory.go`、`mcp_specs.go`、`internal/openplatform/mdmvendor/service.go`、`internal/openplatform/entity/service.go`。
- MCP 本地白名单和严格类型：CLM `McpContractSearchFilterPolicy.java`、`McpContractSearchRequestDTO.java`；我方主体分支在 `ContractLegalEntityQueryBuilderProcessor.java`。无 label 名称包含与 legacy label/ID 分支不能合并成同一个公开契约。
- PC 全部标记在 `CompoundSearchDTO`，MCP DTO/转换未提供对应入口；生产全部关键词与组合证据见验证记录第 11–15 节。
- 多词 AND、批号关键词覆盖、日期精度及未取得阳性样本的自定义字段仍遵守现有验证记录；本轮没有因流程更清晰就将这些后端限制标为已解决。
- 本文是接口缺口与实施落点分析，未修改上述服务端接口，也未发布 CLI。
