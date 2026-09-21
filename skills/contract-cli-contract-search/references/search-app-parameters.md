# contract search App Parameters

本页专用于 app 身份下的旧版合同搜索。

- 命令：`contract-cli contract search --profile contract --as app`
- 接口：`POST /open-apis/contract/v1/contracts/search`
- 身份：仅本页语义使用 `app`
- 官方 OpenAPI：[搜索合同](https://docs.qfei.cn/366415247e0.md)
- 实现基线：CLM `ContractOpenPlatformController`、`ContractSearchDTO`、`ContractOpenSearchServiceSelector`

## 目录

- [CLI 参数映射](#cli-参数映射)
- [请求体字段](#请求体字段)
- [组合条件](#组合条件)
- [逻辑搜索](#逻辑搜索)
- [枚举与约束](#枚举与约束)
- [示例](#示例)
- [响应行为](#响应行为)

## CLI 参数映射

| 参数 | 请求位置 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- | --- |
| `--as app` | 本地身份 | enum | 建议显式传 | 使用 app token，并固定路由到本页接口。 |
| `--user-id-type` | `$query.user_id_type` | string | 服务端必填；CLI 默认 `user_id` | 支持 `user_id`、`union_id`。 |
| `--user-id` | `$query.user_id` | string | 可选 | CLI 通用透传参数；搜索请求体中的操作者字段仍按接口字段填写。 |
| `--contract-number` | `$body.contract_number` | string | 可选 | 覆盖请求体同名字段。V1 按合同编号精确查询。 |
| `--page-size` | `$body.page_size` | integer | 可选 | 覆盖请求体同名字段。 |
| `--page-token` | `$body.page_token` | string | 可选 | 覆盖请求体同名字段。 |
| `--input-file` | `$body` | JSON file | 与 `--data` 互斥 | 复杂条件推荐使用。 |
| `--data` | `$body` | JSON string | 与 `--input-file` 互斥 | 适合简单查询。 |
| `--profile` | 本地上下文 | string | 可选 | 不传时使用当前 profile。 |
| `--output` | CLI 输出 | enum | 可选 | `json`、`yaml`、`table`，默认 `json`。 |
| `--raw` | CLI 输出 | boolean | 可选 | 原样输出服务端响应。 |

## 请求体字段

请求体所有顶层字段均可选；不传请求体时 CLI 会发送空 JSON 对象。

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `page_size` | integer | 可选，默认 `10` | 分页大小。顶层编号精确查询最多一条，此时分页字段不影响结果数量。 |
| `page_token` | string | 可选 | 首次不传；组合查询翻页时原样使用响应中的 `page_token`。其格式由实际搜索分支决定。 |
| `contract_number` | string | 可选 | 顶层合同编号，V1 执行精确查询；一旦非空，后端优先选择编号搜索并忽略其他条件。 |
| `combine_condition` | object | 可选 | 旧组合条件；存在时优先于 `logic_search`。 |
| `logic_search` | object | 可选 | 逻辑条件树；与 `combine_condition` 二选一。 |
| `lang` | string | 可选 | 返回语言；由后端语言工具解析。 |
| `sort` | string | 可选 | 旧排序字段，具体效果取决于组合搜索分支。 |

## 组合条件

以下表格以公开 OpenAPI 字段为准；所有字段在 `combine_condition` 存在时仍是可选。

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `combine_condition.user_id` | string | 可选 | 用户 ID，与 `permission` 配合限定权限范围。 |
| `combine_condition.permission` | integer | 可选 | `0` 全部合同；`1` 我的可见合同；`2` 申请的合同。 |
| `combine_condition.contract_status` | integer | 可选 | 单个合同状态。 |
| `combine_condition.contract_status_in` | string | 可选 | 多个状态 code，使用英文逗号分隔，例如 `"3,6,9"`。 |
| `combine_condition.pay_type` | integer | 可选 | 收支类型：`0` 未知、`1` 收入、`2` 支出、`3` 收入支出、`4` 无金额。 |
| `combine_condition.contract_category_name` | string | 可选 | 合同分类名称。 |
| `combine_condition.contract_category_abbreviation` | string | 可选 | 合同分类缩写。 |
| `combine_condition.contract_name` | string | 可选 | 合同名称。 |
| `combine_condition.contract_number` | string | 可选 | 组合条件中的合同编号；不要与顶层精确编号查询混淆。 |
| `combine_condition.archive_number` | string | 可选 | 归档编号。 |
| `combine_condition.submit_user_id` | string | 可选 | 提交人用户 ID。 |
| `combine_condition.create_time_start/end` | string | 可选 | 创建时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.update_time_start/end` | string | 可选 | 更新时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.submited_time_start/end` | string | 可选 | 提交时间范围；字段名按接口保留 `submited` 拼写。 |
| `combine_condition.archived_time_start/end` | string | 可选 | 归档时间范围，格式 `YYYY-MM-DD HH:mm:ss`。 |
| `combine_condition.form` | array<object> | 可选 | 字段值匹配；公开接口当前说明仅支持关联前置单据搜索。 |

公开 OpenAPI 还声明了 `combine_condition.owner_user_id`，但当前 CLM `ContractSearchDTO` 没有对应反序列化字段。不要依赖该字段生效；需要按归属人查询时应先由服务端接口维护方确认兼容状态。

`form[]` 子项：

| 字段 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `attribute_name` | string | 子项必填 | 字段名称。 |
| `attribute_value` | string | 子项必填 | 字段值。 |
| `module_name` | string | 子项必填 | 模块名称。 |
| `operator` | integer | 可选 | 当前 DTO 支持操作符；公开 OpenAPI 未列为必填。 |

## 逻辑搜索

`logic_search` 顶层存在时，`logic_type` 必填：

| JSON 路径 | 类型 | 必填性 | 说明 |
| --- | --- | --- | --- |
| `logic_search.logic_type` | integer | 必填 | `0` AND；`1` OR；`2` 叶子节点。 |
| `logic_search.children` | array<object> | AND/OR 节点使用 | 当前 CLM DTO 类型是递归对象数组。公开 OpenAPI 仍标成 string，执行时以当前服务实现为准。 |
| `logic_search.data` | object | 叶子节点必填 | 叶子搜索条件。 |
| `logic_search.data.field_name` | string | 叶子节点必填 | 搜索字段。 |
| `logic_search.data.operator` | integer | 叶子节点必填 | 搜索操作符。 |
| `logic_search.data.field_value` | 任意 JSON 值 | 叶子节点必填 | 字段值，类型取决于搜索字段。 |

公开字段白名单：

- `tradingPartyName`：交易方名称
- `contractNumber`：合同编号
- `contractName`：合同名称
- `categoryName`：合同分类名称
- `contractStatus`：合同状态

操作符：`0` 等于、`1` 不等于、`6` 包含、`7` 不包含。

## 枚举与约束

合同状态：

| 值 | 含义 | 值 | 含义 |
| --- | --- | --- | --- |
| `0` | editing | `7` | signed |
| `1` | cancelled | `8` | archiving |
| `2` | revoked | `9` | archived |
| `3` | process | `10` | changing |
| `4` | rejected | `11` | changed |
| `5` | approved | `12` | 我方已签约 |
| `6` | signing | `13` | 对方已签约 |

必须遵守：

- 顶层 `contract_number` 非空时，V1 选择精确编号查询，最多返回一条，且忽略 `combine_condition` 和 `logic_search`。
- 顶层编号未命中时返回业务错误 `110107`，不是成功空列表。
- `combine_condition` 存在时优先于 `logic_search`；不要同时传两者。
- `condition_units`、`filter_units`、`search_tab_code`、`sort_type`、`order` 属于 user MCP 搜索，不应传给 app V1。
- 时间字段使用 `YYYY-MM-DD HH:mm:ss`。

## 示例

精确查询合同编号：

```bash
contract-cli contract search --profile contract --as app --data '{"contract_number":"CT20210708000030"}'
```

组合条件请求文件 `search-app.json`：

```json
{
  "page_size": 20,
  "combine_condition": {
    "user_id": "ou_xxx",
    "permission": 1,
    "contract_status_in": "3,6,9",
    "contract_name": "采购合同",
    "submited_time_start": "2026-08-01 00:00:00",
    "submited_time_end": "2026-08-31 23:59:59"
  }
}
```

```bash
contract-cli contract search --profile contract --as app --input-file search-app.json
```

逻辑搜索请求体：

```json
{
  "page_size": 10,
  "logic_search": {
    "logic_type": 0,
    "children": [
      {
        "logic_type": 2,
        "data": {
          "field_name": "contractName",
          "operator": 6,
          "field_value": "采购"
        }
      },
      {
        "logic_type": 2,
        "data": {
          "field_name": "contractStatus",
          "operator": 0,
          "field_value": 9
        }
      }
    ]
  }
}
```

## 响应行为

- 响应 envelope 为 `code`、`msg`、`data`。
- 常用路径为 `data.items[]`、`data.has_more`、`data.page_token`。
- 顶层编号精确查询固定 `has_more=false`、`page_token=""`。
- 完整合同字段见 [contract-response-fields.md](../../contract-cli-contract/references/contract-response-fields.md)。
