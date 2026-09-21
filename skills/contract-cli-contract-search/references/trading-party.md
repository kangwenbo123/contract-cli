# 需求人与交易方组合

本页用于 user 合同搜索。先区分选定某一交易方主体与按交易方名称关键词查询；两者不能因为本次结果相同就视为等价。

## 1. 用户怎么说

“需求人张洋，交易方是湖南安克电子科技有限公司的合同。”用户截图已选定这两个筛选项，显示 1 个结果。

## 2. 应该怎么查

保留需求人的 `filter_units.CONTRACT_DEMAND_PERSON` 外部 user_id 数组，追加 `filter_units.CONTRACT_TRADING_PARTY_ID` 的真实交易方 ID 字符串数组，两项都用 `MUST`。新增条件后清除旧 `page_token`，从第一页开始。

### Agent 如何取得值

1. 需求人按 [需求人配方](demand-person.md) 复用已确认 user_id，或通过人员查询取得候选；不使用申请人 ID 替代需求人角色，也不把姓名或 PC 内部 employeeId 直接放入该条件。
2. 交易方优先复用同配置、同身份已确认的主体 ID。没有可靠值时按 [交易方主数据 Skill](../../contract-cli-mdm-vendor/SKILL.md) 查询候选：

```bash
contract-cli mdm vendor list --name "湖南安克电子科技有限公司" --page-size 50 --profile contract --as user --output json
```

3. 本次生产返回使用 camelCase：`data.items[].vendorText` 是名称、`id` 是主体 ID，分页为 `data.hasMore/pageToken`。按实际响应读取，不因缺少 snake_case 的 `has_more` 就当作查完。名称查询用于找候选；按名称和已有业务信息消歧，必要时继续分页，不默认取第一条。本次完整返回一个同名候选，ID 与已知合同的 `counter_party_list[].counter_party_id`、截图所选值一致。
4. `vendor` / `counter_party_code` 是编码，不是筛选所需的 ID；不要用我方主体、合同关联行 ID 或展示名替代。已知合同的对方主体列表也可提供成对名称和 ID；有多个交易方时选择用户指定主体，不把全部 ID 一起传入扩大范围。

### 页面字段与 MCP 字段不能只按同名照搬

| 查询意图或路径 | 实际含义与用法 |
| --- | --- |
| 已选定唯一交易方 | `CONTRACT_TRADING_PARTY_ID`＋非空字符串 ID 数组；本次源码白名单、类型校验及生产正反对照已确认 |
| 交易方名称关键词 | `CONTRACT_TRADING_PARTY`＋字符串；本次字段发现返回 `FILTER_UNITS / TEXT`，适用于名称检索，不保证唯一主体 |
| 页面 `CONTRACT_TRADING_PARTY`＋ID 数组 | 只迁移枚举和值、未带 label 时，本次返回 0 条，不能当作有效 ID 筛选 |
| legacy 同枚举＋ID 数组＋`search_label` 名称数组 | 按 label 名称集合查询；本次换成另一交易方 ID、保留正确 label 仍命中，说明 ID 没有参与匹配。不能用补 label 的方式验证精确 ID 条件 |

本次 `search-fields --keyword "交易方"` 仅返回基础名称 TEXT 路径，未返回 `_ID` 项；不能把 TEXT 元数据的示例直接改为 ID 数组。精确主体路径是额外核对 MCP 白名单并完成实测的固定字段。`search_label` 虽在当前 legacy 后端支持，但未列入 MCP 工具单元描述，严格 profile 会拒绝；推荐请求不依赖它。

`CONTRACT_TRADING_PARTY_NAME_PRECISE` 的本地查询实现仍是名称包含匹配；不要仅凭枚举中的 PRECISE 宣称精确名称相等或唯一主体。本次没有把该关键词字段作为精确主体配方。

## 3. 示例 JSON

**选定需求人与交易方的完整模板**：两个占位值都要先替换成已确认真实 ID，不能原样执行。需求人外部 user_id 与交易方主体 ID 是不同类型。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "filter_units": [
    {"search_field": "CONTRACT_DEMAND_PERSON", "search_value": ["REPLACE_WITH_CONFIRMED_USER_ID"], "union_type": "MUST"},
    {"search_field": "CONTRACT_TRADING_PARTY_ID", "search_value": ["REPLACE_WITH_CONFIRMED_TRADING_PARTY_ID"], "union_type": "MUST"}
  ]
}
```

保存为 `search-demand-trading.json`，保留已授权配置后执行：

```bash
contract-cli contract search --profile contract --as user --input-file search-demand-trading.json --output json
```

如果用户明确要“需求人是张洋，交易方名称包含某词”，才改用下面的名称路径。本例名称沿用已验证样本；替换为用户实际关键词，不把名称命中称为唯一主体筛选。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "filter_units": [
    {"search_field": "CONTRACT_DEMAND_PERSON", "search_value": ["REPLACE_WITH_CONFIRMED_USER_ID"], "union_type": "MUST"},
    {"search_field": "CONTRACT_TRADING_PARTY", "search_value": "湖南安克电子科技有限公司", "union_type": "MUST"}
  ]
}
```

## 实测结果与边界

2026-09-20 生产只读验证：

- 正确需求人＋交易方 ID：完整 **1 组、1 条**，命中截图的《【测试文件乱码】（V3.0）》，签订中；合同 ID 与需求人单查的已知样本相同。已知详情支持同一份合同同时具有这两个业务值。
- 保留需求人、改成另一真实交易方 ID：完整 0 条。改用真实名称 TEXT 路径也命中同一份合同；换无匹配名称返回 0 条。两种路径各有对照，不能由此推广为名称与主体 ID 等价。
- 保留目标交易方 ID，改用另一有效 user_id 并加已知合同名称限制：完整 0 条；这是带名称限制的反例，不代表另一人员与该交易方没有任何合同。
- 仅查该交易方 ID 的前三页已返回 152 条、150 组且仍有下一页；这只是局部基线，不报告为总量。目标样本在其中，增加需求人限制后完整收敛为 1 条。
- 页面枚举＋ID 数组不带 label 返回 0 条；带正确 label 返回 1 条；错误主体 ID＋正确 label 仍返回同一条，故 legacy label 路径不能证明按 ID 精确过滤。

两个 `MUST` 在合同组层面组合，可能由同组不同合同分别满足；展开后的每条合同不保证都同时符合。若用户要求同一条目同时满足两个条件，须完整获取结果并核对该条合同的需求人与交易方；本次单组单合同已核对，未验证所有多成员组或多需求人的边界。非数字交易方 sourceId 的兼容性不在本次验证范围内。
