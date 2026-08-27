# contract-cli 命令文档

本文档汇总当前代码里已经实际支持的 `contract-cli` 命令，作为后续继续扩展 app 接口和新业务命令的基线。

## 当前状态

- 当前仅内置 `prod` 环境预设；正式包默认使用 `prod`：`contract-cli config add --env prod --name contract`
- `contract get`、`contract search`、`contract create`、`contract sync-user-groups`、`contract text`、`contract category list`、`contract template list`、`contract template get`、`contract template instantiate`、`contract upload-file`、`mdm vendor list`、`mdm vendor get`、`mdm legal list`、`mdm legal get`、`mdm fields list` 是当前仅有的十五个同时支持 `user` 与 `app` 的结构化业务命令
- `contract search-v2`、`contract field update`、`contract sign switch-to-paper`、`contract sign-url get`、`contract form attribute list`、`contract authorization grant`、`contract esign *`、`contract submit/resubmit/patch/download-file/delete/print-file`、`contract share get/batch-create`、`contract cooperation link/record/search/file`、`contract approval start/get`、`payment *`、`mdm vendor create/update/list-all/query-by-cert`、`mdm legal get --code/create/update`、`mdm fixed-exchange-rate get/update`、`mdm file download`、`event outbound-ip list` 和 `rule table *` 当前仅支持 `--as app`
- 除上述双身份和 app-only 能力外，当前其他结构化业务命令仍只支持 `--as user`
- `app` 目前已经支持登录、状态查看、登出、默认身份切换
- 推荐使用 `npx skills add qfeius/contract-cli -y -g` 安装跨 Agent 平台 skills；`contract-cli skills install` 保留为 CLI 内置兜底
- `update check` 支持手动检查 npm 远端版本；默认输出文本，带 `--json` 时返回飞书式 JSON；CLI 会为符合条件的普通命令按 24 小时缓存检查远端版本，并在 JSON object 输出中注入 `_notice.update`
- `environment inspect` 可以在本地查看当前父进程链对应的客户端来源；所有实际业务 HTTP 请求都会在发送前重新探测并覆盖来源 Header
- 当前全部已支持命令都可以通过 `--help` 查看本地帮助，例如 `contract-cli --help`、`contract-cli contract search --help`、`contract-cli help contract upload-file`
- `app` 业务接口后续继续新增时，优先在本文件补充命令矩阵

## 通用约定

### 通用帮助入口

CLI 内置帮助只做本地渲染，不读取 profile、不发 HTTP、不触发自动版本检查。

常用入口：

```bash
contract-cli --help
contract-cli -h
contract-cli help
contract-cli help contract upload-file
contract-cli contract search --help
contract-cli contract get <contract-id> --help
```

帮助内容按命令层级展示：

- 命令组展示 `Commands`
- 叶子命令展示 `Flags`、`Examples`、`Notes`
- `Notes` 只放身份限制、user/app 路由差异、请求体或文件上传关键约束
- 不兼容旧顶层别名，例如 `contract-cli help vendor` 会返回未知 help topic

### 通用身份规则

- `config` 和 `version` 不需要登录态
- `skills list/install` 不需要登录态；通用 `npx skills add qfeius/contract-cli -y -g` 也不依赖 contract-cli 登录态
- `update check` 不需要登录态
- `environment inspect` 不需要登录态，也不发起 HTTP 请求
- `auth login --as user` 走 OAuth 用户授权
- `auth login --as app` 走 `appId + appSecret -> tenant_access_token/internal`
- 为兼容老用户脚本，旧身份值 `--as bot` 仍可使用，运行时等价于 `--as app`；新文档和示例统一使用 `app`
- `contract ...`、`mdm ...` 结构化命令大多默认只支持 `--as user`
- `/open-apis/contract/v1/mcp/...` 路径大多仍只支持 `--as user`
- `contract search-v2`、`contract field update`、`contract sign switch-to-paper`、`contract sign-url get`、`contract form attribute list`、`contract authorization grant`、`contract esign *`、`contract submit`、`contract resubmit`、`contract patch`、`contract download-file`、`contract delete`、`contract print-file`、`contract share get/batch-create`、`contract cooperation link/record/search/file`、`contract approval start/get`、`payment *`、新增写入和扩展查询型 `mdm *`、`event outbound-ip list` 和 `rule table *` 当前仅支持 `--as app`
- `contract get`、`contract search`、`contract create`、`contract sync-user-groups`、`contract text`、`contract category list`、`contract template list`、`contract template get`、`contract template instantiate`、`contract upload-file`、`mdm vendor list`、`mdm vendor get`、`mdm legal list`、`mdm legal get`、`mdm fields list` 是例外：
  - `contract get --as user` 走 MCP 路径 `/open-apis/contract/v1/mcp/contracts/{contract_id}`
  - `contract get --as app` 走开放平台路径 `/open-apis/contract/v1/contracts/{contract_id}`
  - `--as user` 走 MCP 路径 `/open-apis/contract/v1/mcp/contracts/search`
  - `--as app` 走开放平台路径 `/open-apis/contract/v1/contracts/search`
  - `contract create --as user` 走 MCP 路径 `/open-apis/contract/v1/mcp/contracts`
  - `contract create --as app` 走开放平台路径 `POST /open-apis/contract/v1/contracts`
  - `contract sync-user-groups --as user` 走 `/open-apis/contract/v1/mcp/contracts/user-groups/sync?user_id_type=user_id`
  - `contract sync-user-groups --as app` 走 `/open-apis/contract/v1/contracts/user-groups/sync`
  - `contract text --as user` 走 `/open-apis/contract/v1/mcp/contracts/{contract_id}/text?user_id_type=user_id&...`
  - `contract text --as app` 走 `GET /open-apis/contract/v1/contracts/{contract_id}/text?...`
  - `contract category list --as user` 走 `/open-apis/contract/v1/mcp/contract_categorys`
  - `contract category list --as app` 走 `/open-apis/contract/v1/contract_categorys`
  - `contract template list --as user` 走 `/open-apis/contract/v1/mcp/templates`
  - `contract template list --as app` 走 `/open-apis/contract/v1/templates`
  - `contract template get --as user` 走 `/open-apis/contract/v1/mcp/templates/{template_id}`
  - `contract template get --as app` 走 `/open-apis/contract/v1/templates/{template_id}`
  - `contract template instantiate --as user` 走 `/open-apis/contract/v1/mcp/template_instances`
  - `contract template instantiate --as app` 走 `POST /open-apis/contract/v1/template_instances`
  - `contract upload-file --as user` 与 `contract upload-file --as app` 均走 `POST /open-apis/contract/v1/files/upload`
  - `mdm vendor list --as user` 走 `/open-apis/contract/v1/mcp/vendors`
  - `mdm vendor list --as app` 走 `/open-apis/mdm/v1/vendors`
  - `mdm vendor get --as user` 走 `/open-apis/contract/v1/mcp/vendors/{vendor_id}`
  - `mdm vendor get --as app` 走 `/open-apis/mdm/v1/vendors/{vendor_id}`
  - `mdm legal list --as user` 走 `/open-apis/contract/v1/mcp/legal_entities`
  - `mdm legal list --as app` 走 `/open-apis/mdm/v1/legal_entities/list_all`
  - `mdm legal get --as user` 走 `/open-apis/contract/v1/mcp/legal_entities/{legal_entity_id}`
  - `mdm legal get --as app` 走 `/open-apis/mdm/v1/legal_entities/{legal_entity_id}`，并额外透传同名 query `legal_entity_id`
  - `mdm fields list --as user` 走 `/open-apis/contract/v1/mcp/config/config_list`
  - `mdm fields list --as app` 走 `/open-apis/mdm/v1/config/config_list`
