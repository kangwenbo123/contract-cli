# Contract Search Routing

`contract-cli` 的合同搜索按身份和搜索语义路由到三套接口。先选接口，再读取对应参数文档；不要把三套请求体字段混用。

先按 [主 Skill 总纲](../SKILL.md) 确认当前 profile、配置目录与身份的授权状态，已有有效授权可复用。`auth status --profile contract --as user|app` 不支持 `--output`，不会刷新 Token，也不能提供可靠的当前业务 user_id；缺失或过期按授权 Skill 处理，不索要原始 Token。

用户按“合同文本包含关键词”查询时，先读 [文本搜索配方](text-search.md)，明确页面五类来源与仅正文的区别。

## 接口选择

| 用户意图 | 命令 | 接口 | 参数文档 |
| --- | --- | --- | --- |
| 以当前登录用户的可见权限搜索，使用 PC 兼容关键词或结构化筛选 | `contract search --as user` | `POST /open-apis/contract/v1/mcp/contracts/search` | [search-user-parameters.md](search-user-parameters.md) |
| app 按合同编号精确查询，或使用旧组合条件、逻辑条件 | `contract search --as app` | `POST /open-apis/contract/v1/contracts/search` | [search-app-parameters.md](search-app-parameters.md) |
| app 按合同编号做 ES 模糊查询或批量查询，并分页返回多条结果 | `contract search-v2 --as app` | `POST /open-apis/contract/v1/contracts/searchV2` | [search-v2-parameters.md](search-v2-parameters.md) |

## 不可混用的关键语义

- `contract search --as user` 才支持 `condition_units`、`filter_units`、`search_tab_code` 和 MCP 排序字段；权限由当前 OAuth 用户决定。
- `contract search --as app` 的顶层 `contract_number` 是精确查询，最多返回一条；未命中返回业务错误 `110107`。
- `contract search-v2 --as app` 的顶层 `contract_number` 是 ES 模糊查询；传分号分隔的多个编号时进入精确批量查询。
- app V2 不是 user MCP 搜索的替代品：当前后端明确忽略 `condition_units` 和 `filter_units`。
- app V1 与 app V2 都保留 `combine_condition` / `logic_search`，但 app V2 只有在未传顶层 `contract_number` 时才委托旧搜索逻辑。
- `contract search-fields`、`employee list`、`department list` 仅支持 user；`mdm vendor list/get` 与 `mdm legal list/get` 已支持两种身份，但 app 的查询参数语义仍应读各自文档，不能把 user 验证结论推广为 app 完全对等。
- 本 Skill 的十二字段“全部关键词”和精确对象 filter 配方仅适用于 user。只有 app 授权且无法表达需求时，明确说明缺少的能力，保留原查询意图，不静默切身份或删条件。

## CLI 公共行为

- `contract search` 支持 `--contract-number`、`--page-size`、`--page-token`；这些 flag 会覆盖并写入 JSON body 的同名字段。
- 复杂请求体使用 `--input-file`，简单请求可使用 `--data`；两者互斥。
- 示例必须显式携带 `--as user` 或 `--as app`，避免 profile 默认身份改变实际接口。
- 响应字段入口见 [contract-response-fields.md](../../contract-cli-contract/references/contract-response-fields.md)。
