# User MCP 审批接口说明

本文说明合同开放平台本期新增的 6 个 User MCP 接口，以及它们在 `contract-cli` 中的命令映射、调用顺序和错误处理方式。原始接口定义来自[飞书接口目录](https://ysi13ckdb9.feishu.cn/wiki/X9b9wFf62itL2zkMYwrcDt9jnid)，本文以 2026-09-07 读取到的内容和当前 CLI 实现为准。

## 1. 接口总览

| 能力 | Method 与 Path | CLI 命令 | 操作语义 |
|---|---|---|---|
| 查询审批评论 | `GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments` | `contract approval comment list` | 读取 |
| 下载合同文件 | `GET /open-apis/contract/v1/mcp/contracts/{contract_id}/files/{file_id}/download` | `contract download-file --as user` | 读取 |
| 查询审批流程实例详情 | `GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}` | `contract approval get --as user` | 读取 |
| 创建审批评论 | `POST /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments` | `contract approval comment create` | 非幂等写入 |
| 查询个人任务列表 | `POST /open-apis/contract/v1/mcp/tasks` | `contract approval task list` | 读取 |
| 审批任务通过/拒绝 | `POST /open-apis/contract/v1/mcp/tasks/{task_instance_id}/approval` | `contract approval task approve/reject` | 非幂等写入 |

这 6 个接口都使用个人授权身份：

```http
Authorization: Bearer <user_access_token>
```

调用人、任务归属人、评论发起人都由 Token 唯一确定，不能通过 `user_id`、query 或请求体切换为其他用户。接口频率限制以开放平台统一频控配置为准。

## 2. 推荐调用链

```mermaid
flowchart LR
    A[查询个人任务列表] --> B[取得 task_instance_id 和 process_instance_id]
    B --> C[查询审批流程实例详情]
    C --> D[查询审批评论]
    C --> E[下载流程附件]
    D --> F[下载评论附件]
    B --> G[通过或拒绝普通审批任务]
    D --> H[创建评论或回复]
    G --> I[重新查询任务/流程状态]
    H --> D
```

推荐先查询本人任务，再读取流程和评论。只有用户明确表达通过或拒绝意图后，才调用任务处理接口；写操作返回结果不确定时，先查询最新状态，不要直接重试。

## 3. 通用响应和数据约定

默认响应为传统结构：

```json
{
  "code": 0,
  "msg": "success",
  "data": {},
  "success": true
}
```

- `code=0` 表示业务成功；不能只根据 HTTP 200 判断成功。
- 流程、任务、评论、合同和文件 ID 应按字符串传输和保存。
- 部分响应中的 `tenant_id`、`contract_id`、`applicant_user_id` 等字段仍以 JSON Long 返回，可能超过 JavaScript 安全整数范围，前端或 Node.js 调用方应使用无损数字解析。
- `110125` 通常同时表示资源不存在、当前用户不可见或当前状态不可操作。调用方不应利用该错误枚举资源是否存在。
- CLI 不在输出或日志中记录 Access Token、预签名下载地址等敏感信息。

## 4. 查询审批流程实例详情

```http
GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}
```

权限要求：当前个人用户对审批实例关联合同具有读取权限。

### 请求参数

| 位置 | 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|---|
| Path | `process_instance_id` | String | 是 | 合同审批流程实例 ID |
| Query | `task_instance_filter` | String | 否 | 任务实例 ID；多个值用英文逗号分隔 |
| Query | `notice_filter` | String | 否 | 知会记录 ID；多个值用英文逗号分隔 |
| Header | `X-MCP-Response-Profile` | String | 否 | 不传时为 `clm-web-status-v1.0` |

支持的响应 Profile：

| Profile | 行为 |
|---|---|
| 不传或 `clm-web-status-v1.0` | `code/msg/data` 结构，业务错误通常仍为 HTTP 200 |
| `clm-web-status-v2.0` / `v2.1` | 传统结构，业务错误映射为非 2xx HTTP 状态 |
| `clm-mcp-outcome-v1.0` | `io.unbox/provider-outcome@1` 结构，HTTP 状态固定为 200 |

当前 CLI 不主动发送 `X-MCP-Response-Profile`，保持服务端默认兼容模式。

### 主要响应字段

`data.process_instance` 包含：

| 字段 | 类型 | 说明 |
|---|---|---|
| `process_instance_id` | String | 流程实例 ID |
| `process_definition_id` / `process_definition_key` | String | 流程定义标识 |
| `instance_status` | String | 流程状态 |
| `biz_key` | String | 关联业务单据标识，合同场景通常为合同 ID |
| `initiator_id` | Long | 流程发起人 Lark ID |
| `start_time` / `end_time` | String | 毫秒时间戳字符串；未结束时 `end_time` 可能为空 |
| `process_subject` / `process_name` | Object | `zh/en/ja` 多语言文案 |
| `task_instance_list` | Object[] | 流程实际产生的任务实例 |
| `notice_list` | Object[] | 知会记录 |

任务实例的重点字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `task_instance_id` | String | 后续审批动作使用的任务 ID |
| `node_id` / `node_name` | String / Object | 节点标识和多语言名称 |
| `assignee_ids` | Long[] | 审批人 Lark ID |
| `assignee_user_ids` | String[] | 审批人开放平台 ID，可能不返回 |
| `command_type` / `command_type_name` | String / Object | 审批动作及其多语言名称 |
| `task_comment` | String | 审批意见，可能为空 |
| `attachments` | Object[] | 审批附件，可能省略 |
| `task_status` | Integer | 任务状态，无法补齐时不返回 |

`task_status` 枚举：`0` 运行中、`1` 已完成、`2` 已拒绝、`3` 已撤回、`4` 已撤销、`5` 自动终止。字段缺失时表示状态未知，不要根据 `end_time`、`command_type` 或审批意见推断。

### CLI 示例

```bash
contract-cli contract approval get <process-instance-id> \
  --profile contract --as user \
  --task-instance-filter <task-id-1>,<task-id-2>
```

## 5. 查询审批评论

```http
GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments
```

权限要求：当前用户具有审批实例关联合同的查看权限。

接口没有 query 和分页参数，一次返回完整递归评论树。一级评论位于 `data.items`，回复位于每个评论的 `replies`，回复结构与一级评论相同。

### 主要响应字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.process_instance_id` | String | 审批实例 ID |
| `data.contract_id` | String | 关联合同 ID，可用于下载评论附件 |
| `items[].comment_id` | String | 评论 ID |
| `items[].parent_comment_id` | String | 父评论 ID；一级评论为 `null` |
| `items[].content` | String | 包含已渲染 @ 用户文本的评论正文 |
| `items[].author` | Object | `user_id`、`name` |
| `items[].mentions` | Object[] | 被 @ 用户的 `user_id`、`name` 和正文 `offset` |
| `items[].create_time` / `update_time` | Long | 毫秒时间戳 |
| `items[].is_deleted` | Boolean | 评论是否已删除 |
| `items[].attachments` | Object[] | 当前节点直接关联的附件 |
| `items[].replies` | Object[] | 直接回复，可继续递归 |

附件包含 `file_id`、`file_name`、`file_size`、`mime_type`、`available`。只有 `available=true` 时才应调用下载接口；已删除评论可能仍保留在树中，以维持回复关系。

### CLI 示例

```bash
contract-cli contract approval comment list <process-instance-id> \
  --profile contract --as user
```

## 6. 创建审批评论

```http
POST /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments
Content-Type: application/json
```

评论发起人固定为个人 Token 对应用户。`user_id_type` 只解释被 @ 用户的 ID 类型，不指定评论发起人。

### 请求参数

| 位置 | 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|---|
| Path | `process_instance_id` | String | 是 | 合同审批实例 ID |
| Query | `user_id_type` | String | 否 | `user_id` 数组的 ID 类型；默认 `open_id` |
| Body | `content` | String | 否 | 普通正文；存在 @ 用户时不需要手工拼接 `@姓名` |
| Body | `parent_comment_id` | String | 否 | 被回复的评论 ID；不传时创建一级评论 |
| Body | `user_id` | String[] | 否 | 被 @ 用户列表，类型由 `user_id_type` 决定 |
| Body | `file_ids` | String[] | 否 | 已上传到合同系统的附件 ID |
| Body | `task_instance_id` | String | 否 | 仅用于评论通知跳转兜底 |

字段名固定为 `user_id`，且值必须为数组；不能使用 `user_ids`、`at_user_ids` 或 `at_info_list` 代替。接口会按数组顺序生成 @ 用户文本及位置。

请求示例：

```json
{
  "content": "请查看补充材料",
  "parent_comment_id": "112233445566778899",
  "user_id": ["ou_10001", "ou_10002"],
  "file_ids": ["223344556677889900"],
  "task_instance_id": "task-instance-001"
}
```

成功响应的 `data` 包含 `comment_id`、`process_instance_id` 和 `business_id`。

### CLI 示例

```bash
contract-cli contract approval comment create <process-instance-id> \
  --profile contract --as user \
  --mention-id-type open_id \
  --data '{"content":"请确认审批意见","user_id":["ou_10001"]}'
```

该接口非幂等，重复调用会创建多条评论。遇到超时、网络中断或 CLI 返回“执行结果不确定，请先查询确认”时，应先调用评论查询接口确认结果。

## 7. 查询个人任务列表

```http
POST /open-apis/contract/v1/mcp/tasks
Content-Type: application/json
```

虽然使用 POST，该接口仍是读取操作，只查询 Token 对应用户的待办、已办或抄送/知会任务。直接调用 API 时至少发送空 JSON 对象 `{}`。

### 请求体

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `query` | String | - | 合同名称、合同编号或申请人名称 |
| `task_type_code` | Integer | `0` | `0` 待办、`1` 已办、`2` 抄送/知会 |
| `business_type_codes` | Integer[] | - | 单据业务类型编码 |
| `desc` | Boolean | `true` | 是否按任务时间倒序 |
| `read` | Boolean | - | 知会是否已读，仅类型 `2` 生效 |
| `only_urge` | Boolean | - | 是否仅查询被催办任务，仅类型 `0` 生效 |
| `task_instance_id` | String | - | 待办或已办任务游标 |
| `notice_id` | String | - | 抄送/知会游标 |
| `page_index` | Integer | `0` | 从 0 开始的页码 |
| `page_size` | Integer | `10` | 每页 `1-100` 条 |
| `limit_before` / `limit_after` | Integer | - | 游标前后的任务数量 |
| `node_types` | Integer[] | - | 审批节点类型 |
| `check_cursor_status` | Boolean | `true` | 是否校验游标状态和任务类型匹配 |
| `legal_entity_source_ids` | String[] | - | 我方主体来源 ID |
| `trading_party_source_ids` | String[] | - | 对方主体来源 ID |

CLI 为常用字段提供 `--query`、`--task-type todo|done|notice`、`--page-index`、`--page-size`；其他筛选条件通过 `--input-file` 或 `--data` 传入，命令行 flag 会覆盖 JSON 中的同名字段。

### 主要响应字段

`data` 包含 `task_list`、`total`、`page_index`、`page_size` 和 `has_more`。任务项重点字段如下：

| 字段 | 类型 | 说明 |
|---|---|---|
| `contract_id` | Long | 合同 ID，注意 JavaScript 精度 |
| `contract_name` / `contract_number` | String | 合同标题和单号 |
| `applicant_user_id` | Long | 申请人 Lark ID |
| `task_instance_id` | String | 审批任务 ID；知会任务为空 |
| `notice_id` | String | 知会 ID；待办、已办任务为空 |
| `task_type` | Integer | `0` 待办、`1` 已办、`2` 抄送/知会 |
| `process_instance_id` | String | 流程实例 ID |
| `node_name` | String | 任务节点名称 |
| `timestamp` | Long | 任务时间戳，毫秒 |

### CLI 示例

```bash
contract-cli contract approval task list \
  --profile contract --as user \
  --task-type todo --query "采购合同" --page-size 20
```

盖章和归档任务可能出现在列表中，但不能使用下一个接口处理。

## 8. 审批任务通过或拒绝

```http
POST /open-apis/contract/v1/mcp/tasks/{task_instance_id}/approval
Content-Type: application/json
```

权限要求：当前用户是运行中任务的办理人，并具备首页、合同管理或任务中心中的任一功能权限。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `action` | String | 是 | `approve` 或 `reject`，大小写不敏感 |
| `comment` | String | 条件必填 | `reject` 时必填且不能为空白；`approve` 时选填 |
| `file_ids` | String[] | 否 | `approveAttachment` 类型的审批附件 ID |
| `archive_number` | String | 否 | 兼容保留字段；MCP 当前不处理归档节点，无需传入 |

`file_ids` 中每个 ID 必须是正整数，属于当前租户和当前审批业务，且上传类型为 `approveAttachment`；重复 ID 按首次出现顺序去重。

通过示例：

```json
{
  "action": "approve",
  "comment": "同意",
  "file_ids": ["6911661136408477999"]
}
```

拒绝示例：

```json
{
  "action": "reject",
  "comment": "合同条款风险尚未解决，请修改后重新提交"
}
```

成功响应包含 `task_instance_id`、`process_instance_id`、实际 `action` 和动作后的 `process_status`。

### CLI 示例

```bash
contract-cli contract approval task approve <task-instance-id> \
  --profile contract --as user --comment "同意" \
  --file-id 6911661136408477999

contract-cli contract approval task reject <task-instance-id> \
  --profile contract --as user \
  --comment "合同条款风险尚未解决"
```

本接口只处理普通审批节点，不支持盖章和归档节点，也不支持转办、加签、撤回、作废等动作。它不是幂等接口；成功后重复调用会返回不可操作。出现“执行结果不确定”时，先重新查询任务列表或流程详情。

## 9. 下载合同文件

```http
GET /open-apis/contract/v1/mcp/contracts/{contract_id}/files/{file_id}/download
```

权限要求：当前用户具有合同读取权限，并具备合同管理或任务中心中的任一功能权限。`contract_id`、`file_id` 都必须是正整数且在 Long 范围内，文件必须属于该合同的可见文件集合。

可下载范围包括合同正文、用印前文件、普通附件、扫描件、表单自定义附件、内部模板附件、评论/审批附件、最新用印 OCR 对比文件及归档附件。不支持其他合同文件、外部模板附件或不可见文件。

### 响应字段

接口返回 JSON 元数据，不直接返回二进制：

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.file_id` | String | 文件 ID |
| `data.file_name` | String | 文件名称 |
| `data.file_size` | Long | 文件大小，字节 |
| `data.mime_type` | String | MIME 类型 |
| `data.file_type` | String | 文件业务类型 |
| `data.expires_in_seconds` | Integer | 下载地址有效期，固定 `300` 秒 |
| `data.download_url` | String | 对象存储预签名地址 |

获取成功响应后，应在有效期内直接访问 `download_url`。第二次请求不携带开放平台 `Authorization`；临时地址不得打印、持久化或转发给无权限人员。

当前 CLI 会自动完成两段下载流程：校验 HTTPS 地址、发起无授权请求、检查响应和 `file_size`，普通文件模式先写同目录临时文件，完整成功后再提交目标文件。

```bash
contract-cli contract download-file <file-id> \
  --contract <contract-id> \
  --profile contract --as user \
  --output-file ./附件.pdf
```

`--raw` 会把二进制直接写到 stdout：

```bash
contract-cli contract download-file <file-id> \
  --contract <contract-id> \
  --profile contract --as user --raw > 附件.pdf
```

## 10. 错误码汇总

| 业务码 | 典型场景 | 建议 |
|---|---|---|
| `110000` | 任务查询或处理参数非法 | 检查任务类型、分页、action、comment 和附件 ID |
| `110001` | 下载接口没有有效个人登录身份 | 重新完成 user 授权 |
| `110002` | 系统、BPM 或文件存储链路异常 | 读接口可稍后重试；写接口先查询状态 |
| `110004` | 无法解析个人调用人 | 检查是否使用有效 `user_access_token` |
| `110125` | 流程、合同、文件、任务不存在、不可见或不可操作 | 重新查询本人可见资源，不区分不存在和无权限 |
| `110507` | 盖章或归档任务不支持 MCP 审批 | 转到合同系统页面处理 |
| `110509` | 合同、流程或下游状态不允许该动作 | 根据返回信息检查最新状态 |
| `110602` | 评论写入失败 | 查询评论确认是否成功，再决定后续处理 |
| `110603` | 评论内容超过系统长度限制 | 缩短评论正文 |
| `110903` | 无有效个人身份，或被 @ 用户无效/不属于租户 | 检查授权和 `user_id_type/user_id` |
| `111706` | 下载路径 ID 非正整数或超出 Long 范围 | 修正 `contract_id` / `file_id` |

评论创建还可能返回公共参数错误码，涵盖请求体格式、父评论、附件 ID 或 @ 用户转换错误。

## 11. CLI 重试和幂等策略

| 操作 | CLI 策略 |
|---|---|
| 流程详情、评论列表、下载元数据 | 按读取操作处理，临时网络错误最多自动重试一次 |
| 个人任务列表 | 虽是 POST，仍声明为读取操作，临时网络错误最多自动重试一次 |
| 创建评论 | 非幂等写入，不自动重试；网络错误或 5xx 返回“执行结果不确定” |
| 任务通过/拒绝 | 非幂等写入，不自动重试；网络错误或 5xx 返回“执行结果不确定” |
| 预签名地址的二次文件请求 | 不携带 Token，不输出 URL；失败后重新执行命令获取新地址 |

## 12. 联调检查清单

- 使用 user profile 和有效个人授权，不使用纯 app Token 调用 MCP 路径。
- 验证当前用户对合同、流程或任务确实可见，不向调用方区分“不存在”和“无权限”。
- 所有长 ID 使用字符串或无损整数处理，避免 JavaScript 精度丢失。
- 评论 @ 用户时确认 `user_id_type` 与 `user_id` 数组一致。
- 审批附件先按 `approveAttachment` 上传，并确认属于当前审批业务。
- 只对普通审批节点调用 approve/reject；盖章和归档节点转页面处理。
- 写操作超时后先查询，不盲目重复提交。
- 下载地址仅短期使用，不进入日志、数据库、缓存或对外响应。