- `api call` 是预留能力，当前暂未开放使用；请优先使用已开放的结构化命令

### 通用输出

结构化业务命令共享这些输出参数：

- `--output json|yaml|table`
- `--raw`

默认输出格式是 `json`。

### 通用请求体输入

需要请求体的命令统一使用：

- `--input-file <json-file>`
- `--data '<json-string>'`

约束：

- `--input-file` 与 `--data` 互斥
- `contract create`、`contract template instantiate` 至少需要其一
- `contract search` 可以只传查询 flag，也可以显式传空对象 `{}`，不强制要求 body 输入
- `--file` 只用于真实二进制文件上传，例如 `contract upload-file`
- 不要把 `--file` 当 JSON 请求体输入；JSON 请求体始终用 `--input-file`

### 通用用户标识参数

开放平台命令统一预留了两组通用 query 参数：

- `--user-id-type`
- `--user-id`

当前行为：

- `contract ...`、`mdm ...` 结构化命令会透传到对应底层接口
- `--user-id-type` 不传时默认拼接 `user_id_type=user_id`
- 显式传 `--user-id-type <type>` 时会覆盖默认值
- `--user-id` 传了就拼接到 query string，不传就不带
- 不区分 `user` / `app`
- 除 `mdm vendor create/update` 与 `mdm legal create/update` 外不做命令级校验；这四个 MDM 写接口会本地要求 `--user-id`

## 命令矩阵

### 1. 配置与版本

#### `contract-cli config add`

用途：初始化或更新 profile，并写入 user OAuth 与 app token 的基础配置。

命令：

```bash
contract-cli config add --env prod --name contract
```

支持参数：

- `--env`：环境预设，当前仅支持 `prod`，默认 `prod`
- `--name`：profile 名称，默认 `contract`
- `--resource-metadata-url`：覆盖 protected resource metadata 地址
- `--redirect-url`：覆盖 OAuth callback 地址
- `--scope`：覆盖默认 scope 列表

执行结果：

- 写入 `open_platform_base_url`
- 写入 user OAuth metadata
- 写入 app `app_token_endpoint`
- 将 profile 设为当前 profile

#### `contract-cli version`

用途：查看当前 CLI 版本、commit 和构建时间。

命令：

```bash
contract-cli version
contract-cli --version
```

#### `contract-cli update check`

用途：检查 npm 远端是否存在可升级版本。

命令：

```bash
contract-cli update check
contract-cli update check --channel latest
contract-cli update check --channel latest --json
```

支持参数：

- `--channel`：npm dist-tag；不传时根据当前版本推断，预发布版本默认检查 `beta`，稳定版本默认检查 `latest`
- `--json`：输出飞书式结构化 JSON；默认输出文本提示

执行结果：

- 当前版本是 `dev`、`unknown` 或非语义化版本（例如源码 git hash）时跳过远端检查
- 默认输出文本提示，和飞书 `lark-cli update --check` 的手动校验体验保持一致
- 带 `--json` 时输出顶层 `ok`、`previous_version`、`current_version`、`latest_version`、`action`、`message` 等字段
- 有新版本时 `action=update_available`，并额外包含 `command`，值为 `npm install -g @qfeius/contract-cli@<channel> --registry https://registry.npmjs.org`
- 无新版本时 `action=already_up_to_date`
- 手动 `update check --json` 不注入 `_notice.update`；`_notice.update` 只用于普通 JSON 业务命令的自动提示
- 手动执行 `update check` 会直接访问 npm registry，并把结果写入本机 update cache

自动提示：

- 普通命令会先同步读取当前配置目录的 `update-check.json`；缓存里有可升级版本时，仅在 JSON object 输出中注入 `_notice.update`
- 命中 fresh cache 时不访问 npm registry，因此不会立即发现刚发布的新包
- cache 缺失、channel 不匹配或过期时，当前命令会在短超时内同步刷新远端版本；成功结果会写入当前配置目录的 `update-check.json`
- 网络失败、registry 失败或当前是 dev 构建时不会阻断原命令；刷新失败不会写入失败缓存
- `--raw`、yaml、table、纯文本命令不注入 `_notice.update`
- CI 环境会跳过自动远端检查
- 设置 `CONTRACT_CLI_NO_UPDATE_CHECK=1` 可以关闭自动检查

