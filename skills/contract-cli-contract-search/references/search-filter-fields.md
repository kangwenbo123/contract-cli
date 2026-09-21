# 搜索字段发现

`contract-cli contract search-fields` 对应 MCP 工具 `list-contract-search-filter-fields`，查询当前用户可用于合同搜索的字段元信息，包括基础字段和自定义字段。此命令只读、仅支持 user OAuth 身份，本次开发版新增；旧 CLI 1.8.6 尚不支持。

## 调用方式

已知字段展示名时，优先按该名称查询：

```bash
contract-cli contract search-fields --keyword "项目区域" --profile contract --as user --output json
```

`--keyword` 是按字段展示名进行字面量模糊匹配的可选字符串，不是正则表达式、字段 key 或合同内容关键词。接口固定返回中文说明，一次返回全部匹配字段。

只有用户明确要求完整字段清单时，才省略 `--keyword`：

```bash
contract-cli contract search-fields --profile contract --as user --output json
```

| CLI 参数 | MCP 参数或作用 |
| --- | --- |
| `--keyword` | query `keyword`，可选；已知字段展示名时优先传入 |
| `--profile` | 选择已授权配置，发现与搜索保持一致 |
| `--as user` | 当前用户 OAuth 身份；省略时也使用 user，不支持 app |
| `--output json` | 保留字段结构，方便继续构造搜索请求 |
| `--raw` | 原样输出服务端响应 |

接口为 `GET /open-apis/contract/v1/mcp/contracts/search/filter_fields`，query 只有可选 `keyword`。不接受 `--data` / `--input-file` 请求体、分页、`--lang` 或 `--user-id-type` 参数。保留当前 `CONTRACT_CLI_CONFIG_DIR`；不能从其他租户或身份复用字段 key、选项或人员候选。

## 将元数据转换为搜索条件

先检查业务响应成功，再从 `data.items[]` 选中用户实际需要的字段。结合 `field_type`、展示名和业务含义区分系统字段与自定义字段；不要默认选择首项，也不要只凭名称或控件类型决定请求位置。

| 返回信息 | 如何使用 |
| --- | --- |
| `request_location`、`request_paths` | 决定写入 `combine_condition`、`condition_units` 或 `filter_units` 及具体路径；按返回路径执行 |
| `search_field` | 搜索单元使用的字段枚举，按原值传入 |
| `filter_unique_key` | 限定具体自定义字段的唯一 key，与 `search_field` 一起使用；不能用展示名代替 |
| `search_value_type` | 这是语义类型枚举；结合下表和值说明确定 JSON 形状，不是 JSON 原生类型名 |
| `value_description`、`usage_hint` | 确认值的含义、ID 类型、日期单位、范围规则和使用限制 |
| `value_scopes` | 用 label 理解业务含义，实际请求取所选项的 `value` 并保留 JSON 类型；有 children 时按真实层级选择，有 enabled 时检查候选状态 |
| `examples[].request_fragment` | 参考请求片段的路径与结构，替换为用户条件和真实候选；示例值不自动等于用户查询值 |

字段可能映射到不同请求区域，不能把所有返回项统一放入 `filter_units`。自定义字段的同名展示名称也不等于固定搜索参数，例如顶层 `contract_number` 与同名自定义“合同编号”是不同字段。

已实测的基础字段例子：“合同申请日期”返回 `BASIC_FIELD + COMBINE_CONDITION + DATETIME_STRING_RANGE`，实际写入 `combine_condition.submited_time_start/end`。即使元数据同时提供 `search_field` 和 `filter_unique_key`，也不能把它改成 filter 单元；完整配方及页面映射见 [申请日期说明](submission-date.md)。

交易方是另一处需区分语义的例子：本次元数据的 `CONTRACT_TRADING_PARTY / TEXT` 是名称路径，不能照搬页面 ID 数组。精确主体使用另行核实白名单且已生产验证的 `CONTRACT_TRADING_PARTY_ID` 字符串数组；当前按“交易方”发现未返回该项，不将它冒充为 TEXT 元数据的用法。候选来源、与需求人的组合及 legacy label 限制见 [交易方配方](trading-party.md)。

