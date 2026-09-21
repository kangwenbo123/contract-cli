# Entity Query Parameters

这份附录专门解决 `mdm legal list` / `mdm legal get` 的参数映射问题，其中 `mdm legal get --code` 用于按法人实体编码查询。

## 1. 请求面导航

```text
mdm legal list
├── --name       -> $query.legalEntity
├── --page-size  -> $query.page_size
├── --page-token -> $query.page_token
├── --profile    -> profile selector
├── --as         -> identity selector
├── --user-id    -> common query passthrough
├── --user-id-type -> common query passthrough
├── --output     -> output format
└── --raw        -> raw response switch

mdm legal get
├── <legal-entity-id> -> $path.legal_entity_id
├── --code            -> $query.legalEntity
├── --page-size       -> $query.page_size
├── --page-token      -> $query.page_token
├── --profile         -> profile selector
├── --as              -> identity selector
├── --user-id         -> common query passthrough
├── --user-id-type    -> common query passthrough
├── --output          -> output format
└── --raw             -> raw response switch
```

## 2. `mdm legal list`

请求形态：

- 方法：`GET`
- `--as user` 路径：`/open-apis/contract/v1/mcp/legal_entities`
- `--as app` 路径：`/open-apis/mdm/v1/legal_entities/list_all`

| CLI 参数 | 请求位置 | 类型 | 必填性 | 业务含义 | 联动/注意 |
| --- | --- | --- | --- | --- | --- |
| `--name` | `$query.legalEntity` | `string` | 可选 | 法人实体查询关键字。 | 传值时按名称模糊搜索；CLI flag 叫 `--name`，底层 query key 实际是 `legalEntity`。 |
| `--page-size` | `$query.page_size` | `integer` | 可选 | 每页条数。 | 当前接口默认 `10`，建议不超过 `50`。 |
| `--page-token` | `$query.page_token` | `string` | 可选 | 分页令牌。 | 首次请求通常不传，翻页时使用上一页返回的 `page_token`。 |
| `--profile` | 本地上下文 | `string` | 可选 | 选择 profile。 | 未传时走默认 profile。 |
| `--as` | 本地上下文 | `string` | 可选 | 选择身份。 | `user` 走 MCP 路径，`app` 走开放平台 `mdm/v1/legal_entities/list_all` 路径。 |
| `--user-id` | `$query.user_id` | `string` | 可选 | 通用用户标识参数。 | 文档里未显式列出，但 CLI 仍按共享约定透传。 |
| `--user-id-type` | `$query.user_id_type` | `string` | 可选 | 通用用户标识类型。 | app 文档里显式列出；CLI 不做必填校验。 |
| `--output` | CLI 输出 | `string` | 可选 | 输出格式。 | 常用 `json` / `yaml` / `table`。 |
| `--raw` | CLI 输出 | `boolean` | 可选 | 返回原始 envelope。 | 排障时常用。 |

## 3. `mdm legal get`

请求形态：

- 按 ID 查询方法：`GET`
- 按 ID 查询 `--as user` 路径：`/open-apis/contract/v1/mcp/legal_entities/{legal_entity_id}`
- 按 ID 查询 `--as app` 路径：`/open-apis/mdm/v1/legal_entities/{legal_entity_id}`
- 按编码查询方法：`GET`
- 按编码查询路径：`/open-apis/mdm/v1/legal_entities`
- 按编码查询身份：仅支持 `--as app`

| CLI 参数 | 请求位置 | 类型 | 必填性 | 业务含义 | 联动/注意                                                                       |
| --- | --- | --- | --- | --- |-----------------------------------------------------------------------------|
| `<legal-entity-id>` | `$path.legal_entity_id` | `string` | 条件必填 | 法人实体 id。 | 不传 `--code` 时必填；`user` 走 MCP 路径；`app` 也会把同一值额外透传到 `$query.legal_entity_id`。 |
| `--code` | `$query.legalEntity` | `string` | 条件必填 | 法人实体编码。 | 按编码查询时必填；不能同时传 `<legal-entity-id>`；该模式只支持 app。                              |
| `--page-size` | `$query.page_size` | `integer` | 可选 | 每页条数。 | 仅 `--code` 模式可用。                                                            |
| `--page-token` | `$query.page_token` | `string` | 可选 | 分页令牌。 | 仅 `--code` 模式可用。                                                            |
| `--profile` | 本地上下文 | `string` | 可选 | 选择 profile。 | 未传时走默认 profile。                                                             |
| `--as` | 本地上下文 | `string` | 可选 | 选择身份。 | `user` 走 MCP 路径，`app` 走开放平台 `mdm/v1/legal_entities/{legal_entity_id}` 路径。 |
| `--user-id` | `$query.user_id` | `string` | 可选 | 通用用户标识参数。 | 文档里未显式列出，但 CLI 仍按共享约定透传。                                                    |
| `--user-id-type` | `$query.user_id_type` | `string` | 可选 | 通用用户标识类型。 | app 文档里显式列出；CLI 不做必填校验。 |
| `--output` | CLI 输出 | `string` | 可选 | 输出格式。 | 常用 `json` / `yaml` / `table`。                                               |
| `--raw` | CLI 输出 | `boolean` | 可选 | 返回原始 envelope。 | 排障时常用。                                                                      |

## 4. 使用建议

- 2026-09-20 当前生产 user 授权已实测 `list --name` 返回一个已知主体且 `has_more=false`，同一 ID 的 `get` 也业务成功；名称和 ID 与已知合同我方主体对应。这证明该身份可查列表/详情，不等于合同搜索已支持通用我方 ID 精确筛选。
- 本次 user 列表结构为 `data.items[].id/legal_entity_text/legal_entity`，分页 `data.has_more/page_token`；详情包在 `data.legalEntity`。按实际响应读取，保留主体 ID 原字符串，不用 `legal_entity` 编码代替 ID，也不把交易方列表的 camelCase 结构照搬到此处。
- 先 `list` 拿候选，再 `get` 看详情，是最稳的两步走方式
- 已知法人实体编码但没有 id 时，用 `mdm legal get --code`
- 如果只是为合同选择我方主体，优先记住法人实体 id
- 这组命令只做查询，不做本地字段裁剪或结构转换
- `mdm legal list/get`、`mdm vendor list/get` 和 `mdm fields list` 现在都支持按身份自动路由；`mdm legal get --code` 是 app-only；其中 `mdm fields list --as app` 仅支持 `vendor` / `legal_entity`，`vendor_risk` 走 user/MCP 路径