#### `contract-cli environment inspect`

用途：在本地检查当前 CLI 是由 Doubao、WorkBuddy、Codex 还是未知环境调用。

命令：

```bash
contract-cli environment inspect
contract-cli environment inspect --output json
contract-cli environment inspect --output json --include-processes
```

支持参数：

- `--depth`：父进程最大回溯深度，范围 `1-128`，默认 `32`
- `--output`：`text` 或 `json`，默认 `text`
- `--include-processes`：在本地诊断结果中包含采集到的 PID、PPID、进程名和可执行文件路径；不读取或输出完整命令行参数

业务请求行为：

- 每一次实际业务 HTTP 请求发送前都会重新探测，包括 OpenPlatform Client 的请求重试和 Token 刷新后的业务请求重放
- 识别结果只作用于本次请求，不写入 profile、OAuth Token 或其他持久化配置
- macOS 校验代码签名并匹配 Bundle ID + Team ID
- Windows 优先匹配 Package Family Name；非商店桌面程序通过系统 WinVerifyTrust 校验 Authenticode，并匹配证书 SHA-256 + 可执行文件路径
- Linux 当前按可执行文件路径或进程名降级识别
- 已发现 macOS/Windows 平台身份但身份不匹配时返回 `unknown`，不再降级为路径或进程名命中
- 业务请求只透传 `X-Qfei-Request-Source-Type`、`X-Qfei-Channel-Type`、`X-Qfei-Evidence-Type`、`X-Qfei-Channel-Confidence`、`X-Qfei-Detector-Version` 和 `X-Qfei-Rule-Id`
- 业务 Header 不包含 PID、进程路径、完整命令行或用户目录信息

#### `contract-cli skills list`

用途：列出当前二进制内置的 Codex skills。

命令：

```bash
contract-cli skills list
```

输出内容：

- skill 名称
- skill 版本
- skill 描述

#### `npx skills add qfeius/contract-cli -y -g`

用途：使用通用 `skills` installer 从 GitHub 仓库安装 contract-cli 的 Agent skills。

推荐命令：

```bash
npx skills add qfeius/contract-cli -y -g
```

适用场景：

- 推荐给 Codex、Cursor、Trae、Claude Code 等多类 Agent 环境使用
- 从 GitHub 仓库的 `skills/` 目录安装，适合快速获得最新 skill 文档
- `-g` 表示全局安装，安装位置和平台适配由通用 `skills` installer 决定

注意事项：

- 该命令依赖 npm、npx 和 GitHub 网络访问
- 安装内容来自远程 `qfeius/contract-cli` 仓库，不读取本地未 push 的改动
- 若通用 installer 不可用，使用 `contract-cli skills install` 作为兜底

#### `contract-cli skills install`

用途：将当前二进制内置的 Codex skills 安装到本机 Codex skills 目录，作为通用 `npx skills add ...` 不可用时的兜底方案。

命令：

```bash
contract-cli skills install
contract-cli skills install --target ~/.codex/skills
contract-cli skills install --force
```

支持参数：

- `--target`：安装目标目录；默认优先使用 `$CODEX_HOME/skills`，否则使用 `~/.codex/skills`
- `--force`：覆盖已存在的同名 skill；默认不覆盖，会跳过已有目录

执行结果：

- 复制内置 `auth`、`contract-cli-shared`、`contract-cli-contract`、`contract-cli-payment`、`contract-cli-mdm-vendor`、`contract-cli-mdm-legal`、`contract-cli-mdm-fields` 等 skill
- 保留 `SKILL.md`、`agents/openai.yaml` 和 `references/*.md`

### 2. 鉴权

#### `contract-cli auth login`

##### `contract-cli auth login --as user`

用途：发起 OAuth 用户授权。

命令：

```bash
contract-cli auth login --profile contract --as user
```

支持参数：

- `--profile`
- `--as user`
- `--timeout`
- `--no-open-browser`

##### `contract-cli auth login --as app`

用途：使用 app `appId/appSecret` 直接换取 tenant access token。

命令：

```bash
contract-cli auth login --profile contract --as app --app-id <id> --app-secret <secret>
```

支持参数：

- `--profile`
- `--as app`
- `--app-id`
- `--app-secret`

补充说明：

- app 凭证优先级：flag > env > 已保存 secrets
- 登录成功后会保存 app token，并将默认身份切到 `app`
- `auth logout --as app` 只清 token，不删除 `appId/appSecret`
- 兼容旧命令 `auth login --as bot`，实际按 app 身份登录并写入 `identities.app`

#### `contract-cli auth status`

用途：查看某个 profile 的 user 或 app 身份状态。

命令：

```bash
contract-cli auth status --profile contract --as user
contract-cli auth status --profile contract --as app
```

支持参数：

- `--profile`
- `--as user|app`

当前状态语义：

- user：`authorized` / `unauthorized`
- app：`authorized` / `expired` / `configured` / `unconfigured`

#### `contract-cli auth logout`

用途：清理指定身份的 token。

命令：

```bash
contract-cli auth logout --profile contract --as user
contract-cli auth logout --profile contract --as app
```

支持参数：

- `--profile`
- `--as user|app`

补充说明：

- user logout：清空 user token
- app logout：只清空 app token，保留 app 凭证

#### `contract-cli auth use`

用途：切换 profile 默认业务身份。

命令：

```bash
contract-cli auth use --profile contract --as user
contract-cli auth use --profile contract --as app
```

支持参数：

- `--profile`
- `--as user|app`

### 3. 原始开放平台调用（暂未开放）

`contract-cli api call` 是预留调试入口，当前不对外开放。

当前行为：

- 执行 `contract-cli api ...` 会直接返回：`api call 暂未开放使用，请使用已开放的结构化命令`
- 不读取 profile，不发 HTTP 请求
- 不出现在 `contract-cli --help`、`contract-cli help` 或内置 skills 安装列表中
- 需要开放平台能力时，请优先使用 `contract ...`、`mdm ...` 等结构化命令
- 显式 `--as app` 调用 `contract/v1/mcp` 路径会直接报错

