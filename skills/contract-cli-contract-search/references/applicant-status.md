# 申请人姓名与合同状态

适用 user 搜索。指定某个申请人时，按 [主总纲](../SKILL.md) 优先解析人员候选并使用本页的 ID＋状态模板。用户明确要求姓名关键词时，可以使用唯一姓名关键词＋独立状态路径；姓名与其他关键词并列时请读 [另一种组合](search-combinations.md#申请人姓名与全部关键词直接并列不能保证-and)。

## 1. 用户怎么说

“按申请人姓名关键词‘范学冬’，查变更中的合同。”指定“范学冬（Peter）这个人”时改用本页 ID 模板。

本次用户已确认页面所选人员为“范学冬（Peter）”。实际查询时使用用户确认的姓名，不擅自替换相似字；用户已确认的名字不重复询问。

## 2. 应该怎么查

姓名放入唯一的 `condition_units.CONTRACT_SUBMIT_NAME`，使用字符串和 SHOULD；变更中放入 `combine_condition.contract_status_in`，值为字符串 `"10"`。不要额外添加毛鹏或其他关键词。姓名是唯一关键词分支时，它仍然必须命中，状态另行限制；不能由 SHOULD 字面含义误判为姓名可忽略。

明确要求姓名关键词的查询无需先找用户 ID。若用户指定具体人员，则按 [候选来源规则](search-combinations.md#申请人和正文同时满足) 取得并确认外部 `user_id`，优先使用下方 ID 模板，不将“名称可以查”当作精确人员筛选的默认方案；已有可靠 ID 可直接复用，业务身份不重复询问。

2026-09-20 生产元数据确认：系统“合同状态”为 `CONTRACT_STATUS / COMBINE_CONDITION / STATUS_CODE_CSV`，路径 `combine_condition.contract_status_in`，候选 `label=变更中 / value="10"`。`10` 是变更中，`11` 是已变更。不要把页面的 `CONTRACT_STATUS + [10]` filter 直接复制到 MCP；该写法已实测返回 `code=110000/success=false`，不是没有合同。

## 3. 示例 JSON

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "combine_condition": {"contract_status_in": "10"},
  "condition_units": [
    {"search_field": "CONTRACT_SUBMIT_NAME", "search_value": "范学冬", "union_type": "SHOULD"}
  ]
}
```

保存为 `search-applicant-status.json`，保留原授权配置后执行：

```bash
contract-cli contract search --profile contract --as user --input-file search-applicant-status.json --output json
```

已有确认的外部用户 ID、需要按该具体人员筛选时，改用以下完整模板；必须先替换占位值，不能把姓名或 PC 内部 employeeId 写入其中。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "combine_condition": {"contract_status_in": "10"},
  "filter_units": [
    {"search_field": "CONTRACT_SUBMIT_ID", "search_value": ["REPLACE_WITH_CONFIRMED_USER_ID"], "union_type": "MUST"}
  ]
}
```

## 实测对照

- 姓名＋变更中返回 **3 组、5 条**，全部查完。使用已确认外部用户 ID＋同一状态，合同 ID 集合及组 ID 集合完全一致，且所有返回成员的申请人均是该目标人员。
- 唯一姓名条件换成无匹配姓名，保留状态 10，返回 0 条、has_more=false；不会扩大成所有变更中合同。这个结论仅适用于此结构，不能用于姓名与其他关键词并列。
- 在同一姓名条件下，加一份已知单成员组合同的名称进行状态对照：状态 10 返回该合同 1 条，改成状态 11 返回 0 条，均已查完。这验证了变更中与已变更没有混用。
- 组数、两组各两个成员及一组单成员的结构、截图中的可见名称/前缀均对应。未取得页面完整合同 ID 清单，不声称页面逐条 ID 全量核对完成；姓名与 ID 的集合相等仅是本次样本结果，不保证所有同名人员下都相等。

## 结果按合同组解释

页面显示 3 个结果，CLI 展开为 5 条：两组各有 1 条审批中（3）和 1 条变更中（10），另一组只有 1 条变更中。本次所有条目属于目标申请人，每组也都有目标人员且状态 10 的条目。

因此可报告：“找到 3 个合同组，展开 5 条，其中 3 条变更中、2 条为同组审批中成员。”不能把 5 条都写成变更中，也不能仅因 2 条状态不同就判定筛选失效。页面分组展示时保留成员并标出各自状态；用户只要变更中条目时，可在完整取回后按每条实际状态核对并展示那 3 条，说明这是对组展开结果的进一步筛选。

源码中姓名/申请人和状态分别约束组中的合同成员，不保证任意数据下都由同一份合同同时满足。本次检查确认每组存在同一条目同时满足；若用户要求同一份合同满足所有条件，需完整取回并逐条核对申请人及状态，不能拿组命中作替代，也不能在首批数据上筛选后宣称全量。

添加其他关键词会改变查询结构，应重新按 [组合规则](search-combinations.md) 核实。不要把本场景的姓名 SHOULD 简单追加到多字段 OR 中。
