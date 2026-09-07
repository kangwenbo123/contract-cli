# User MCP 审批、评论、任务与文件下载

这组命令全部以个人 Token 确定当前用户。除 `contract approval get` 和 `contract download-file` 兼容 app 身份外，评论与个人任务命令仅支持 `--as user`。

## 命令与接口

| CLI 命令 | 接口 | 语义 |
|---|---|---|
| `contract approval get <process-instance-id> --as user` | `GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}` | 读取 |
| `contract approval comment list <process-instance-id> --as user` | `GET /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments` | 读取 |
| `contract approval comment create <process-instance-id> --as user` | `POST /open-apis/contract/v1/mcp/process_instances/{process_instance_id}/comments` | 非幂等写入 |
| `contract approval task list --as user` | `POST /open-apis/contract/v1/mcp/tasks` | 读取 |
| `contract approval task approve <task-instance-id> --as user` | `POST /open-apis/contract/v1/mcp/tasks/{task_instance_id}/approval` | 非幂等写入 |
| `contract approval task reject <task-instance-id> --as user` | 同上 | 非幂等写入 |
| `contract download-file <file-id> --contract <contract-id> --as user` | `GET /open-apis/contract/v1/mcp/contracts/{contract_id}/files/{file_id}/download` | 读取并下载 |

## 流程详情与评论

- `contract approval get` 可带 `--notice-filter` 和 `--task-instance-filter`。`task_status` 可能缺失；存在时应优先使用它，不要从 `end_time` 或审批意见推断状态。
- CLI 不主动发送 `X-MCP-Response-Profile`，沿用后端默认 `clm-web-status-v1.0` 兼容配置。
- `comment list` 返回完整递归评论树，不分页。已删除评论可能仍保留；附件元数据中的 `available=false` 表示不可下载。
- `comment create` 请求体支持 `content`、`parent_comment_id`、`user_id`、`file_ids`、`task_instance_id`。
- `user_id` 和 `file_ids` 都是字符串数组。`--mention-id-type` 映射到 query `user_id_type`，仅解释被 @ 用户 ID；不指定评论发起人。
- 创建评论是非幂等操作。若 CLI 返回“执行结果不确定”，先调用 `comment list` 确认，不能盲目重试。

评论示例：

```bash
contract-cli contract approval comment list <process-instance-id> --profile contract --as user
contract-cli contract approval comment create <process-instance-id> --profile contract --as user --data '{"content":"请补充付款条件说明"}'
contract-cli contract approval comment create <process-instance-id> --profile contract --as user --mention-id-type open_id --input-file comment.json
```

## 个人任务查询

`contract approval task list` 支持把完整 JSON 对象放进 `--input-file` / `--data`，也提供以下常用 flag；flag 会覆盖请求体同名字段：

| CLI flag | JSON 字段 | 规则 |
|---|---|---|
| `--query` | `query` | 合同名称、编号或申请人名称 |
| `--task-type todo` | `task_type_code: 0` | 待办，默认类型 |
| `--task-type done` | `task_type_code: 1` | 已办 |
| `--task-type notice` | `task_type_code: 2` | 抄送/知会 |
| `--page-index` | `page_index` | 从 0 开始 |
| `--page-size` | `page_size` | 1-100 |

任务列表接口虽然是 POST，但属于读取操作，CLI 只会对临时网络错误安全重试一次。响应里的 `contract_id`、`task_instance_id` 等长 ID 按字符串保存。

```bash
contract-cli contract approval task list --profile contract --as user --task-type todo --page-size 20
contract-cli contract approval task list --profile contract --as user --data '{"business_type_codes":[0],"only_urge":true}'
```

## 任务通过与拒绝

- approve 的 `--comment` 可选；reject 的 `--comment` 必填且不能为空白。
- `--file-id` 可重复，按 `file_ids: String[]` 传入；必须是正整数、属于当前审批业务，且上传类型为 `approveAttachment`。
- CLI 不暴露 `archive_number`。盖章和归档节点不支持该接口，返回 `110507`，应转到合同系统页面处理。
- 任务处理非幂等，成功后重复调用会变为不可操作。若返回“执行结果不确定”，先用 task list 或 approval get 查询状态。

```bash
contract-cli contract approval task approve <task-instance-id> --profile contract --as user --comment "同意" --file-id <file-id>
contract-cli contract approval task reject <task-instance-id> --profile contract --as user --comment "合同条款风险尚未解决"
```

## 统一文件下载

user 身份下载必须同时提供合同 ID 和文件 ID。MCP 接口返回元数据和 300 秒有效的 `download_url`，CLI 会立即进行第二次无 Authorization 请求并写入文件，不会打印、持久化或记录该临时地址。

```bash
contract-cli contract download-file <file-id> --contract <contract-id> --profile contract --as user --output-file ./附件.pdf
contract-cli contract download-file <file-id> --contract <contract-id> --profile contract --as user --raw > 附件.pdf
```

普通文件模式先写同目录临时文件，校验 `file_size` 后再提交；`--force` 只在完整下载成功后替换旧文件。不支持硬链接的文件系统会退化为独占创建，正常错误返回时会清理未完成的新文件。`--raw` 直接输出二进制，不能配 JSON/YAML 消费方式。
