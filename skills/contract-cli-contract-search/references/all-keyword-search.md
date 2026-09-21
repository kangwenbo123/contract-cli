# 页面“全部”关键词搜索

本页用于 `contract-cli contract search --profile contract --as user`，复现页面搜索框的默认逻辑。合同页签“全部合同”决定可见范围；搜索指定内容中的“全部”决定关键词查询哪些字段，两者是不同设置。

## 1. 用户怎么说

“关键词搜索毛鹏。”用户提供的截图中，搜索指定内容选中“全部”，全部合同页签，未添加其他筛选。

或明确说：“在页面的全部关键词范围里搜索毛鹏。”

也适用于“搜索毛鹏”“搜索里斯”“查华东”这类只指定普通搜索词、没有指定字段或业务角色的 user 查询，默认使用全部关键词并说明范围。

显式关键词意图优先于实体外观：即使“毛鹏”能匹配人员候选，也不把“关键词搜索毛鹏”改成申请人精确筛选。没有显式关键词意图、但“毛鹏的合同”“法务的合同”等表达可合理推测是人员或部门时，先按 [引导式澄清](guided-search.md) 查询候选并确认业务角色。推测依据不足时回到页面默认关键词，不凭词语外观虚构角色。

## 2. 应该怎么查

使用下列十二字段的同一个关键词，各项都是 `SHOULD`，任一来源命中即可。进入本路径后，关键词看起来是人名也不能自动改成“该人申请的合同”；不额外增加申请日期、状态或姓名 ID 筛选。

页面提示中的合同名称、交易方、申请人、合同文本、附件文本、合同说明、合同编号、归档编号和自定义字段是用户语义分类，并不与请求单元一一对应。此前页面抓包的默认请求实际展开为以下十二项；尤其人员相关项使用 `CONTRACT_OWNER_NAME`，不能据此声称已精确限定申请人。用户明确指定申请人时仍走候选和 `CONTRACT_SUBMIT_ID` 路径。

| 页面语义及实际索引来源 | 实际请求字段 |
| --- | --- |
| 合同名称、合同编号 | `CONTRACT_NAME_FIELD`、`CONTRACT_NUMBER_FIELD` |
| 页面默认人员关键词 | `CONTRACT_OWNER_NAME`；这是抓包中的实际字段，不等于申请人精确筛选 |
| 合同说明或备注、交易方、归档编号 | `CONTRACT_REMARK_FIELD`、`CONTRACT_CAUSE_FIELD`、`CONTRACT_TRADING_PARTY`、`ARCHIVE_NUMBER_FIELD` |
| 动态表单 | `CONTRACT_FORM_FIELD`，不指定自定义字段唯一 key |
| 合同文本和附件文本索引 | `CONTRACT_TEXT_FIELD`、`CONTRACT_SCAN_TEXT`、`CONTRACT_ARCHIVE_ATTACHMENT`、`CONTRACT_ATTACHMENT_TEXT` |

这是页面默认请求的实际范围；不把“全部”扩展为所有字段枚举，也不自行增加申请人或部门。用户明确只查合同文本时使用 [五类文本配方](text-search.md)；明确“毛鹏申请的”时按 [主总纲](../SKILL.md) 先解析目标人员并优先 ID 筛选。“毛鹏的合同”这类可合理推测为人员但角色不完整的表达，按引导参考先查候选再确认；其他普通词直接使用本配方。

无需为这个固定关键词配方先调用 `search-fields`；如果用户限定某一个自定义字段及其值，再按 [字段发现](search-filter-fields.md) 选择具体 key 和请求位置。动态表单关键词命中不证明某个指定字段精确等于该值。

## 3. 示例 JSON

将十二处“毛鹏”一起替换为用户实际关键词，保持用户明确的搜索范围。**关键词含 `;` 或 `；` 时不要直接套用本例**，先按下方能力边界处理。

```json
{
  "search_tab_code": 0,
  "page_size": 50,
  "condition_units": [
    {"search_field": "CONTRACT_NAME_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_NUMBER_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_OWNER_NAME", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_REMARK_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_TRADING_PARTY", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "ARCHIVE_NUMBER_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_FORM_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_TEXT_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_CAUSE_FIELD", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_SCAN_TEXT", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_ARCHIVE_ATTACHMENT", "search_value": "毛鹏", "union_type": "SHOULD"},
    {"search_field": "CONTRACT_ATTACHMENT_TEXT", "search_value": "毛鹏", "union_type": "SHOULD"}
  ]
}
```

保存为 `search-all.json`，复用原授权配置后执行：

```bash
contract-cli contract search --profile contract --as user --input-file search-all.json --output json
```

## 实测与结果解释

2026-09-20，在同一已授权生产身份下：

- 上述十二字段“毛鹏”请求返回 **30 个唯一合同、27 个有效唯一合同组**，`has_more=false`，没有缺失组 ID。与截图的 27 个结果对应；截图中 3 个可见完整名称和 3 个可见编号均命中。未取得页面完整 ID 集合，不能声称逐条全量完全一致。
- 同样十二字段全部换成无匹配哨兵关键词，返回 0 条、`has_more=false`。正常无命中与业务错误仍需区分。
- 单独使用 `CONTRACT_SUBMIT_NAME` 搜索相同姓名，返回 **25 条、24 组**，均已查完；本样本是全部关键词结果的子集。不能把两种查询混用，也不能推断所有姓名下都存在相同包含关系。

按合同组和合同条目分别报告数量，规则见 [结果计数](submission-date.md#结果计数与排序)。没有命中字段或片段时，不因出现人名就断言这份合同命中了申请人、正文或某个表单字段。

## 能力边界

- 页面 `allConditionUnits=true` 不在 MCP DTO 的声明或透传路径中，`all_condition_units` 也不是受支持参数；本例不传它。不要因为未知字段请求成功就认为它生效。源码中该标志影响人员候选获取和表单人员值扩展，并非 AND/OR 开关；本次样本对齐不代表所有页面行为都已对齐。
- 十二字段包含编号。源码中编号条件遇 `;` 或 `；` 会进入批量编号分支，清除其余关键词条件；本次“毛鹏”不受影响。若用户要批量编号，改用 [编号组合配方](search-combinations.md#分号批量编号和正文同时满足)；若明确要把分号作为普通关键词的一部分，当前不能承诺该十二字段配方保留原意，不删除分号或删掉编号字段后声称仍是原范围。
- 本例验证一个词跨来源 OR。多个词同时满足、排除或新增结构化条件时，保留用户完整意图并核实组合契约；不能通过重复十二项、改成全部 MUST 或把多个词拼成短语来宣称已支持 AND。
- 全部关键词再加申请人时，不能直接追加姓名 MUST；无匹配姓名仍返回原来的 30 条、27 组，有匹配姓名也已复现结果扩大。按 [已验证的完整 ID 配方](search-combinations.md#申请人姓名与全部关键词直接并列不能保证-and)，由 Agent 从已有合同结果取得已确认人员的外部 ID；范学冬样本得到 3 组、5 条且反例为零，无需用户手工找 ID。只有用户明确要求姓名关键词时，单个姓名＋独立状态才直接用姓名，见 [另一种组合](applicant-status.md)；指定人员仍优先 ID。
- 查询依赖服务端索引和当前用户权限，不承诺所有历史版本、停用人员或所有自定义控件内容都已覆盖。出现页面差异时，先核对身份、页签、关键词、选中范围、附加筛选和分页，不自动切换身份或扩大范围。
