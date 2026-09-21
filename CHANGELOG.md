# Changelog

## Unreleased

- 为每个 OpenPlatform 逻辑请求生成 W3C `traceparent`，同时发送同值 `X-Log-Id`；重试复用 `trace_id`、每次 HTTP attempt 使用新 `span_id`，失败信息包含可检索的 `trace_id`
- 新增逐请求调用环境识别：实际业务 HTTP 请求发送前回溯父进程；macOS 校验 Bundle ID + Team ID，Windows 校验 Package Family Name 或 Authenticode 证书指纹，并透传归一化来源 Header
- 新增 `contract-cli environment inspect` 本地诊断命令；识别结果不写入 profile 或 OAuth Token
- 新增 `internal/build`，支持 `contract-cli version` 与 `--version`
- 新增 `build.sh`、`Makefile`、`.goreleaser.yml`
- 新增 npm/npx 薄包装：`package.json`、`scripts/install.js`、`scripts/run.js`
- 新增 `tests/cli_e2e/smoke.sh` 作为发布前冒烟脚本

## 1.9.2-beta.1

- 优化合同搜索 Skill 的自动发现描述和 `agents/openai.yaml`，覆盖客户常用的自然语言搜索表达。
- 调整搜索意图路由：显式关键词和普通搜索词优先复现页面默认“全部”搜索；可合理推测人员、部门、主体或自定义字段时保留候选发现与业务引导。
- 指定人员、部门、交易方和我方主体时优先查询候选；自定义字段先发现字段元数据，仅在仍有业务歧义时询问。
- 保持 user/app 搜索分流、完整分页以及合同组与合同条目的结果口径。

## 1.9.1-beta.1

- 新增合同搜索字段发现，以及独立的人员、部门候选查询命令和 Skill。
- 合同搜索 Skill 增加身份分流总纲、精确对象筛选及经过生产对照的场景配方。
- 修复 user 合同搜索业务失败的退出状态、JSON 数值保真与搜索 ID 类型校验。
- 保留 1.9.0-beta.1 的本地 Device 授权复用能力，内嵌 Skills 随安装版本同步。

## 1.9.0-beta.1

- 已确认的 macOS/Windows 本地客户端使用系统安全存储跨任务复用 Device 授权，云端和未知环境保留任务隔离。
- 授权、刷新和退出共用真实系统用户范围的锁；已有凭证时 auth init 返回 authorized 并按需刷新。
- 支持新配置目录恢复共享身份，保留旧授权码登录配置；升级不自动迁移旧任务凭证。
- 同步更新内置 Skills，新增跨任务、并发刷新、共享退出和环境识别测试。

## 0.1.0

- 初始化 `contract-cli` 构建、发布与分发脚手架