### 4. 合同命令

共享参数：

- `--profile`
- `--as`
- `--output`
- `--raw`
- 需要请求体的命令额外支持 `--input-file` / `--data`
- `contract upload-file` 额外支持 `--file` / `--file-type` / `--file-name`
- `contract download-file` 额外支持 `--output-file` / `--force`

#### `contract-cli contract search`

用途：搜索合同。

命令：

```bash
contract-cli contract search --profile contract --as user --input-file search.json
contract-cli contract search --profile contract --as app --input-file search.json
contract-cli contract search --profile contract --as app --input-file search.json --user-id ou_xxx --user-id-type employee_id
```

支持参数：

- `--contract-number`
- `--page-size`
- `--page-token`
- `--input-file`
- `--data`
- `--user-id-type`
- `--user-id`

按身份路由：

- `--as user`：
  - 走 `/open-apis/contract/v1/mcp/contracts/search`
- `--as app`：
  - 走 `/open-apis/contract/v1/contracts/search`
- 额外传入 `--user-id-type` / `--user-id` 时，会原样拼到 query string
- 未显式传 `--as` 时：
  - 若 profile 默认身份是 `app`，则会直接走 app 搜索路由
  - 若 profile 默认身份是 `user`，则走 user 搜索路由

#### `contract-cli contract search-v2`

用途：app 身份搜索合同 V2，复杂条件直接透传 JSON body。

命令：

```bash
contract-cli contract search-v2 --profile contract --as app --input-file search-v2.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/searchV2`。
- `--input-file` / `--data` 必须传一个且互斥。

#### `contract-cli contract get`

用途：获取合同详情。

命令：

```bash
contract-cli contract get <contract-id> --profile contract --as user
contract-cli contract get <contract-id> --profile contract --as app
contract-cli contract get <contract-id> --profile contract --as app --user-id ou_xxx --user-id-type employee_id
```

支持参数：

- `--profile`
- `--as`
- `--output`
- `--raw`
- `--user-id-type`
- `--user-id`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/contracts/{contract_id}`
- `--as app`
  - 走 `/open-apis/contract/v1/contracts/{contract_id}`
- 额外传入 `--user-id-type` / `--user-id` 时，会原样拼到 query string

#### `contract-cli contract sync-user-groups`

用途：同步用户分组。

命令：

```bash
contract-cli contract sync-user-groups --profile contract --as user
contract-cli contract sync-user-groups --profile contract --as app
contract-cli contract sync-user-groups --profile contract --as app --user-id ou_xxx
```

支持参数：

- `--profile`
- `--as`
- `--output`
- `--raw`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/contracts/user-groups/sync?user_id_type=user_id`
- `--as app`
  - 走 `/open-apis/contract/v1/contracts/user-groups/sync`
  - 额外传入 `--user-id-type` / `--user-id` 时，会原样拼到 query string

#### `contract-cli contract text`

用途：获取合同文本。

命令：

```bash
contract-cli contract text <contract-id> --profile contract --as user
contract-cli contract text <contract-id> --profile contract --as app
contract-cli contract text <contract-id> --profile contract --as app --user-id-type employee_id
```

支持参数：

- `--profile`
- `--as`
- `--output`
- `--raw`
- `--full-text`
- `--offset`
- `--limit`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/contracts/{contract_id}/text?user_id_type=user_id&...`
- `--as app`
  - 走 `GET /open-apis/contract/v1/contracts/{contract_id}/text?...`
  - 额外传入 `--user-id-type` / `--user-id` 时，会原样拼到 query string

#### `contract-cli contract create`

用途：创建合同。

命令：

```bash
contract-cli contract create --profile contract --input-file create.json
contract-cli contract create --profile contract --data '{"title":"demo"}'
contract-cli contract create --profile contract --as app --data '{"contract_name":"demo","create_user_id":"ou_xxx"}'
```

支持参数：

- `--input-file`
- `--data`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/contracts`
- `--as app`
  - 走 `POST /open-apis/contract/v1/contracts`
  - 请求体需要自己带上 `create_user_id`
  - 额外传入 `--user-id-type` / `--user-id` 时，会原样拼到 query string

字段参考：

- [create-contract-fields.md](../skills/contract-cli-contract/references/create-contract-fields.md)
- [create-contract-field-tree.md](../skills/contract-cli-contract/references/create-contract-field-tree.md)
- [create-contract-enums.md](../skills/contract-cli-contract/references/create-contract-enums.md)

#### `contract-cli contract field update`

用途：app 身份更新合同字段信息，目前主要用于修改下拉列表选项范围。

命令：

