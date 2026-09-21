# contract search-v2 Parameters

本页专用于 app 身份下的合同搜索 V2。

- 命令：`contract-cli contract search-v2 --profile contract --as app`
- 接口：`POST /open-apis/contract/v1/contracts/searchV2`
- 身份：仅 `app`
- 请求体：JSON 必填，`--input-file` 与 `--data` 二选一
- 官方 OpenAPI：[搜索合同-V2](https://docs.qfei.cn/460498306e0.md)
- 实现基线：CLM `PlatformContractSearchV2OpenService`、`ContractSearchDTO`

## 目录

- [适用场景](#适用场景)
- [CLI 参数映射](#cli-参数映射)
- [请求体字段](#请求体字段)
- [编号搜索与条件回退](#编号搜索与条件回退)
- [分页语义](#分页语义)
- [枚举与约束](#枚举与约束)
- [示例](#示例)
- [响应行为](#响应行为)

## 适用场景

使用 V2 的主要理由是顶层 `contract_number` 需要 ES 模糊查询、返回多条结果并分页。它不是 user MCP 搜索，也不是一套新的结构化筛选协议。

- 需要 app 合同编号精确查单条：使用 `contract search --as app`。
- 需要 app 合同编号模糊查询或分号批量查询：使用 `contract search-v2 --as app`。
- 需要 `condition_units`、`filter_units`、页签和 MCP 排序：使用 `contract search --as user`。

## CLI 参数映射

| 参数 | 请求位置 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- | --- |
| `--as app` | 本地身份 | enum | 必须是 app | user 身份会被 CLI 拒绝。 |
| `--user-id-type` | `$query.user_id_type` | string | 服务端必填；CLI 默认 `user_id` | 支持 `user_id`、`union_id`。 |
| `--user-id` | `$query.user_id` | string | 可选 | CLI 通用 query 透传参数。 |
| `--input-file` | `$body` | JSON file | 二选一必填 | 推荐用于复杂请求。 |
| `--data` | `$body` | JSON string | 二选一必填 | 适合简单编号搜索。 |
| `--profile` | 本地上下文 | string | 可选 | 不传时使用当前 profile。 |
| `--output` | CLI 输出 | enum | 可选 | `json`、`yaml`、`table`，默认 `json`。 |
| `--raw` | CLI 输出 | boolean | 可选 | 原样输出服务端响应。 |

`search-v2` 不提供 `--contract-number`、`--page-size`、`--page-token` 快捷 flag；这些字段必须写入 JSON body。

## 请求体字段

请求体对象必填，但对象内部没有统一业务必填字段。

| JSON 路径 | 类型 | 必填性 | 默认值/约束 | 说明 |
| --- | --- | --- | --- | --- |
| `page_size` | integer | 可选 | 默认 `10`；服务端要求 `1..100` | V2 入口统一校验。 |
| `page_token` | string | 可选 | 首次不传 | 语义取决于最终搜索分支，见下文。 |
| `contract_number` | string | 可选 | 非空关键词 | ES 模糊查询；包含 `;` 或 `；` 时进入精确批量编号逻辑。 |
| `combine_condition` | object | 可选 | 旧组合条件 | 只有顶层 `contract_number` 为空时生效。 |
| `logic_search` | object | 可选 | 旧逻辑条件 | 只有顶层 `contract_number` 和 `combine_condition` 都为空时使用。 |
| `lang` | string | 可选 | 后端默认语言 | 例如 `zh-CN`、`en-US`。 |
| `sort` | string | 可选 | 旧排序字段 | 编号 ES 搜索会透传。 |
| `condition_units` | array<object> | 可选但无效 | 明确忽略 | V2 不支持 MCP 关键词单元。 |
| `filter_units` | array<object> | 可选但无效 | 明确忽略 | V2 不支持 MCP 结构化筛选单元。 |

`combine_condition` 与 `logic_search` 的公开字段、枚举和嵌套结构与 app V1 相同，完整定义见 [search-app-parameters.md](search-app-parameters.md)。常用字段包括：

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `combine_condition.user_id` | string | 可选 | 与 `permission` 配合限定权限范围。 |
| `combine_condition.permission` | integer | 可选 | `0` 全部、`1` 我的可见、`2` 我申请的。 |
| `combine_condition.contract_status` | integer | 可选 | 单个状态。 |
| `combine_condition.contract_status_in` | string | 可选 | 英文逗号分隔的状态 code。 |
| `combine_condition.pay_type` | integer | 可选 | `0..4`，含义见 app V1 参数文档。 |
| `combine_condition.contract_category_name` | string | 可选 | 合同分类名称。 |
| `combine_condition.contract_name` | string | 可选 | 合同名称。 |
| `combine_condition.contract_number` | string | 可选 | 组合条件编号，和顶层 V2 模糊编号不是同一路径。 |
| `combine_condition.archive_number` | string | 可选 | 归档编号。 |
| `combine_condition.submit_user_id` | string | 可选 | 提交人用户 ID。 |
| `combine_condition.*_time_start/end` | string | 可选 | 时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.form` | array<object> | 可选 | 字段值匹配；子项要求 `attribute_name/value/module_name`。 |
| `logic_search.logic_type` | integer | 父对象存在时必填 | `0` AND、`1` OR、`2` 叶子。 |
| `logic_search.children` | array<object> | AND/OR 节点使用 | 当前 CLM DTO 使用递归对象数组。 |
| `logic_search.data` | object | 叶子节点必填 | 包含 `field_name`、`operator`、`field_value`。 |

## 编号搜索与条件回退

顶层 `contract_number` 非空时：

1. 构造 `CONTRACT_NUMBER_FIELD` 的 `SHOULD` 条件执行 ES 模糊查询。
2. 丢弃 `combine_condition`、`logic_search`、`condition_units` 和 `filter_units`。
3. 使用全数据权限搜索并重新从数据库组装合同详情。
4. 未命中时成功返回空列表，不返回 V1 的 `110107`。

顶层 `contract_number` 为空时：

1. V2 委托 V1 的 `ContractOpenSearchServiceSelector`。
2. `combine_condition`、`logic_search` 和对应旧分页行为继续生效。
3. `condition_units` 和 `filter_units` 仍然被忽略。

## 分页语义

- 顶层编号 ES 搜索：`page_token` 是从 `0` 开始的十进制页码；下一页响应为当前页码加一。
- 无顶层编号且最终走 ES 的组合搜索：token 同样按页码处理。
- 无顶层编号且委托旧 RDS 组合搜索：token 可能是最后一条合同 ID 游标，必须原样回传，不能转成页码。
- ES 页码分支要求 `page_index * page_size` 不得进入 10000 条之外；实际校验保证下一页结束偏移不超过 `10000`。
- `page_token` 是否必须为纯数字取决于实际搜索分支；不要跨不同请求条件复用 token。

## 枚举与约束

- `page_size` 必须大于 `0` 且小于等于 `100`；官方示例中的 `0` 与当前实现冲突，不可照抄。
- 顶层普通合同编号按 ES 模糊语义搜索。
- 顶层编号中出现 `;` 或 `；` 时，底层合同搜索进入编号精确批量逻辑，并把实际 `page_size` 上限收敛为 `50`。
- 顶层编号非空时忽略所有其他筛选条件；需要多个条件稳定组合时不要使用 app V2 顶层编号。
- `combine_condition` 存在时优先于 `logic_search`。
- `condition_units` 和 `filter_units` 会被忽略；不要因为 DTO 能反序列化字段就认为 V2 支持它们。
- V2 与 V1 返回同一个 `ContractSearchVO` 外形，但编号查询的空结果和分页行为不同。

## 示例

合同编号模糊搜索：

```bash
contract-cli contract search-v2 --profile contract --as app --data '{"contract_number":"CT2026","page_size":20}'
```

合同编号精确批量查询，保存为 `search-v2.json`：

```json
{
  "contract_number": "CT-001;CT-002；CT-003",
  "page_size": 50
}
```

```bash
contract-cli contract search-v2 --profile contract --as app --input-file search-v2.json
```

使用旧组合条件回退：

```json
{
  "page_size": 20,
  "combine_condition": {
    "user_id": "ou_xxx",
    "permission": 1,
    "contract_status_in": "3,6,9",
    "contract_name": "采购合同"
  }
}
```

注意：以下请求不会按金额筛选，因为 V2 明确忽略 `filter_units`：

```json
{
  "contract_number": "CT2026",
  "filter_units": [
    {
      "search_field": "CONTRACT_AMOUNT",
      "search_value": [1000, 5000]
    }
  ]
}
```

## 响应行为

- 常用路径：`data.items[]`、`data.has_more`、`data.page_token`。
- 顶层编号查询可能返回多条；未命中是成功空列表。
- 翻页时只在 `has_more=true` 时回传响应中的 `page_token`。
- 完整合同字段见 [contract-response-fields.md](../../contract-cli-contract/references/contract-response-fields.md)。
