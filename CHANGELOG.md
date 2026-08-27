# Changelog

## Unreleased

- 新增逐请求调用环境识别：实际业务 HTTP 请求发送前回溯父进程；macOS 校验 Bundle ID + Team ID，Windows 校验 Package Family Name 或 Authenticode 证书指纹，并透传归一化来源 Header
- 新增 `contract-cli environment inspect` 本地诊断命令；识别结果不写入 profile 或 OAuth Token
- 新增 `internal/build`，支持 `contract-cli version` 与 `--version`
- 新增 `build.sh`、`Makefile`、`.goreleaser.yml`
- 新增 npm/npx 薄包装：`package.json`、`scripts/install.js`、`scripts/run.js`
- 新增 `tests/cli_e2e/smoke.sh` 作为发布前冒烟脚本

## 0.1.0

- 初始化 `contract-cli` 构建、发布与分发脚手架
