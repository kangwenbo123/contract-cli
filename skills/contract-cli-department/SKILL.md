---
name: contract-cli-department
description: "使用 contract-cli department list 按名称关键词或开放部门 ID 查询部门，读取部门列表、department_id、层级与状态。需要为人员查询或合同部门筛选取得真实部门 ID 时使用；合同筛选交由合同搜索 Skill。"
---

# 部门查询

CRITICAL — 开始前 MUST 先读取 [../contract-cli-shared/SKILL.md](../contract-cli-shared/SKILL.md)。复用当前已授权的 contract profile、user 身份与配置目录。

## 命令

执行前可用 `contract-cli department list --help` 确认当前二进制支持此命令；旧版 CLI 1.8.6 不支持，应使用包含该命令的本地开发版或后续发布版。`unknown command` 不是部门无匹配。

```bash
contract-cli department list --name "研发" --profile contract --as user --output json
```

用户要求查看部门列表时，可不传名称读取一页：

```bash
contract-cli department list --page-size 20 --profile contract --as user --output json
```

已取得部门 ID 时，将下面占位值替换为真实返回值：

```bash
contract-cli department list --department-id "REPLACE_WITH_DEPARTMENT_ID" --profile contract --as user --output json
```

| 参数 | 规则 |
| --- | --- |
| `--name` | 部门名称关键词，不保证唯一部门 |
| `--department-id` | 单个开放部门 ID，映射 query department_collection；与 name 互斥，不传内部数字 ID、名称或多个 ID |
| `--page-size` | 1–200，默认 10 |
| `--page-token` | 原样使用上一页 token，保留其他条件 |

显式空条件和重复参数会被拒绝。仅支持 user，省略 `--as` 也使用 user；无 `--department-id-type`、`--user-id-type`、`--data` 或 `--input-file` 参数。

底层只读 `GET /open-apis/contract/v1/mcp/departments`；使用网关默认的 open_department_id 契约，每次只读取一页。

## 候选与后续查询

- 检查退出码及业务 code/success；只有业务成功且 `data.items=[]` 才是无匹配。CLI 原样保留候选及未知字段，不把失败当作空结果。
- 返回的 `department_id` 是开放部门 ID（实测为 `od-...`），按字符串原样使用；`dept_level_no` 等层级字段不是要传入的部门 ID，也不从其中截取内部数字 ID。
- 结合 `department_name`、层级与状态确认目标，不盲取第一条或静默丢弃失效部门。候选有歧义时只澄清业务身份；名称可能更改，同一 ID 的目录名称与历史合同展示名可能不同。
- `has_more=true` 时按需继续同条件分页；未读完不能声称目录完整或候选唯一。缺失/重复 token 时停止；条件变化后清除 token；`has_more=false` 时停止，即使 token 非空。
- 查某部门人员时，将确认的 `department_id` 原样传给 [人员 Skill](../contract-cli-employee/SKILL.md) 的 `employee list --department-id`。查询失败时报告错误，不通过猜内部 ID 或切换 ID 类型重试。
- 查合同时交给 [合同搜索 Skill](../contract-cli-contract-search/SKILL.md)，先确定申请部门、需求人部门或自定义部门角色，再按元数据写入 ID 字符串数组。申请部门 `DEPARTMENT_AUTHORITY`、需求人部门 `CONTRACT_DEMAND_PERSON_DEPARTMENT` 和名称关键词 `CONTRACT_DEPARTMENT_NAME` 不能默认互换；不承诺部门 ID 筛选排除子部门。

仅按部门名称关键词搜索合同可以直接走搜索 Skill。本命令不创建、修改或删除部门。