```bash
contract-cli contract field update --profile contract --as app --input-file field-update.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `PUT /open-apis/contract/v1/attribute_definition`。
- 最小 body 通常包含 `module_name`、`attribute_name`、`value_scopes`。

#### `contract-cli contract sign switch-to-paper`

用途：app 身份将电子签合同转为纸质签。

命令：

```bash
contract-cli contract sign switch-to-paper --profile contract --as app --business-id <contract-id> --business-type-code 0
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/signType/switchToPaper`。
- `--business-id` 和 `--business-type-code` 必填，不接受 `--input-file` / `--data`。

#### `contract-cli contract sign-url get`

用途：app 身份获取合同签署链接。

命令：

```bash
contract-cli contract sign-url get <contract-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/sign_url`。
- 不接受 `--input-file` / `--data`。

#### `contract-cli contract form attribute list`

用途：app 身份按合同类型和流程类型获取合同流程字段。

命令：

```bash
contract-cli contract form attribute list --profile contract --as app --category-id <category-id> --business-type-code 0
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/form_definition/attribute`。
- `--business-type-code`：0 申请、1 变更、2 终止、3 合同组申请。

#### `contract-cli contract authorization grant`

用途：app 身份授予合同权限。

命令：

```bash
contract-cli contract authorization grant --profile contract --as app --input-file authorization.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/authorizations`。
- 最小 body 通常包含 `business_id`、`authorized_user_id`、`start_time`、`end_time`。

#### `contract-cli contract esign personal-auth-url`

用途：app 身份获取个人认证和授权页面链接。

命令：

```bash
contract-cli contract esign personal-auth-url --profile contract --as app --input-file psn-auth-url.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/esign/auth/psnAuthUrl`。
- 最小 body 需要包含 `psnAuthConfig`。

#### `contract-cli contract esign org-auth-url`

用途：app 身份获取机构认证和授权页面链接。

命令：

```bash
contract-cli contract esign org-auth-url --profile contract --as app --input-file org-auth-url.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/esign/auth/orgAuthUrl`。
- 最小 body 需要包含 `orgAuthConfig`。

#### `contract-cli contract upload-file`

用途：上传合同相关文件，返回后端原始 JSON，重点关注 `data.file_id`。

命令：

```bash
contract-cli contract upload-file --profile contract --as user --file ./合同正文.docx --file-type text
contract-cli contract upload-file --profile contract --as app --file ./附件.pdf --file-type attachment --file-name 附件.pdf
```

支持参数：

- `--file`：必填，本地待上传文件路径。
- `--file-type`：必填，透传后端文件类型。
- `--file-name`：可选；不传时默认使用本地文件名。
- `--user-id-type`
- `--user-id`

身份规则：

- `--as user` 和 `--as app` 均支持。
- 走 `POST /open-apis/contract/v1/files/upload`。
- 请求是 `multipart/form-data`，字段为 `file_name`、`file_type`、`file`。
- 不接受 `--input-file` / `--data`；这两个参数只用于 JSON 请求体。

本地校验：

- `--file` 必须存在且是普通文件。
- 文件大小必须小于等于 `200MB`。
- CLI 不在本地校验扩展名白名单，扩展名和 `file_type` 合法性由后端最终校验。

常用 `file_type`：

- `text`：合同文本。
- `attachment`：其他附件。
- `scan`：归档扫描件。
- `cause`：合同附件。
- `archiveAttachment`：归档附件。
- `customPictureAttachment` / `customTableAttachment` / `customFileAttachment`：自定义附件。

#### `contract-cli contract submit`

用途：app 身份提交合同。

命令：

```bash
contract-cli contract submit <contract-id> --profile contract --as app
contract-cli contract submit <contract-id> --profile contract --as app --data '{"comment":"ok"}'
```

支持参数：

- `--input-file`：可选，透传 JSON 请求体。
- `--data`：可选，透传 JSON 请求体。
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/{contract_id}/submit`。
- 不传 `--input-file` / `--data` 时不发送请求体。

#### `contract-cli contract resubmit`

用途：app 身份重新提交合同。

命令：

```bash
contract-cli contract resubmit <contract-id> --profile contract --as app
contract-cli contract resubmit <contract-id> --profile contract --as app --input-file resubmit.json
```

支持参数：

- `--input-file`：可选，透传 JSON 请求体。
- `--data`：可选，透传 JSON 请求体。
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/{contract_id}/resubmit`。
- 不传 `--input-file` / `--data` 时不发送请求体。

#### `contract-cli contract patch`

用途：app 身份更新合同。

命令：

```bash
contract-cli contract patch <contract-id> --profile contract --as app --input-file patch.json
contract-cli contract patch <contract-id> --profile contract --as app --data '{"title":"demo"}'
```

支持参数：

- `--input-file`
- `--data`
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `PATCH /open-apis/contract/v1/contracts/{contract_id}`。
- `--input-file` / `--data` 必须传一个且互斥。

#### `contract-cli contract download-file`

用途：app 身份下载合同相关文件。

命令：

```bash
contract-cli contract download-file <file-id> --profile contract --as app
contract-cli contract download-file <file-id> --profile contract --as app --output-file ./contract.pdf
contract-cli contract download-file <file-id> --profile contract --as app --raw > contract.pdf
```

支持参数：

- `--output-file`：保存到指定文件；不传时默认拉起保存文件弹窗。
- `--force`：覆盖已存在的 `--output-file`。
- `--raw`：把文件内容写到 stdout，不打印额外提示。
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/files/{file_id}`。
- 不实现 `dowload-file` 拼写别名。
- 无 GUI、远程、CI、Agent 环境推荐显式传 `--output-file`。

#### `contract-cli contract delete`

用途：app 身份删除草稿合同。

命令：

```bash
contract-cli contract delete <contract-id> --profile contract --as app
```

支持参数：

- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `DELETE /open-apis/contract/v1/contracts/{contract_id}`。
- 命令直接删除，不额外要求 `--yes`。

#### `contract-cli contract print-file`

用途：app 身份生成合同打印文件。

命令：

```bash
contract-cli contract print-file --profile contract --as app --input-file print-file.json
contract-cli contract print-file --profile contract --as app --data '{"contract_id":"<contract-id>"}'
```

支持参数：

