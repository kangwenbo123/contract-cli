# 下载合同全部文件到指定目录

用户要求下载合同的全部文件时使用。由 Agent 编排现有 CLI 命令：查询合同详情 → 收集合同和表单文件 → 查询各流程的审批附件及评论树附件 → 按文件 ID 去重 → 逐个获取临时 URL 并下载。没有独立的批量下载命令；仅要求某个文件时按指定范围选择，不自动扩大为全部下载。

## 1. 读取合同和文件字段

沿用用户指定的 profile，全程使用同一 user 身份。名称或编号不能唯一定位合同时先展示候选；确定合同 ID 后执行：

```bash
contract-cli contract get <contract-id> --profile <profile> --as user --output json
```

确认命令成功及业务响应成功后，读取 `data.contract`；下表路径相对于该对象。ID 使用无损 JSON 解析并保留为字符串，不经过 JavaScript Number 等浮点数转换。

| 文件来源 | 文件项路径 |
| --- | --- |
| 正文 | `contract_files.contract_text`（对象，不是数组） |
| 签订依据 | `contract_files.contract_causes[]` |
| 普通附件 | `contract_files.contract_attachments[]` |
| 扫描件 | `contract_files.contract_scans[]` |
| 归档附件 | `contract_files.contract_archive_attachments[]` |
| 表单自定义附件 | `contract_files.contract_form_attachments[]` |
| 模板自定义附件 | `contract_files.contract_template_custom_attachments[]` |
| OCR 比对文件 | `contract_files.contract_ocr_comparisons[]` |

这些文件项使用 `file_id`、`file_name`；大小和 MIME 可作为参考。收集候选不代表已通过下载权限校验。不要使用详情里的旧 `download_url`，后续统一通过下载接口取得新地址。

`form` 是 JSON 字符串，先解析，再遍历字段及明细行。文件字段的 `attribute_value[]` 包含 `file_id`、`file_name` 等信息；结合 `attribute_type` 和实际字段结构确认文件值，保留字段名称、模块和行定位。多文件必须逐项收集，不能只取第一项；不要把任意 `id`、字段配置 ID、合同 ID 或 URL 当成文件 ID。嵌套对象/数组按实际结构遍历，不假设所有字段都在顶层；未知字段类型或无法解析的内容记录为清单不完整，不猜测字段值。

## 2. 收集流程与评论附件

收集 `process_instance_ids[]`，并合入非空的 `process_instance_id`，按流程 ID 去重。后者是单个字符串；只有当前流程时只查询一次。不要用 `task_instance_id` 替代流程 ID，也不要为寻找流程枚举 ID。

对每个流程执行以下两个读取命令，不加任务过滤，以免遗漏附件；一个读取失败仍继续其他可执行的读取：

```bash
contract-cli contract approval get <process-instance-id> --profile <profile> --as user --output json
contract-cli contract approval comment list <process-instance-id> --profile <profile> --as user --output json
```

- 流程附件：读取 `data.process_instance.task_instance_list[].attachments[]` 中的文件项，收集 `file_id`；检查 `data.process_instance.complete` 和 `limitations`。
- 评论附件：从 `data.items[]` 开始，收集每层 `attachments[]`，递归遍历 `replies[]`，包括深层回复。评论接口不分页。`available=false` 的附件记录为不可用；不要因为父评论已删除就停止遍历其回复。
- 若响应给出所属合同 ID，核对其与目标合同一致；不一致的来源停止处理并记录原因，不擅自切换合同 ID 下载。
- 历史流程能查询到附件，不保证文件仍属于当前可下载集合；仍以统一下载接口的结果为准。拒绝时记为失败，不绕过校验获取存储地址。

## 3. 去重并准备保存路径

为每项保存合同 ID、文件 ID、原始文件名、所有来源（分类/字段/流程/评论）及处理状态。在当前合同内按字符串 `file_id` 去重，合并来源，每个文件只下载一次；多个合同分别处理，不跨合同混用文件 ID。不要按文件名去重。一个来源标记不可用但其他来源仍可用时，可对该 ID 调用一次统一下载接口，最终以服务端判断为准。

使用用户指定目录；未给出目录且需要批量本地保存时，先询问保存目录。构造完整文件路径传给 `--output-file`，不能把目录直接当文件路径。可创建指定目录中缺失的子目录。

只取文件名的安全末段，替换路径分隔符、控制字符及目标系统不允许的字符，拒绝 `.`/`..`，确保结果仍在指定目录内。缺少名称时使用 `file-<file_id>`，不猜扩展名。同名不同 ID 在扩展名前追加文件 ID；仍与已有文件冲突时继续添加序号。默认不使用 `--force`，已有文件不能仅凭名称视为本次已下载。

## 4. 逐个下载与汇总

```bash
contract-cli contract download-file <file-id> --contract <contract-id> --profile <profile> --as user --output-file <完整保存路径>
```

CLI 自动获取 300 秒有效的临时 URL，再请求文件流，先写同目录临时文件并校验，完成后才落到目标路径。不要自行把下载元数据当文件保存，也不要用 `--raw > 文件` 实现批量保存，否则无法享受 CLI 的临时文件保护。成功提示是文本，不要按 JSON 解析下载命令 stdout。

每项检查退出码；成功后确认目标文件存在并读取实际字节数。记录成功、失败或跳过及原因，单文件失败继续剩余项。认证失效时暂停后续依赖授权的调用，保留已完成结果并提示重新授权。权限/参数错误不盲目重试；临时网络失败可重新调用下载命令至多一次以取得新 URL，不能重新下载已成功项。不输出或写入报告任何临时 URL、Token 或 Cookie。

最终返回每项的文件 ID、名称、来源、绝对保存路径（成功项）、实际大小和失败/跳过原因，并汇总去重后的候选数及各状态数量。全部查询成功、结构已解析且每项均下载成功时，才能说“本次查询范围内的文件全部下载完成”。任一流程 `complete=false`、来源查询失败、字段缺失或结构未知时，说明已下载哪些文件以及清单完整性未知；不要将缺失字段等同于空集合，也不要把“没有发现”写成“没有文件”。

本流程使用现有临时 URL 下载契约，不承诺下载时重新解析最新版本或额外水印处理。文件是否可下载最终由服务端决定。
