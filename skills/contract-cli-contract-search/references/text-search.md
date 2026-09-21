# 合同文本搜索

本页用于 `contract-cli contract search --profile contract --as user` 的文本内容查询。接口为 `POST /open-apis/contract/v1/mcp/contracts/search`；按当前 OAuth 用户的可见权限执行。

## 选择文本范围

先根据用户文字或截图识别搜索范围。若选的是“全部”，使用 [全部关键词搜索](all-keyword-search.md) 的十二字段配方；本文“合同文本”的五类文件文本只覆盖其中一部分，不包括名称、归属人、动态表单等。范围已明确时不追加确认。

页面的“合同文本”搜索会对同一关键词同时查询以下五类来源，并按 OR 组合，任一来源命中即可：

| 字段 | 页面/文件来源 |
| --- | --- |
| `CONTRACT_TEXT_FIELD` | 合同正文文本 |
| `CONTRACT_SCAN_TEXT` | 归档扫描件文本 |
| `CONTRACT_ATTACHMENT_TEXT` | 其他附件文本 |
| `CONTRACT_ARCHIVE_ATTACHMENT` | 归档附件文本 |
| `CONTRACT_CAUSE_FIELD` | 合同附件文本，不是“原因”字段 |

- “查合同文本包含某词”“和页面合同文本一样查询”：使用五字段配方，并说明查询包含正文及上述附件来源。
- “只查正文”“不查附件”：只使用 `CONTRACT_TEXT_FIELD`。
- 明确限定扫描件、合同附件或其他附件时，只保留对应字段。
- 搜索文件内容与搜索合同名称不同；不要自动加 `CONTRACT_NAME_FIELD`，也不要扩大到表单、自定义字段。

## 页面范围配方

将下列 JSON 保存为 `search-text.json`，把五处“里斯”一起替换成用户的实际关键词。

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "condition_units": [
    {"search_field": "CONTRACT_SCAN_TEXT", "search_value": "里斯", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "里斯", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_ATTACHMENT_TEXT", "search_value": "里斯", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_ARCHIVE_ATTACHMENT", "search_value": "里斯", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_CAUSE_FIELD", "search_value": "里斯", "union_type": "SHOULD"}
  ]
}
```

```bash
contract-cli contract search --profile contract --as user --input-file search-text.json
```

## 仅正文配方

仅在用户限定正文时使用，保存为 `search-body.json`：

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "condition_units": [
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "里斯", "union_type": "SHOULD"}
  ]
}
```

```bash
contract-cli contract search --profile contract --as user --input-file search-body.json
```

## 组合、结果和限制

- 这两个配方的关键词单元使用 `SHOULD`，省略时接口默认也是 `SHOULD`。它表示指定文本来源中至少一个命中，不表示可以完全忽略关键词。
- 2026-09-20 的生产只读对照中，相同正文关键词的 `MUST` 请求返回空结果，`SHOULD` 和省略 union_type 都能命中。不能把这一现象归因于“没有上传合同”，也不能据此宣称所有字段都不支持 MUST。
- 金额、币种等额外结构化条件应放在 `filter_units`；状态等按 [user 参数](search-user-parameters.md) 放在 `combine_condition`。不要将它们加入五字段 OR 组。
- 用户要求两个关键词都命中时，不能把十个文本单元简单并列为 `SHOULD`，那会变成任一词命中；当前配方只验证单个关键词跨来源搜索。复杂 AND/排除条件需先确认服务端支持的表达方式，不能自动放宽。
- 首次不传 `page_token`。成功响应仅在 `has_more=true` 时携带返回 token 翻页，并保持原文件的查询条件；不要只凭非空 token 继续。
- 检查 `code`/`success` 后再解释 items。业务错误、无命中和未查完分页是三种不同结果。
- 这些字段查询的是服务端文本索引，不是本地逐字遍历所有上传文件；不能承诺每种文件、每个历史版本都已入索引。若仍与页面不一致，核对双方用户/租户、页签、关键词、文本来源及分页，不擅自换身份或扩大范围。
- 返回列表可能没有命中片段；需要核对原文时按 [合同 Skill](../../contract-cli-contract/SKILL.md) 读取对应合同文本或文件。`contract text <id>` 是读取命令，不替代本页的搜索。