- `--input-file`
- `--data`
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/files`。
- `--input-file` / `--data` 必须传一个且互斥。

#### `contract-cli contract share get`

用途：app 身份查询合同分享记录。

命令：

```bash
contract-cli contract share get <contract-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/share_records`。

#### `contract-cli contract share batch-create`

用途：app 身份批量分享合同。

命令：

```bash
contract-cli contract share batch-create --profile contract --as app --input-file batch-share.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/contract/batch_share`。
- 最小 body 通常包含 `contract_id` 和 `user_ids`。

#### `contract-cli contract cooperation link get`

用途：app 身份查询合同协商邀请链接。

命令：

```bash
contract-cli contract cooperation link get <contract-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/cooperation_link`。

#### `contract-cli contract cooperation record get`

用途：app 身份查询合同协商操作记录信息。

命令：

```bash
contract-cli contract cooperation record get <contract-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/cooperation_record_info`。

#### `contract-cli contract cooperation search`

用途：app 身份查询协商列表。

命令：

```bash
contract-cli contract cooperation search --profile contract --as app --input-file cooperation-search.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/cooperation/search`。
- 最小 body 需要包含 `user_id`，分页可放在 body 的 `page_size` / `page_token`。

#### `contract-cli contract cooperation file get`

用途：app 身份查询合同协商文件信息。

命令：

```bash
contract-cli contract cooperation file get <contract-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/cooperation/file_info`。

#### `contract-cli contract cooperation file download`

用途：app 身份下载合同协商文件。

命令：

```bash
contract-cli contract cooperation file download <file-id> --profile contract --as app --output-file ./cooperation.docx
contract-cli contract cooperation file download <file-id> --profile contract --as app --raw > cooperation.docx
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/cooperation/{file_id}/download_file`。
- `--raw` 会把二进制内容写到 stdout，不打印额外提示。

#### `contract-cli contract approval start`

用途：app 身份发起流程审批。

命令：

```bash
contract-cli contract approval start <process-instance-id> --profile contract --as app --input-file approval.json
```

支持参数：

- `--input-file`
- `--data`
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/process_instances/{process_instance_id}/task_approval`。
- `--input-file` / `--data` 必须传一个且互斥。

#### `contract-cli contract approval get`

用途：app 身份查询审批实例详情。

命令：

```bash
contract-cli contract approval get <process-instance-id> --profile contract --as app
contract-cli contract approval get <process-instance-id> --profile contract --as app --notice-filter notice_filter --task-instance-filter task_instance_filter
```

支持参数：

- `--notice-filter`
- `--task-instance-filter`
- `--user-id-type`
- `--user-id`

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/process_instances/{process_instance_id}`。
- 不接受 `--input-file` / `--data`。

#### `contract-cli contract category list`

用途：列出合同分类。

命令：

```bash
contract-cli contract category list --profile contract
contract-cli contract category list --profile contract --as app --lang zh-CN
```

支持参数：

- `--lang`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/contract_categorys`
- `--as app`
  - 走 `/open-apis/contract/v1/contract_categorys`

#### `contract-cli contract template list`

用途：列出模板。

命令：

```bash
contract-cli contract template list --profile contract
contract-cli contract template list --profile contract --as app --category-number CAT-1 --page-size 20 --user-id ou_xxx --user-id-type employee_id
```

支持参数：

- `--category-number`
- `--page-size`
- `--page-token`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/templates`
- `--as app`
  - 走 `/open-apis/contract/v1/templates`
  - 按生产文档，`category_number`、`user_id`、`user_id_type` 都属于 app 接口查询参数
  - CLI 继续按现有约定只透传，不做本地必填校验

#### `contract-cli contract template get`

用途：获取模板详情。

命令：

```bash
contract-cli contract template get <template-id> --profile contract
contract-cli contract template get <template-id> --profile contract --as app --user-id ou_xxx --user-id-type employee_id
```

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/templates/{template_id}`
- `--as app`
  - 走 `/open-apis/contract/v1/templates/{template_id}`
  - 按生产文档，`user_id`、`user_id_type` 都属于 app 接口查询参数
  - CLI 继续按现有约定只透传，不做本地必填校验

#### `contract-cli contract template instantiate`

用途：创建模板实例。

命令：

```bash
contract-cli contract template instantiate --profile contract --input-file template-instance.json
contract-cli contract template instantiate --profile contract --as app --data '{"template_number":"TMP001","create_user_id":"ou_xxx"}' --user-id-type employee_id
```

支持参数：

- `--input-file`
- `--data`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/template_instances`
- `--as app`
  - 走 `POST /open-apis/contract/v1/template_instances`
  - 按生产文档，query 里只有 `user_id_type`，请求体里需要 `create_user_id`
  - CLI 继续按现有约定只透传，不做本地必填校验

#### `contract-cli contract enum list`

用途：查询枚举值。

命令：

```bash
contract-cli contract enum list --profile contract --type contract_status
```

支持参数：

- `--type`

### 5. 付款命令

`payment` 这一组命令当前全部仅支持 `--as app`。命令参数采用“主操作对象 ID 用位置参数，父资源 ID 用 flag”的方式。

#### `contract-cli payment create`

用途：创建付款申请。

命令：

```bash
contract-cli payment create --contract <contract-id> --profile contract --as app --input-file payment.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/{contract_id}/payments`。
- `--contract` 必填。
- `--input-file` / `--data` 必填且互斥。

#### `contract-cli payment update`

用途：更新付款信息。

命令：

```bash
contract-cli payment update <payment-id> --contract <contract-id> --profile contract --as app --input-file payment-update.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `PATCH /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}`。
- `--contract` 必填。
- `--input-file` / `--data` 必填且互斥。

#### `contract-cli payment get`

用途：查看付款信息。

命令：

```bash
contract-cli payment get <payment-id> --contract <contract-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}`。
- `--contract` 必填。
- 不接受 `--input-file` / `--data`。

#### `contract-cli payment list`

用途：查询付款申请列表。

命令：

```bash
contract-cli payment list --contract <contract-id> --profile contract --as app
contract-cli payment list --contract <contract-id> --profile contract --as app --page-size 10 --page-token next
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/payments`。
- `--contract` 必填。
- `--page-size` / `--page-token` 可选。
- 不接受 `--input-file` / `--data`。

#### `contract-cli payment plan notify`

用途：同步付款记录。

命令：

```bash
contract-cli payment plan notify --profile contract --as app --input-file notify.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/payment/notify`。
- `--input-file` / `--data` 必填且互斥。

#### `contract-cli payment plan search`

用途：搜索付款计划。

命令：

```bash
contract-cli payment plan search --profile contract --as app --input-file payment-plan-search.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/payments/search`。
- `--input-file` / `--data` 必填且互斥。

#### `contract-cli payment record create`

用途：创建付款记录。

命令：