保留字符串 ID 的原值；范围、选项及人员/部门值均遵守该字段元数据。多个匹配项或同名字段经上下文与元数据仍无法区分时，按 [引导式澄清](guided-search.md) 用业务含义让用户选择；唯一且含义明确的匹配不追加确认。缺少执行必需的 key、值类型或候选时继续核实元数据或说明缺失信息，不让用户猜技术值，也不丢弃条件后继续查。

将片段合并进已有完整请求：`combine_condition` 保留其他键，条件数组保留原有其他条件；修改同一条件时替换目标项。不要用片段覆盖整份请求，不自动改变原查询的 AND/OR 含义。新增或修改条件后移除旧 `page_token`，从第一页查询；保持条件不变的续页才复用 token。组合边界见 [组合规则](search-combinations.md)。同一配置、身份和已确认字段的后续追问可复用元数据；遇字段失效或契约变化再刷新。

## 语义类型与 JSON 形状

以当前字段返回的 `search_field + search_value_type + value_description + request_paths` 联合确定请求。下表只解释元数据中的类型；不能按展示名、控件类型或 `search_field` 单独推断。`examples` 用于理解结构，最终值取用户实际条件或真实 `value_scopes[].value`；元数据内部冲突时先核实，不任意选一种表示。

| `search_value_type` | JSON 形状 | 取值与位置 |
| --- | --- | --- |
| `TEXT` | 字符串 | 用户提供的文本；无需枚举候选 |
| `STATUS_CODE_CSV` | 字符串 | 将所选状态 value 按英文逗号连接，如 `"3,9"`，写入返回的具体路径 |
| `INTEGER`、`ENUM_CODE` | 整数 | 取实际枚举 value，不按展示文案猜 code |
| `BOOLEAN_INT` | 整数 `0` / `1` | 不传 JSON boolean；各值含义仍以字段说明为准 |
| `ENUM_CODE_ARRAY`、`INTEGER_ARRAY` | 非空整数数组 | 取实际候选整数值 |
| `CURRENCY_CODE_ARRAY` | 非空字符串数组 | 基础币种，取实际国际代码，如 `["CNY"]` |
| `CURRENCY_ID_ARRAY` | 非空整数数组 | 自定义币种，取该字段返回的整数 code；不能传展示 label 或国际代码替代 |
| `OPTION_LABEL_ARRAY` | 非空字符串数组 | 普通选项，取返回的 value；即使 label 与 value 看起来相同，也不绕过真实值来源 |
| `NUMBER_RANGE_ARRAY` | 两位置数字范围数组 | `[start,end]`；单边 `null` 和边界语义见 [user 参数](search-user-parameters.md#范围日期与部门角色) |
| `DATETIME_STRING_RANGE` | 两个独立日期时间字符串 | 按 `request_paths` 分别写入开始/结束属性，不塞成一个 search_value 数组 |
| `MILLIS_RANGE_ARRAY` | 两位置毫秒时间戳数组 | 日期字符串不能代替毫秒数；输入单位不等于字段查询精度 |
| `USER_ID_ARRAY` | 非空字符串数组 | 已确认的外部人员 ID，按该字段说明使用；不使用示例占位值 |
| `DEPARTMENT_OPEN_ID_ARRAY` | 非空字符串数组 | 已确认的部门 OPENID，不转换为内部部门 ID |
| `CATEGORY_NUMBER_ARRAY` | 非空字符串数组 | 通过合同分类命令取得真实分类 number，按返回路径写入，不用数据库 ID |
| `HYPERLINK_OBJECT` | 对象 | 按字段说明传 `title` 和 `url` |

特别注意：`CONTRACT_FORM_FIELDS_OPTION` 同时用于普通选项和自定义币种。前者可返回 `OPTION_LABEL_ARRAY`，后者返回 `CURRENCY_ID_ARRAY`，因此不是一律字符串数组。2026-09-20 生产字段元数据已确认自定义币种的整数 value；当次选定样本用整数与展示 label 查询都返回 0 条，仅能确认元数据类型，不能据此宣称两种值等价或完成命中对照。

## 空候选如何处理

`value_scopes=[]` 不等于字段不可查询：

- 自由文本、金额和日期范围通常由用户提供值，没有枚举候选也可按元数据构造请求。
- 分类若说明要求通过分类接口取得 number，使用已有 `contract category list`；不要照抄示例或用内部 ID 替代。
- 人员、部门字段没有已有可靠值时，分别使用 [人员 Skill](../../contract-cli-employee/SKILL.md) 的 `employee list --name`、[部门 Skill](../../contract-cli-department/SKILL.md) 的 `department list --name` 取得候选并消歧；选项值仍按字段元数据处理，不把姓名或部门名塞入 ID 字段。
- 合同需求人的 `USER_ID_ARRAY + value_scopes=[]` 可先按人员 Skill 查询姓名得到 user_id，再回填需求人筛选。另有已完成的样本验证：定位已知合同、读取详情的 `demand_user_ids` 候选并交叉核验。列表不保证包含该字段，详情里的 ID 也需核对来源，完整规则见 [需求人配方](demand-person.md)。
- `examples` 中的 `ou_xxx`、`od-xxx`、示例选项等不是可执行候选；不能用猜测值补全空候选。

## 从真实文本字段响应构造请求

以下是已有生产链路的脱敏结构示例。真实 key 和测试文本已替换，**不能原样执行**。该链路已确认 `TEXT + value_scopes=[]` 的请求业务成功且返回 0 条；这证明空候选允许文本查询，不代表有符合条件的合同。

**用户怎么说：**“查自定义字段‘合同编号’包含‘示例编号’的合同。”

**应该怎么查：**先按展示名执行 `search-fields --keyword "合同编号"`，确认用户指定的是自定义字段。不要因为同名而改用顶层系统 `contract_number`。例如响应中的目标项为：

```json
{
  "code": 0,
  "success": true,
  "data": {
    "items": [
      {
        "field_type": "CUSTOM_FIELD",
        "field_display_name": "合同编号",
        "search_field": "CONTRACT_FORM_FIELDS_TEXT",
        "filter_unique_key": "REPLACE_WITH_FILTER_UNIQUE_KEY",
        "filter_unique_key_required": true,
        "request_location": "FILTER_UNITS",
        "request_paths": ["filter_units"],
        "search_value_type": "TEXT",
        "value_description": "传普通字符串。",
        "value_scopes": [],
        "examples": [
          {
            "request_fragment": {
              "filter_units": [
                {
                  "search_field": "CONTRACT_FORM_FIELDS_TEXT",
                  "filter_unique_key": "REPLACE_WITH_FILTER_UNIQUE_KEY",
                  "search_value": "示例文本"
                }
              ]
            }
          }
        ]
      }
    ]
  }
}
```

**示例 JSON：**保留真实响应的字段和唯一 key，将示例文本替换为用户关键词；下面仍为待替换 key 的模板。用户另有条件时按前述合并规则保留，不丢弃已有条件。

```json
{
  "search_tab_code": 0,
  "page_size": 20,
  "filter_units": [
    {
      "search_field": "CONTRACT_FORM_FIELDS_TEXT",
      "filter_unique_key": "REPLACE_WITH_FILTER_UNIQUE_KEY",
      "search_value": "示例编号",
      "union_type": "MUST"
    }
  ]
}
```

构造完整请求后，使用同一身份执行：

```bash
contract-cli contract search --profile contract --as user --input-file search-user.json --output json
```

自定义选项字段的“用户怎么说 → 应该怎么查 → 示例 JSON”见 [主 Skill 场景 8](../SKILL.md#场景-8按自定义选项字段搜索)，其他请求参数见 [user 搜索参数](search-user-parameters.md)。

## 结果与能力边界

- 新命令遇业务非零 `code` 或 `success=false` 时返回失败退出码并保留错误信息，不能解释为空字段结果。
- 成功但没有匹配字段，表示当前用户的字段查询未匹配，不能推断没有相关合同。可核对展示名或请用户明确字段，不自动改为全量枚举。
- 本命令查询字段定义，不返回人员或部门列表。固定申请人姓名、部门名称关键词仍可直接通过 `CONTRACT_SUBMIT_NAME`、`CONTRACT_DEPARTMENT_NAME` 搜索，无需先发现字段。
- 自定义人员/部门字段若要求真实 ID，仍须取得并确认候选；字段元数据不能替代人员消歧或部门角色确认。
