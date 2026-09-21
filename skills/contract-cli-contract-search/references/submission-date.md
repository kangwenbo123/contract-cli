# 合同申请日期

本页适用于 user 身份的系统“合同申请日期”，即提交时间。不是自定义日期、创建时间、归档时间或签订时间。

## 用户怎么说

“查合同申请日期在 2026 年 9 月 15 日到 16 日的合同，和页面一样。”

## 应该怎么查

使用 `combine_condition.submited_time_start` 与 `combine_condition.submited_time_end`。保留接口的 `submited` 拼写，格式为 `YYYY-MM-DD HH:mm:ss`。已确认业务时区的日期范围，按开始日 00:00:00 至结束日 23:59:59 构造；相对日期先换算并说明具体范围，缺少必要年份或业务时区且无法从上下文确定时再澄清。

2026-09-20 生产字段发现确认以下映射。同配置同身份已取得该元数据时直接复用；需重新核实时使用 `search-fields --keyword "合同申请日期"`。

| 元数据 | 实际含义 |
| --- | --- |
| `field_type=BASIC_FIELD` | 系统字段，不需要租户自定义 key |
| `search_field=CONTRACT_SUBMITED_TIME` | 保留返回值，但本字段不使用搜索单元写法 |
| `filter_unique_key=submitedTime`、`filter_unique_key_required=false` | 不把此 key 填入自定义日期筛选 |
| `request_location=COMBINE_CONDITION` | 写入 combine 对象 |
| `request_paths` | `combine_condition.submited_time_start`、`combine_condition.submited_time_end` |
| `search_value_type=DATETIME_STRING_RANGE` | 两个独立字符串属性，不传毫秒数组 |

页面抓包使用 `CONTRACT_SUBMITTED_TIME`、`submittedTime` 和毫秒范围，与上述 MCP 拼写和位置不同。将页面这个筛选单元直接放入 CLI `filter_units`，已实测返回 `code=110000/success=false`，不是“没有合同”。按 MCP 元数据转换，不能自行修正英文拼写后继续套用 filter。

## 示例 JSON

以下是截图对应的请求：当前用户全部可见范围，业务时区 UTC+08:00。日期来自用户截图，不是默认查询时间。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "sort_type": "SUBMITTED_TIME",
  "order": "DESC",
  "combine_condition": {
    "submited_time_start": "2026-09-15 00:00:00",
    "submited_time_end": "2026-09-16 23:59:59"
  }
}
```

保存为请求文件后，复用原配置与身份执行 `contract-cli contract search --profile contract --as user --input-file search-user.json --output json`。保留其他已确认条件；只按日期查询时，不额外添加状态、名称或申请人。用户同时要求已归档时，在同一 `combine_condition` 中加入 `"contract_status_in":"9"`。

## 已验证范围与时间边界

2026-09-20 使用已授权生产身份做了以下只读对照：

- 页面毫秒区间 `[1789401600000,1789574399999]` 对应 UTC+08:00 的 9 月 15 日 00:00:00.000 至 9 月 16 日 23:59:59.999。按上述秒格式查询并取完 3 页，得到 110 个唯一合同 ID、103 个有效唯一组 ID；所有返回的申请时间落在页面范围内，7 个截图可见名称均匹配。
- 在同一日期请求中加入已归档状态，返回 38 条且已查完；合同 ID 集合恰好等于完整日期结果中状态为 9 的子集。验证的是这组日期与状态路径，不推广到所有日期字段。
- 取一个返回时间含非零毫秒的已知合同，名称与其所在秒的相同起止时间组合可命中；移到下一秒、同日午夜、前日或次日的对照点均无命中，且均已查完。日期约束有效，不能描述成只按天比较。
- 小数字符串点查也曾命中，但这不证明小数部分参与了比较。遵守元数据的秒格式，不将 `.999` 或数值时间戳推荐为扩展写法。

这里确认了页面计数和可见样本对应，未取得页面完整合同 ID 集合，不能声称逐条全量完全一致。未取得当天最后一秒内的专门边界样本，也未核对实际 ES 时间原值；不能宣称末尾 999 毫秒一定遗漏或一定完整覆盖。返回 `submitted_time` 含毫秒不等于筛选具备毫秒精度；用户要求精确毫秒范围时，需核实服务端能力，不能偷偷扩展区间。

本地源码中搜索使用 ES 时间筛选，返回值重新读取数据库时间；增量索引 DTO 使用秒格式，另有全量生成路径。这是解释时间精度差异的源码线索，不是生产索引内容的直接证明。此项不替代自定义日期按天/跨日边界验证。

## 结果计数与排序

- 本次 `page_size=50` 的 3 页实际返回 51、56、3 条，分别对应 50、50、3 个合同组。页面的 103 个结果与完整组数相同，不能误报成 CLI 多查了 7 个结果或只查了 103 条合同。
- 遵守 `has_more` 和服务端 token 分页，不能按本页条目数与 `page_size` 比较决定是否结束，也不截断超出的条目。条件变化后清除旧 token。
- 未返回统计时，只有完整取回后才能分别按 `contract_id` 和有效 `group_id` 去重计数；缺少组 ID 时不能把空值视为一组，也不能断言页面总数。报告“合同条目”和“合同组”两个口径；尚有后续页时明确只统计已取回部分。
- 本次请求提交时间倒序，但展平后的合同条目并非全局倒序。不要称第一页是条目级精确的“最新 N 条”；需条目全局排序时先完整取回再本地排序，并说明处理方式。
