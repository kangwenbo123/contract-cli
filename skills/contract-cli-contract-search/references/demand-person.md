# 合同需求人

本页用于 user 搜索的系统“合同需求人”，不等于合同申请人、归属人、创建人，也不等于同名自定义人员字段。

## 1. 用户怎么说

“合同需求人是张洋的合同。”用户截图已选择系统“合同需求人：张洋”，显示 1 个结果。

## 2. 应该怎么查

使用 `filter_units.CONTRACT_DEMAND_PERSON`，`search_value` 为已确认的外部 `user_id` 字符串数组，`union_type=MUST`。当前没有需求人姓名专用关键词参数；不使用 `CONTRACT_SUBMIT_NAME`，不把姓名或页面内部 employeeId 填进该数组。

尚无同配置、同身份可复用的字段元数据时，定向发现：

```bash
contract-cli contract search-fields --keyword "合同需求人" --profile contract --as user --output json
```

本次生产元数据为：

| 字段 | 返回值及用法 |
| --- | --- |
| `field_type` | `BASIC_FIELD`，系统需求人 |
| `search_field` | `CONTRACT_DEMAND_PERSON` |
| `request_location / request_paths` | `FILTER_UNITS / ["filter_units"]` |
| `search_value_type` | `USER_ID_ARRAY`；描述明确要求飞书 user_id，不接受 CLM 内部 employeeId |
| `filter_unique_key` | `demandEmployeeId`，`filter_unique_key_required=false`，不是人员 ID |
| `value_scopes` | 空数组，不代表字段不能查询，也不是人员列表 |

### 如何取得真实人员 ID

1. 优先复用同租户、同身份、已经可靠对应到目标人员的外部 user_id，不默认让用户手工找技术 ID。
2. 尚无可靠 ID 时，按 [人员 Skill](../../contract-cli-employee/SKILL.md) 执行 `contract-cli employee list --name "用户提供的姓名" --profile contract --as user --output json`，取得候选 `user_id`，结合状态、部门及分页消歧。确认目标后直接用于需求人筛选，无需先找到合同。
3. 有已知合同或截图样本需要交叉核验时，用其名称/编号定位合同。出现同名合同时结合用户已给出的交易方、状态等信息消歧，不直接取第一条。这里的名称检索用于定位样本，不能把它作为需求人查询的最终结果。
4. 通过 [合同 Skill](../../contract-cli-contract/SKILL.md) 的 `contract get <contract-id> --profile contract --as user --output json` 读取详情，查看 `data.contract.demand_user_ids`。本次列表未返回需求人字段，不能承诺从列表取得；详情 `form` 也不保证包含预置需求人。
5. 保留返回 ID 的原值及字符串类型，核对其人员对应关系，再将候选回填 `CONTRACT_DEMAND_PERSON`，验证能否命中这份已知合同。本次详情只有一个需求人候选，需求人部门映射中的内部 ID 也与用户截图所选人员一致；完成回查后才确认该候选可用。

仅有 `demand_user_ids` 字段名、数字外观或 `saas_demand_employee_ids` 不足以证明 ID 类型：网关转换受配置影响，转换失败可能保留原值。本次实际详情中两个数组恰好相同，均与截图内部 ID 不同；这不能推广为所有环境和合同的固定行为。

多个候选且缺少姓名到 ID 的可靠对应时，不取首个，不把全部 ID 都传入扩大范围；按人员 Skill 补齐分页或仅澄清部门/身份等业务信息。人员查询失败时报告失败，不改用申请人条件，也不自动要求用户提供技术 ID。上面的详情回查是已知样本的交叉核验方式，不是按姓名开始查询的前置要求。

## 3. 示例 JSON

以下是完整模板，须先把占位值替换成经过确认和验证的真实外部 user_id，不能原样执行。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "filter_units": [
    {"search_field": "CONTRACT_DEMAND_PERSON", "search_value": ["REPLACE_WITH_CONFIRMED_USER_ID"], "union_type": "MUST"}
  ]
}
```

保存为 `search-demand-person.json`，复用原配置后执行：

```bash
contract-cli contract search --profile contract --as user --input-file search-demand-person.json --output json
```

## 实测结果与限制

2026-09-20 生产只读验证：

- 按截图名称找到 3 份同名合同，结合截图中的签订中状态和交易方唯一定位样本；详情返回一个需求人 user_id 候选。用该候选作为需求人筛选，完整返回 **1 组、1 条**，合同 ID 正是该已知样本，名称、状态、交易方与截图对应。
- 将同一位置的值直接写成“张洋”或截图内部 employeeId，均返回业务成功、0 条、has_more=false。由于正确 ID 已命中，不能把这两次零结果解释为“张洋没有合同”；先核对 ID 类型与来源，不改用申请人条件。
- 换成另一个已确认有效的外部用户 ID，并保留已知样本名称限制，完整返回 0 条。不能把这个受名称限制的反例说成“另一个人没有需求人合同”。

`combine_condition.demand_employee_ids` 存在历史文案与实现冲突：旧 MCP 描述称飞书 user_id，本地 CLM 代码却将单个值直接交给内部需求人条件，网关转换受运行时配置影响；严格 MCP 的 combine 白名单也未包含此项。本次没有把该历史路径作为推荐或实测通过的配方，统一使用当前字段元数据的 filter 路径。

本次用已知单需求人合同取得并验证 ID，不等于搜索接口允许直接传姓名，也不保证所有详情都返回可直接复用的外部 ID；按姓名取得候选使用前述独立人员查询。后续“需求人＋指定交易方”的已验证组合见 [交易方配方](trading-party.md)，其他多字段及多需求人边界仍需核实；响应按合同组和展开条目分别解释。
