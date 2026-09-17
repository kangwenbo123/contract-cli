# 交易方维护专项 AI 变更记录

## 2026-09-16：交易方 CLI 维护能力

- `mdm vendor create` 按实际身份分流：App 保持调用 `POST /open-apis/mdm/v1/vendors` 并要求 `--user-id`；User 调用 `POST /open-apis/contract/v1/mcp/vendors`，操作人只来自认证身份。
- 新增 `mdm vendor patch <vendor-id>`：App 调用 V1 App PATCH，User 调用个人 PATCH。两条链路均保留原始 JSON 字节；ID 只允许放在 path，个人额外拒绝编码、状态、风险、系统字段和只读的 `vendorAccounts[].bankId`。
- 新增 User-only `mdm vendor enable|disable <vendor-id>`，分别生成 `target_status=1/0`；旧 `mdm vendor update` 继续调用原 V1 PUT，命令、参数和接口保持不变。
- 个人创建和 PATCH 支持可选 `--department-id-type department_id|open_department_id`。参数缺省时不发送；非法枚举和 App 身份误用在 HTTP 请求前拒绝。
- 所有新增写请求明确标记为 WRITE。业务失败原样返回；网络错误、5xx、畸形响应或 `UNKNOWN` 不自动重放，并提示先查询最终状态。

## 2026-09-16：交易方 Skill 与安装流程

- 扩展 `contract-cli-mdm-vendor`，覆盖简单/复杂创建、普通字段 PATCH、子项增改删、附件整体替换、部门查询后写入、启停和结果未知后的查询确认。
- 固定自然语言决策：字段不确定先查字段配置，子项 ID 不确定先查详情；只允许选择 `status=1` 的启用部门，停用部门不得写入；目标或删除范围存在歧义时先询问用户。
- 明确 Skill 只指导 Agent 执行 `contract-cli`，不调用或增强远程 MCP Server，不在 CLI 与 MCP 两条链路之间切换执行方式。
- 新增同包命令 `contract-cli-setup`。CLI、Skill 和 references 由同一 npm 包分发；已有 Skill 默认保留并输出明确同步命令，只有显式 `--force` 才覆盖。
- 正式包版本仍为 `1.8.3`，生产 profile 和生产域名硬锁保持不变；本阶段没有发布 npm 包，也没有修改 MDM、CLM、open-platform、Higress 或数据库结构。

## 本地验证

- `go test ./...`、`go test -race ./...`、`go vet ./...` 通过。
- `npm run test:installer` 5/5 通过，覆盖 setup 委托、默认不强制覆盖和缺少二进制时的明确失败。
- `tests/release/package-dry-run.sh` 通过，npm 清单包含 94 个文件。
- `tests/release/local-install.sh` 和 `make release-check` 通过，覆盖 11 个 Skill 的首次安装、用户修改保留、明确同步提示及 `--force` 覆盖。
- GoReleaser snapshot 构建通过，生成 macOS、Linux、Windows 的 amd64/arm64 产物；现有归档配置有上游弃用提示，但不影响本次构建。

## Test 环境验收

- 从正式功能提交创建本地 `feature/VendorApi-test-e2e` 临时分支，通过 `test_e2e` 编译标签生成 `1.8.3-test-e2e`。该构建只允许 `test-open.qtech.cn` 和 `test-myaccount.qtech.cn`，实测会在发起网络请求前拒绝 Prod 环境。
- 配置、凭证、二进制、HOME、`CODEX_HOME` 和 Skill 全部位于 `/tmp/contract-cli-vendor-test-e2e`。正式 `~/.contract-cli`、全局 CLI 和用户现有 Skill 未被覆盖；Test 适配分支只保留在本地，不推送、不进入正式功能提交。
- Test User OAuth 登录成功。已验证字段配置、交易方列表和详情查询、基础及复杂创建、普通字段修改和清空、联系人及地址增改删、同值 `NO_CHANGE`、停用/重复停用/停用后编辑/启用，以及必填空串、纯空 PATCH、个人修改状态、非法部门类型、不存在部门和不存在交易方等失败场景。
- 已验证 `open_department_id` 写入链路：选择启用部门 `od-8be71c9a48d30ac9e11d7f2d80878352` 创建和修改成功，数据库保存内部部门 ID `1075812179120227417`；内部数字 ID 在不传类型时仍可使用。
- Test 数据为 `1180207945539912527`、`1180208448319521653`、`1180208582285591413`，对应编码 `V00175007`、`V00175008`、`V00175009`。数据库核对主数据、联系人、地址、账户密文、删除标记和状态联动符合预期，验收结束后三条交易方均保持停用，未删除审计历史。
- CLI-only Agent 在隔离 HOME 与 `CODEX_HOME`、不连接 `contract-group` MCP 的条件下完成查询、PATCH、启用、重复启用和停用；命令选择、先查后写及 `APPLIED/NO_CHANGE` 解释符合 Skill 约束。

## 未关闭门禁与存量问题

- Test App 凭证未配置，因此 App PATCH 和旧 App PUT 的 Test 实网验收未完成；未复用 Dev 或生产凭证绕过该门禁。
- 当前 CLI 没有部门列表查询命令。此次为验证部门写入，使用 Test 个人令牌直接调用既有只读部门接口取得启用部门；“自然语言先查询部门再写入”尚不能完全由 CLI-only 链路完成，后续需单独确认是否新增部门查询命令。
- 个人详情接口未返回 `ownerDepts`，但数据库已正确保存部门 ID；该现象属于存量查询响应能力，不由本次 CLI 写入造成。
- Test 后端历史/缓存日志会记录解密后的测试账户和电话，属于现有服务日志安全债务，不在本次 CLI 代码范围内，需单独治理。
- macOS 本机测试和多平台构建已通过；Linux、Windows 及不同架构的实机运行验收按计划留到正式发布阶段。

## 2026-09-17：Test 验收包 Device 授权修复

- 修复 Test E2E 构建在 Device Grant 返回后仍按生产授权域名校验 `verification_uri_complete` 的问题。Test 构建只接受 `test-myaccount.qtech.cn`，正式构建仍只接受 `myaccount.qfei.cn`，未放宽未知域名。
- 修复 Test E2E 构建 `config add` 的默认环境仍为 `prod`、帮助文案与实际支持环境不一致的问题；Test 构建默认 `test`，正式构建默认 `prod`。
- 新增 Test 构建回归测试，覆盖合法 Test 授权链接放行、生产授权链接拒绝、Test 默认环境及帮助文案。
- 隔离目录实网验证 `config add` 成功，`auth init` 正常返回 `pending`；未执行 `auth complete`，未产生已授权用户 Token。
