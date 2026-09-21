---
name: contract-cli-employee
description: "使用 contract-cli employee list 按姓名关键词或部门查询人员，取得 user_id、所属部门和人员状态。需要把姓名解析为合同申请人、需求人或自定义人员字段的真实 ID 时使用；合同筛选交由合同搜索 Skill。"
---

# 人员查询

CRITICAL — 开始前 MUST 先读取 [../contract-cli-shared/SKILL.md](../contract-cli-shared/SKILL.md)。复用当前已授权的 contract profile、user 身份与配置目录。

## 命令

执行前可用 `contract-cli employee list --help` 确认当前二进制支持此命令；旧版 CLI 1.8.6 不支持，应使用包含该命令的本地开发版或后续发布版。`unknown command` 不是人员无匹配，也不要改用其他接口猜 ID。

```bash
contract-cli employee list --name "张三" --profile contract --as user --output json
```

按部门查人员时，先使用 [部门 Skill](../contract-cli-department/SKILL.md) 取得并确认真实 `department_id`，将下面占位值替换后执行：

```bash
contract-cli employee list --department-id "REPLACE_WITH_DEPARTMENT_ID" --page-size 20 --profile contract --as user --output json
```

| 参数 | 规则 |
| --- | --- |
| `--name` | 人员姓名关键词；不是精确姓名或人员 ID |
| `--department-id` | 单个开放部门 ID，原样使用 department list 返回的 department_id；映射 query department_collection，不传内部数字 ID、名称或多个 ID |
| `--page-size` | 1–200，默认 10 |
| `--page-token` | 原样使用上一页 token，保留其他参数 |

`--name` 与 `--department-id` 必须二选一，不支持无条件人员目录查询；显式空条件和重复参数会被拒绝。仅支持 user，省略 `--as` 也使用 user。没有 `--user-id`、`--user-id-type`、`--data` 或 `--input-file` 参数。

底层只读 `GET /open-apis/contract/v1/mcp/employees`，固定 `user_id_type=user_id`。每次只读取一页，不自动遍历租户目录。

## 候选与分页

- 检查退出码及业务 code/success；业务失败不等于没有人员。成功响应的 `data.items` 是候选，CLI 原样保留字段和状态。
- 使用 `user_id`，不使用不存在的通用 `id`，不转换为内部 employeeId；数字字符串也保持字符串。姓名、部门和状态一起用于确认身份。
- `status=1` 为正常/在职，`0` 为删除/离职类，`2` 为停用，`-1` 为未知；缺失或未知状态不能视为在职。历史人员查询不自动删除非在职候选；评论 @ 人员等后续写入的状态规则由对应业务 Skill 决定。
- 姓名搜索可返回重名或部分匹配。不能直接取第一条，也不能仅凭一页一个候选断言唯一；结合用户已明确的信息消歧，有必要时询问具体部门/身份，不让用户手工找技术 ID。
- `has_more=true` 时按需继续同条件分页，直到得到足以确认的候选；声称候选完整前必须读完。token 缺失或重复时停止并说明异常。修改姓名/部门条件后清除旧 token；`has_more=false` 即停止，即使 token 非空。
- 成功空列表只说明当前身份和查询条件下无候选，不能推断该人没有历史合同。

## 接入合同查询

取得目标 `user_id` 后交给 [合同搜索 Skill](../contract-cli-contract-search/SKILL.md)，保留原有全部条件与同一身份。申请人使用 `CONTRACT_SUBMIT_ID`，需求人使用 `CONTRACT_DEMAND_PERSON`；自定义人员字段先发现真实 search_field/filter_unique_key。按元数据把 ID 填入字符串数组，不把姓名放进 ID 字段，也不因候选查询失败改用其他人员角色。

简单申请人姓名关键词搜索可直接走搜索 Skill；需要精确人员或与其他关键词组合时才解析 ID。此人员命令不创建或修改人员。