```bash
contract-cli payment record create --contract <contract-id> --payment <payment-id> --profile contract --as app --input-file payment-record.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `POST /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}/payment_records`。
- `--contract` / `--payment` 必填。
- `--input-file` / `--data` 必填且互斥。

#### `contract-cli payment record update`

用途：更新付款记录。

命令：

```bash
contract-cli payment record update <payment-record-id> --contract <contract-id> --payment <payment-id> --profile contract --as app --input-file payment-record-update.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `PATCH /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}/payment_records/{payment_record_id}`。
- `--contract` / `--payment` 必填。
- `--input-file` / `--data` 必填且互斥。

#### `contract-cli payment record get`

用途：查询付款记录详情。

命令：

```bash
contract-cli payment record get <payment-record-id> --contract <contract-id> --payment <payment-id> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/{contract_id}/payments/{payment_id}/payment_records/{payment_record_id}`。
- `--contract` / `--payment` 必填。
- 不接受 `--input-file` / `--data`。

#### `contract-cli payment record list`

用途：根据付款计划 ID 查询付款记录。

命令：

```bash
contract-cli payment record list --plan <payment-plan-uuid> --profile contract --as app
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/contract/v1/contracts/payments/{payment_plan_uuid}/payment_records`。
- `--plan` 必填。
- 不接受 `--input-file` / `--data`。

### 6. MDM 命令

这一组命令里，当前 `mdm vendor list`、`mdm vendor get`、`mdm legal list`、`mdm legal get` 和 `mdm fields list` 同时支持 `user` 与 `app`。

共享参数：

- `--profile`
- `--as`
- `--output`
- `--raw`

#### `contract-cli mdm vendor list`

用途：查询交易方列表。

命令：

```bash
contract-cli mdm vendor list --profile contract --name 供应商 --page-size 10
```

支持参数：

- `--name`
- `--page-size`
- `--page-token`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/vendors`
  - 当前仍保留既有 MCP 查询行为
- `--as app`
  - 走 `/open-apis/mdm/v1/vendors`
  - 生产文档把 query `vendor` 描述成“供应商编码”
  - CLI 继续沿用现有 `--name -> vendor` 的透传映射，不在本地改名，也不做额外校验

#### `contract-cli mdm vendor get`

用途：查询交易方详情。

命令：

```bash
contract-cli mdm vendor get <vendor-id> --profile contract
```

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/vendors/{vendor_id}`
- `--as app`
  - 走 `/open-apis/mdm/v1/vendors/{vendor_id}`
  - 生产文档里 query 只看到 `user_id_type`
  - CLI 继续按共享约定透传 `--user-id-type` / `--user-id`，不做本地校验

#### `contract-cli mdm legal list`

用途：查询法人主体列表。

命令：

```bash
contract-cli mdm legal list --profile contract --name 主体A --page-size 10
```

支持参数：

- `--name`
- `--page-size`
- `--page-token`

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/legal_entities`
- `--as app`
  - 走 `/open-apis/mdm/v1/legal_entities/list_all`
  - 生产文档显示文本使用 `legal_entities/list_all`，但超链接目标误指到了 `vendors`
  - 文档还写了“查询参数采用驼峰式”，但当前 CLI 继续沿用既有 `legalEntity/page_size/page_token` 透传映射，不在本地改名

#### `contract-cli mdm legal get`

用途：查询法人主体详情。

命令：

```bash
contract-cli mdm legal get <legal-entity-id> --profile contract
contract-cli mdm legal get --profile contract --as app --code L0001 --page-size 10
```

支持参数：

- `<legal-entity-id>`：按 ID 查询详情
- `--code`：按法人实体编码查询，映射到底层 query `legalEntity`
- `--page-size`：仅 `--code` 模式可用
- `--page-token`：仅 `--code` 模式可用

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/legal_entities/{legal_entity_id}`
- `--as app`
  - 按 ID 查询走 `/open-apis/mdm/v1/legal_entities/{legal_entity_id}`
  - 按这次确认方案，除了 path 参数外，还会额外拼接同名 query `legal_entity_id`
  - 文档里把 `legal_entity_id` 放在查询参数表里，因此 CLI 按“path + query 双带”的方式实现
  - 传 `--code` 时走 `GET /open-apis/mdm/v1/legal_entities`
  - `--code` 模式不同于 `mdm legal list --as app` 的 `/open-apis/mdm/v1/legal_entities/list_all`

#### `contract-cli mdm fields list`

用途：查询字段配置。

命令：

```bash
contract-cli mdm fields list --profile contract --biz-line vendor
```

支持参数：

- `--biz-line`

当前支持的典型值：

- `vendor`
- `legal_entity`：user 原样透传；app 会自动映射为 `legalEntity`
- `vendor_risk`：仅 user/MCP 路径可用，app 当前不支持

身份规则：

- `--as user`
  - 走 `/open-apis/contract/v1/mcp/config/config_list`
  - `--biz-line` 可传 `vendor`、`legal_entity`、`vendor_risk`
- `--as app`
  - 走 `/open-apis/mdm/v1/config/config_list`
  - 文档显示文本就是这条路径，但超链接目标误指到了 `vendors`
  - 后端当前只接受 `vendor` 或 `legalEntity`
  - CLI 允许继续传 `legal_entity`，并在 app 路由下自动映射为 `legalEntity`
  - `vendor_risk` 在 app 身份下会被本地拒绝，不再发送请求

#### `contract-cli mdm vendor create`

用途：app 身份创建交易方。

命令：

```bash
contract-cli mdm vendor create --profile contract --as app --user-id <operator-user-id> --input-file vendor-create.json
```

身份规则：

- 当前仅支持 `--as app`。
- 必须传 `--user-id`，用于提供当前操作人上下文。
- 走 `POST /open-apis/mdm/v1/vendors`。
- 创建请求体不要包含后端生成的 `vendor` 编码。
- 字段是否必填受后台动态配置影响，可先查 `mdm fields list --biz-line vendor`。

#### `contract-cli mdm vendor update`

用途：app 身份按 ID 更新交易方。

命令：

```bash
contract-cli mdm vendor update <vendor-id> --profile contract --as app --user-id <operator-user-id> --input-file vendor-update.json
```

身份规则：

- 当前仅支持 `--as app`。
- 必须传 `--user-id`，用于提供当前操作人上下文。
- 走 `PUT /open-apis/mdm/v1/vendors/{vendor_id}`。
- 请求体必须包含后端返回的 `id` 和 `vendor` 编码；CLI 不在本地补动态字段。

#### `contract-cli mdm vendor list-all`

用途：app 身份分页查询交易方全量数据。

命令：

```bash
contract-cli mdm vendor list-all --profile contract --as app --page-size 10 --page-token next
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/mdm/v1/vendors/list_all`。
- 不接受 `--input-file` / `--data`。

#### `contract-cli mdm vendor query-by-cert`

用途：app 身份根据证件 ID 和国家地区精确查询交易方。

命令：

```bash
contract-cli mdm vendor query-by-cert --profile contract --as app --certification-id 91110105 --ad-country CN
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/mdm/v1/vendors/query_vendors`。
- `--certification-id` 和 `--ad-country` 必填。

#### `contract-cli mdm legal create`

用途：app 身份创建法人主体。

命令：

```bash
contract-cli mdm legal create --profile contract --as app --user-id <operator-user-id> --input-file legal-create.json
```

身份规则：

- 当前仅支持 `--as app`。
- 必须传 `--user-id`，用于提供当前操作人上下文。
- 走 `POST /open-apis/mdm/v1/legal_entities`。
- 创建请求体不要包含后端生成的 `legalEntity` / `legal_entity` 编码。
- 字段是否必填受后台动态配置影响，可先查 `mdm fields list --biz-line legal_entity`。

#### `contract-cli mdm legal update`

用途：app 身份按 ID 更新法人主体。

命令：

```bash
contract-cli mdm legal update <legal-entity-id> --profile contract --as app --user-id <operator-user-id> --input-file legal-update.json
```

身份规则：

- 当前仅支持 `--as app`。
- 必须传 `--user-id`，用于提供当前操作人上下文。
- 走 `PUT /open-apis/mdm/v1/legal_entities/{legal_entity_id}`。
- 请求体必须包含后端返回的 `id` 和 camelCase `legalEntity` 编码，不要写成 `legal_entity`；CLI 不在本地补动态字段。

#### `contract-cli mdm fixed-exchange-rate get`

用途：app 身份查询固定汇率。

命令：

```bash
contract-cli mdm fixed-exchange-rate get --profile contract --as app --source-currency CNY --target-currency USD --effective-date 2026-06-01
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/mdm/v1/fixed_exchange_rate`。
- CLI flag `--effective-date` 会映射到底层 query 参数 `date`。
- 不接受 `--input-file` / `--data`。

#### `contract-cli mdm fixed-exchange-rate update`

用途：app 身份新增或更新固定汇率。

命令：

```bash
contract-cli mdm fixed-exchange-rate update --profile contract --as app --input-file fixed-exchange-rate.json
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `PUT /open-apis/mdm/v1/fixed_exchange_rate`。
- 请求体直接透传。

#### `contract-cli mdm file download`

用途：app 身份下载主数据附件。

命令：

```bash
contract-cli mdm file download <file-id> --profile contract --as app --output-file ./attachment.bin
contract-cli mdm file download <file-id> --profile contract --as app --raw > attachment.bin
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/mdm/v1/file/download/{file_id}`。
- `--raw` 会把二进制内容写到 stdout，不打印额外提示。

### 7. 事件命令

#### `contract-cli event outbound-ip list`

用途：app 身份分页查询开放平台事件出口 IP。

命令：

```bash
contract-cli event outbound-ip list --profile contract --as app --page-size 10 --page-token next
```

身份规则：

- 当前仅支持 `--as app`。
- 走 `GET /open-apis/event/v1/outbound_ip`。
- `--page-size` 可选，传入时必须在 `10` 到 `50` 之间。
- 不接受 `--input-file` / `--data`。

### 8. 审批矩阵规则表命令

这一组命令当前全部仅支持 `--as app`，公共定位参数是 `--product-id`、`--group-id`，涉及单表时再传 `--table-id`。

| 命令 | 方法与路径 | 请求体 |
| --- | --- | --- |
| `rule table list` | `GET /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables` | 不接受 |
| `rule table pre-release` | `PATCH /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/pre_release` | 可选 |
| `rule table release` | `PATCH /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/release` | 可选 |
| `rule table column-headers list` | `GET /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_columns/column_headers` | 不接受 |
| `rule table row create` | `POST /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_rows` | 必填 |
| `rule table row get` | `GET /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_rows/{table_row_id}` | 不接受 |
| `rule table row list` | `GET /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_rows` | 不接受 |
| `rule table row search` | `POST /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_rows/search` | 必填 |
| `rule table row update` | `PUT /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_rows/{table_row_id}` | 必填 |
| `rule table row delete` | `DELETE /open-apis/rule_engine/v1/products/{product_id}/groups/{group_id}/rule_tables/{rule_table_id}/table_rows/{table_row_id}` | 不接受 |

示例：

```bash
contract-cli rule table list --profile contract --as app --product-id <product-id> --group-id <group-id> --page-size 10
contract-cli rule table row create --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id> --input-file row.json
contract-cli rule table row delete <row-id> --profile contract --as app --product-id <product-id> --group-id <group-id> --table-id <table-id>
```

## 后续扩展 app 接口时的建议落点

- 新增 app 业务接口时，优先直接沉淀成结构化命令，避免把预留的 `api call` 暴露给最终用户
- 如需临时验证开放平台路径和鉴权，建议在本地测试或开发工具里完成，不把验证入口写入公开文档
- 一旦新增结构化 app 命令，先更新本文档的“命令矩阵”和“身份规则”，再补实现与测试
- 如果未来同一命令同时支持 user 和 app，需要在文档里明确写出路径差异、参数差异和默认身份规则
