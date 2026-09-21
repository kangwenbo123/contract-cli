# 合同搜索逐项验证与调整

日期：2026-09-20。对应 [优化评估](contract-search-agent-review.md)。复用用户已授权生产 profile，只执行合同查询与定向字段发现；真实响应保留在本机临时目录，仓库不写真实人员、合同编号或租户字段 key。

## 1. user 搜索业务错误信号：已修复、生产复测通过

- 生产验证：`search_tab_code=0`、`page_size=1`、无效排序值，返回 `code=110000/success=false`；正式 CLI 1.8.6 退出码为 0。
- TDD：真实错误形状及缺失 code、success=false、正常空结果、缺省 success 的回归先失败；修复 user 搜索响应校验后通过。app 保留原响应契约，raw/JSON 均保留错误响应。
- 修复后用同一个生产请求复测，仍为 `code=110000/success=false`，退出码变为 1。
- 正常对照：正文关键词“里斯”，`page_size=50`，退出码 0、code=0/success=true、18 条、has_more=false。此完整集合用于后续组合对照，不推断其他身份或时间的总量。
- 证据目录：本机 `contract-search-optimization-51_xz0mc`；`invalid-sort.before.response.json` 与 `invalid-sort.response.json`，以及 `body-complete.response.json`。原始数据不随仓库提交。

## 2. 申请人 ID 与正文 AND：已补可靠路径、生产通过

- 从完整 18 条正文结果中取得一个非空 submitter_user_id；不创建人员、不猜 ID。
- `filter_units.CONTRACT_SUBMIT_ID:[真实ID]/MUST` 与正文 SHOULD 组合，返回 1 条、has_more=false，与完整正文结果中该申请人的 ID 集合完全一致。
- 保留人员 ID，正文换成无匹配哨兵词，返回 0 条、has_more=false。
- Skill 新增组合参考与申请人场景入口；缺少可靠 ID 时不退回姓名 MUST。部门等价关系与同名人员仍不能由本测试推出。
- 本机证据：`applicant-id-and-body.*`、`applicant-id-and-no-body.*`。

## 3. 合同名称与正文 AND：已补可靠路径、生产通过

- 使用正文基线中真实合同名称，放 `combine_condition.contract_name`；名称单查 1 条，正文 SHOULD + combine 名称返回相同 1 条。
- 三个集合均完整，组合结果等于名称与正文集合交集；无匹配名称 + 相同正文返回 0 条。
- Skill 补充三段式配方，禁止将两条件机械放入 SHOULD；名称仍是模糊匹配，不宣称精确合同 ID。
- 本机证据：`name-standalone.*`、`name-and-body.*`、`no-name-and-body.*`。

## 4. 多正文单元 AND：生产确认现配方不可靠，已收紧 Skill

- 相同正文词一项 SHOULD、一项 MUST：业务成功却返回 0 条；基线为完整 18 条，集合不等价。
- 第二项换成无匹配词后同样为 0 条。该负例不能推翻相同词正例已经失败的事实，不能据此认定 AND 有效。
- 文档明确没有可执行的多词 AND 配方；不把两个 SHOULD、拼接短语或全部 MUST 当作替代。服务端能力尚待解决，未改服务端。
- 本机证据：`body-two-identical-and.*`、`body-and-no-body.*`。

## 5. 批量编号吞掉关键词：已复现，Skill 改用顶层编号路径

- 两个真实编号放 `condition_units.CONTRACT_NUMBER_FIELD`，单查 2 条；加入无匹配正文 MUST 仍返回同一完整 ID 集合。说明该路径未保留正文限制。
- 相同分号编号改放顶层 `contract_number`，无匹配正文 SHOULD 返回 0 条；正文“里斯”返回 2 条，等于编号与正文完整集合交集。
- 新增 user 专属组合配方，禁止把条件数组批量编号叠加正文作为 AND；不推广到 app V2 或全部组合。
- 本机证据：`batch-condition-alone.*`、`batch-condition-and-no-body.*`、`batch-top-and-no-body.*`、`batch-top-and-body.*`。

## 6. 固定金额、币种与正文组合：闭区间及单边范围已验证

- 以完整正文集合中的真实正金额 a、币种 c 为基准，原正文条件始终保留。
- `[a,a]` 返回 5 条并包含目标；`[a+1,a+1]` 返回 0 条并排除目标；`[a,null]` 返回 16 条，`[null,a]` 返回 6 条。
- 四组均 has_more=false，实际 ID 集合与基线按金额/币种筛出的结果完全一致。已同步固定金额的闭区间、单边 null、严格大于不能等同至少的说明。
- 本机证据：`amount-{exact,excluded,lower,upper}.*`、`amount-comparison-summary.json`。不将此结果推广为所有自定义数字的生产验证。

## 7. JSON 数值与 ID 类型：已修复，区分本地和生产证据

- 本地 HTTP transport 截获先红后绿：9007199254740993、高精度小数、科学计数、嵌套数组和超长整数均保留原 JSON 数值；继续拒绝多文档、尾部垃圾和非对象输入。覆盖共享解析的创建、模板实例化、审批列表兼容，无生产写入。
- 生产正常金额请求在数值修复前后均返回 5 条且完整 ID 集合相同。未取得生产超大数命中样本，因此超大数修复证据来自本地请求链路，不能标为服务端大数接受性验证。
- user 搜索显式不支持的 ID 类型现在本地报错；mock 验证 0 次搜索 HTTP，app 仍透传，按最终身份处理。真实已授权 profile 下 union_id 返回退出码 1 和明确提示，user_id 正常返回 5 条。
- 本机证据：`id-type-rejected.*`、`id-type-valid.*`、`amount-exact-after-number-fix.*`；新增 Go 回归及 search 专属帮助已同步。

## 8. 字段类型与转换：元数据已核实，缺少命中样本的部分保持待验证

- 按币种、金额、日期、部门进行定向字段发现，未无条件枚举全量字段。
- 生产确认自定义币种为 `CONTRACT_FORM_FIELDS_OPTION + CURRENCY_ID_ARRAY`、整数候选 value；基础币种为 `CURRENCY_CODE_ARRAY` 字符串。已修正 OPTION 固定字符串类型的错误，并补语义类型表。
- 选定一个真实自定义币种字段，整数 value 与展示 label 各查询一次，均为业务成功 0 条；这不是阳性对照，不能声称两者等价或字段命中已证实。
- 复用真实 TEXT 空候选字段的已成功查询链路，补完整脱敏元数据→请求示例、真实值来源、条件合并、复用规则；结果为 0 条，未宣称命中。
- 部门角色只区分生产元数据与本地索引路径，不将名称字段与 DEPARTMENT_AUTHORITY 当作同一角色的两种查询。自定义日期的毫秒输入已确认，按天截断仅有源码依据。
- 已读取一份已知合同详情，其 form 为空，无法提供上述自定义字段阳性样本；已向用户询问带部门角色、自定义日期的已知合同。

## 9. 名称与状态：正反对照通过

- 同一真实合同名称加实际状态返回目标 1 条；换成另一个有效状态返回 0 条，均已查完。
- 本轮验证的是 combine_condition 名称与状态路径。生产 submitted_time 为 13 位数字字符串，未据客户端时区猜测输入日界，也未将其与自定义日期粒度混为一谈。
- 本机证据：`name-status-matches.*`、`name-status-other.*`。

## 10. 合同申请日期：生产范围与状态组合通过，精度边界单独记录

- 用户页面截图：合同申请日期 2026-09-15 至 2026-09-16，103 个结果。页面毫秒范围 `[1789401600000,1789574399999]` 对应 UTC+08:00 两日，结束于 23:59:59.999。
- 当前生产元数据为 `BASIC_FIELD / CONTRACT_SUBMITED_TIME / COMBINE_CONDITION / DATETIME_STRING_RANGE`，写 `combine_condition.submited_time_start/end`，保留 `submited` 拼写；不是自定义日期。页面 `CONTRACT_SUBMITTED_TIME/submittedTime` 毫秒 filter 直接复制，实测业务失败 code=110000、退出码 1。
- 按元数据秒字符串查询、page_size=50，完整 3 页分别为 51/56/3 条，50/50/3 组；共 110 个唯一合同、103 个有效唯一组。所有返回申请时间位于页面毫秒范围内，7 个截图可见名称均匹配。未取得页面完整 ID 集合，因此仅确认数量口径和可见样本对应。
- 保持日期范围并增加已归档状态 9，返回 38 条、has_more=false；ID 集合恰好等于上述完整结果中状态为 9 的子集。
- 取一个非零毫秒时间样本，名称加所在秒的点范围命中；下一秒、同日午夜、前日、次日的点范围均无命中，且均已查完。小数字符串点范围也命中，不能据此证明小数参与比较。日期有效但不能从返回毫秒推断筛选毫秒精度。
- 本地源码显示 ES 时间参与筛选、DB 时间参与返回，增量 DTO 使用秒格式，另有全量索引生成路径。未直接读取生产 ES 原值或取得末秒边界样本；不断言根因已证实、不承诺最后 999 毫秒一定覆盖或遗漏，也不将此用于自定义日期验证。
- 请求 SUBMITTED_TIME/DESC 后，展平条目并非全局倒序；Skill 不再把首批称为条目级精确“最新 N 条”。本机证据：`contract-search-submission-date-dbix49v8` 的范围、精度、反例、状态组合及摘要文件。

## 11. 页面“全部”关键词毛鹏：计数与可见样本对应，区别于申请人查询

- 用户截图选中“全部”关键词范围、全部合同页签，filterUnits 为空；conditionUnits 共十二个同词 SHOULD。逐项保留名称、编号、OWNER_NAME、备注、交易方、归档编号、动态表单及五类文本，不沿用上一项日期筛选，也不替换为申请人字段。
- 仅发送 MCP 声明参数，search_tab_code=0、page_size=50。生产返回 30 个唯一合同、27 个有效唯一组，has_more=false、无缺失组 ID；截图中 3 个完整可见名称与 3 个编号均命中。未取得页面完整 ID 集合，不宣称逐条完全一致。
- 同样十二字段换无匹配哨兵词，正常返回 0 条且已查完。单独 CONTRACT_SUBMIT_NAME 搜索相同姓名返回 25 条、24 组；本次是全部关键词集合子集，但不是等价搜索，也不推断任意姓名均有包含关系。
- 本地源码：十二字段均在 condition 白名单；OWNER 与 SUBMIT 分别解析归属人和申请人，不能互换。人员改写为 SHOULD、外层 minimumShouldMatch(1) 与此次 OR 一致；不是任意 MUST 组合的验证。
- MCP DTO 不声明或透传 allConditionUnits/all_condition_units，legacy 未知字段可能被忽略；页面标志影响人员候选获取与表单人员扩展，不是 AND/OR 开关。当前样本对应不能推广为所有页面行为等价；未用传未知字段“成功”证明支持。
- 编号条件含中英文分号时，源码显示会进入批量编号并清除其他关键词条件；“毛鹏”无分号。本限制写入 Skill，不把十二字段配方无条件推广为任意文字查询。
- 本机证据：`/tmp/contract-search-all-keyword-f4_uchub` 的 all、all-no-match、applicant 请求/响应/完整集合及 summary；原始个人及合同信息不入仓库。

## 12. 全部关键词与申请人：姓名 MUST 不可靠，外部 ID 组合通过

- 用户已明确第 3、4 项均以截图的“范学冬（Peter）”为准；确认后继续实际人员对照，直接从同身份已完整取得的合同结果解析到唯一外部 ID，无需用户找 ID。
- 独立负例：保留十二字段毛鹏，加无匹配 CONTRACT_SUBMIT_NAME/MUST，返回完整 30 条/27 组，ID 集合与关键词单查一致；本应为零，说明该姓名 AND 写法不可靠。
- 本地源码确认姓名解析会重写为 SHOULD，无候选时删除分支；filter 白名单无 SUBMIT_NAME，combine 无申请人姓名属性。MCP SUBMIT_ID 按外部 user_id 解析，不是 PC 内部 employeeId。
- 已有完整关键词结果可提供 submitter_user_name/user_id 候选；优先由 Agent 取得并确认，不要求用户找技术 ID。合同搜索不等于完整人员目录，仍须避免取首条、重名、页未查全或混用角色。
- 真姓名 MUST 追加至十二字段毛鹏：首批 51 条、50 组、has_more=true，其中 16 条非目标申请人、34 条不在完整毛鹏基线。已足以证明结果扩大，没有继续翻页，也不报告为总量。
- 目标外部 ID/MUST 与十二字段组合返回完整 5 条、3 组，合同 ID 集合等于完整毛鹏 30 条结果中目标申请人的子集，且所有条目均为目标申请人。与截图的 3 组计数及已匹配的可见样本对应；未取得完整页面 ID 清单，不宣称页面逐条 ID 核对完成。
- 保留真实 ID，十二字段均换成无匹配词，完整返回 0 条。Skill 已补完整可替换 ID 的请求模板，不把姓名单项查询和 AND 组合混为一谈。
- 本机证据：`/tmp/contract-search-applicant-and-vqb5l99h` 的 no-applicant-name-must、confirmed-name-must-and-keyword、confirmed-id-and-keyword、confirmed-id-and-no-keyword 及 confirmed-summary；原始真实 ID 不入仓库。

## 13. 申请人姓名与变更中：姓名直接组合通过，保留组展开语义

- 生产元数据确认合同状态为 BASIC_FIELD / CONTRACT_STATUS / COMBINE_CONDITION / STATUS_CODE_CSV，路径 combine_condition.contract_status_in，变更中 label 对应字符串 value="10"。复制 PC 的 CONTRACT_STATUS/[10]/filter 组合实测 code=110000、success=false、退出码 1。
- 仅一个 CONTRACT_SUBMIT_NAME/SHOULD 的“范学冬”，加 combine 状态 "10"：完整返回 5 条、3 组。改成确认后的外部申请人 ID/MUST，合同 ID 及组 ID 集合完全一致；本场景不必预先解析 ID，仍保留姓名关键词与唯一人员消歧的区别。
- 唯一姓名改成无匹配姓名，保留状态 10，完整返回 0 条。这个分支不会扩大成所有状态合同；不将姓名与其他关键词并列时的失败泛化至此。
- 同姓名＋已知单成员组合同名称：状态 10 返回 1 条且含目标，状态 11 返回 0 条；均已查完，变更中与已变更没有混用。
- 返回 5 条均为目标申请人，其中 3 条状态 10、2 条状态 3；两组各包含审批中/变更中两个成员，另有一组单成员。每组都存在同一条目满足目标申请人和状态 10；结构及截图可见名称/前缀对应，未取得完整页面 ID 清单。
- 源码显示申请人与状态是独立组内 nested 条件，返回又展开组成员；不能保证任意数据下每条都同时满足，不给 5 条统一标变更中。Skill 补组/条目说明及完整取回后按条目核对的边界。
- 本机证据：同目录 confirmed-name-and-changing、confirmed-id-and-changing、no-name-and-changing、known-contract-changing/changed、pc-style-status、status-metadata 及两个 summary。新增 applicant-status.md，并同步主场景/字段/组合/引导与结果呈现。

## 14. 合同需求人张洋：详情取得外部 ID 后命中，姓名不能直接替代

- 用户截图为系统“合同需求人：张洋”，关键词为空，页面 1 个结果。PC 使用 `CONTRACT_DEMAND_PERSON/demandEmployeeId`，但 searchValue 是内部 employeeId，不能直接当作 MCP 的外部 user_id。
- 定向生产元数据确认 `BASIC_FIELD / FILTER_UNITS / USER_ID_ARRAY`，明确要求飞书 user_id，`filter_unique_key_required=false`，`value_scopes=[]`。空候选不代表不支持查询，也不提供人员目录。
- 按截图合同名称搜索得到完整 3 条同名合同，结合签订中状态与交易方唯一定位样本；列表不含需求人字段。读取详情取得唯一 `demand_user_ids` 候选，需求人部门映射内部 ID 与截图所选人员对应；未凭字段名猜测 ID 类型。
- 用详情候选回填需求人 filter，完整返回 1 条、1 组，合同 ID 与已知样本相同，名称、交易方、状态对应截图。直接填姓名或页面内部 ID 均业务成功 0 条且已查完；因此这两种值不能替代正确 ID，零结果不能解释为该人员没有合同。
- 另一个已确认有效的外部用户 ID 加同一已知名称限制返回 0 条；这是受名称限制的负例，不代表另一个人没有需求人合同。没有验证任意多需求人或多字段组合。
- 本地网关转换受配置影响，详情字段值不能无条件保证是外部 ID；历史 `combine_condition.demand_employee_ids` 的 MCP 文案、CLM 处理和严格白名单有冲突，本次使用当前元数据的 filter 路径。Skill 补已知合同取候选、身份核对、回填验证和多需求人消歧规则，不默认让用户找技术 ID。
- 本机证据：`/tmp/contract-demand-person-5pd_mid4` 的 metadata、known-name、target-detail、detail-id、literal-name、pc-internal-id、other-valid-id-known-name 及摘要；真实人员和合同 ID 留在临时证据中，不入仓库。

## 15. 需求人张洋与指定交易方：精确 ID 组合通过，页面 label 路径不等价

- 用户截图选定需求人张洋与交易方湖南安克电子科技有限公司，关键词为空，页面 1 个结果；沿用第 14 项已核验的需求人外部 user_id，不更换角色。
- 当前按“交易方”发现只返回 `CONTRACT_TRADING_PARTY / FILTER_UNITS / TEXT`，未列 `_ID`。本地 MCP 白名单、严格 DTO 及查询实现支持 `CONTRACT_TRADING_PARTY_ID` 字符串数组；生产采用此精确主体路径，不把 TEXT 元数据改成 ID 数组。
- 真实需求人＋真实交易方 `_ID` 两项 MUST 完整返回 1 条、1 组，合同 ID 与需求人单查样本一致；已知详情确认该条需求人、交易方、签订中状态与截图相符。保留需求人、替换另一真实交易方 ID 后完整返回 0 条。
- 名称 TEXT 路径加同一需求人返回同一条，换无匹配名称为零；不能由本次结果相同推断名称与唯一主体 ID 等价。另一有效需求人＋正确交易方 ID＋已知合同名称限制也为零，不推广为该人员与交易方无任何合同。
- 只查交易方 ID 的前三页共 152 条、150 组，仍有下一页；目标样本在已取部分内。该基线未查完，不报告为总量，也不声称与完整交易方集合做过全量交集对照。
- 页面 `CONTRACT_TRADING_PARTY` ID 数组不带 label 为零；附带正确 `search_label` 名称数组命中一条，改成另一真实交易方 ID 而保留 label 仍命中同一条。源码表明该 legacy 分支按名称集合查询、未用 ID 匹配；严格 profile 拒绝 label，不能把补 label 当作精确 ID 修复。
- 现有 `mdm vendor list --as user --name` 完整返回一个同名候选；实际结构为 `data.items[].vendorText/id`、`data.hasMore/pageToken`，ID 与合同 `counter_party_id` 和截图值一致。新增配方保留主体消歧、编码与 ID 区分及组内多条合同分别命中的边界。
- 本机证据：`/tmp/contract-demand-trading-7dz5ur13` 的 metadata、precise-id、name-text、pc-id-array、other-valid-party-id、no-match-name、other-valid-demand-known-name、party-id-alone、两份 label 对照和 vendor-name；原始 ID 仅留本机证据。

## 16. 我方主体候选接口支持 user：列表与详情实测通过

- 用户确认 `mdm legal list` 即我方主体候选入口，并要求验证 user 权限。复用当前生产授权，以已有合同我方主体名称查询，列表返回 1 个候选、`has_more=false`；候选名称及 ID 与已知合同一致。
- 按同一真实主体 ID 调用 `mdm legal get --as user`，业务成功。两次退出码 0、code=0、success=true，证明当前身份可读取该主体列表及详情，不推广为任意账号均有相同权限。
- 列表字段为 `data.items[].id/legal_entity_text/legal_entity`，分页 snake_case；详情在 `data.legalEntity`。不将编码当作 ID，也不照搬交易方列表的 camelCase 分页。
- user 路由分别为 `/open-apis/contract/v1/mcp/legal_entities` 及其 `/{legal_entity_id}`；app 路由也已实现，但本轮没有实测 app。我方列表接口不缺，待明确的是合同搜索的正式精确主体契约，两者不可混为一谈。
- 本机证据：`/tmp/contract-legal-user-sa0s0h1y` 的 list/get 请求结果及 summary；本轮没有测试或发布我方合同搜索的 legacy ID 分支。

## 验证结论进入 Skill 的位置

| 验证项 | Skill 中的对应说明 |
| --- | --- |
| 1. 业务失败与空结果 | 主文件“判断响应与翻页”、user 参数“响应摘要” |
| 2–5. 申请人、名称、正文 AND、批量编号 | `references/search-combinations.md`，主场景 4、6 的入口 |
| 6. 固定金额与币种 | user 参数“范围、日期与部门角色”，组合参考“严格边界” |
| 7. 数值保真与 ID 类型 | user 参数“CLI 参数映射”及其后 JSON 输入说明，区分本地链路与生产证据 |
| 8. 字段类型及空候选 | `references/search-filter-fields.md`；部门、自定义日期待验证说明留在 user 参数 |
| 9. 名称与状态 | 组合参考独立配方；主场景 3、9 统一到已验证 combine 路径，保留“我的合同”页签未做阳性验证的边界 |
| 10. 申请日期、计数、排序 | `references/submission-date.md`；主场景 5、结果说明和引导/字段参考均已链接 |
| 11. 全部关键词搜索 | `references/all-keyword-search.md` 三段式配方；主入口、文本范围、user 字段与引导同步区分全部/文本/申请人 |
| 12. 全部关键词与申请人 | 组合参考的姓名 MUST 反例、ID 来源规则；主场景 6、全部关键词与引导入口同步 |
| 13. 申请人姓名与状态 | `references/applicant-status.md` 两种完整请求和组展开说明；主场景 6、字段/组合/引导同步 |
| 14. 系统合同需求人 | `references/demand-person.md` 三段式配方、ID 来源及正反对照；主入口、申请人场景及字段/引导同步区分角色 |
| 15. 需求人与交易方 | `references/trading-party.md` 两种完整 JSON、候选解析及 label 陷阱；主入口、需求人/字段/组合/引导同步 |
| 16. 我方候选 user 权限 | 搜索总纲区分候选与精确搜索；`contract-cli-mdm-legal/references/entity-query-parameters.md` 记录实测字段和身份边界 |

## 当前仍需样本或服务端能力确认

- 多词正文 AND：正例等价测试失败，当前无可靠配方；已记录限制，未声称服务端已修复。
- 部门名称、申请部门、需求部门的结果角色对照：待业务角色明确且部门不同的真实合同。
- 自定义日期日内/跨日及业务时区、自定义币种的阳性命中：待已填写字段的真实合同。
- 页面全部关键词与 MCP 在人员候选、表单人员扩展上的完全等价性：allConditionUnits 尚未在 MCP 暴露；当前“毛鹏”样本计数及可见样本对应，不将此问题标为已解决。
- 交易方精确 ID 搜索已有白名单和生产对照，但本次字段发现仅返回名称 TEXT，未暴露 `_ID` 路径；Skill 已补已验证配方，服务端元数据补齐仍未完成。

主 Skill 已前置接口选择和执行顺序，保留 9 个三段式场景；明确只问必要业务歧义、修改条件/范围/排序重置分页、业务失败与字段零匹配/合同零命中分开解释。未记录完成的项目继续保留上述边界。

## 交付与回归

- 前述验证使用开发二进制 `bin/contract-cli-search-optimization`，版本 `1.9.0-search.local`。2026-09-21 按用户指定从当前源码构建并本机安装 `1.9.1-beta.1`，PATH 的 `~/.local/bin/contract-cli` 已更新；未发布 npm/GitHub。
- 搜索 Skill 含申请日期、全部关键词、申请人状态、需求人和交易方参考后共 15 个文件。总纲重整轮先同步搜索/shared/mdm-legal三项共25文件；2026-09-21安装轮已将全部14个Skill/106文件统一到 `1.9.1-beta.1`，Codex/agents两处均与安装二进制内嵌内容逐文件一致，原授权配置未变。
- 新增测试均先红后绿；CLI 包 race 最终通过，其他包 race 通过（除下项独立快照检查）；Skill 示例、引用/安装、格式及 6 项独立 Agent 离线行为验证通过。
- `golangci-lint run --enable=exhaustive ./...`：0 issues；`git diff --check` 通过。
- 完整 race 首轮确认本地 ignored `mcp.yaml` 缺少 `list-contract-search-filter-fields`，`TestContractMCPToolSpecsStayAlignedWithMCPYAML` 失败。该旧配置保持原样，后续检查跳过此一项；不将其报告为已通过。
- 本机备份：`~/.codex/backups/contract-search-optimization-20260920-182419`。
- 申请日期补充轮：场景 3 与续页切换到实测 combine 配方，回归先红后绿；搜索/引用/安装回归、25 段 JSON 与 Skill 格式校验通过，lint 0 issues。日期文案经独立 Agent 对照本机响应复核通过。
- 申请日期补充轮的搜索 Skill 共 11 个文件，与当时重新构建的内嵌内容逐一一致，Codex/agents 两处均校验通过；备份为 `~/.codex/backups/contract-search-date-20260920-184013`。
- 全部关键词补充轮：搜索场景/引用/安装回归、26 段 JSON 及 Skill 格式通过；新示例与生产成功请求完全一致，独立 Agent 检查三类人名查询路由及证据边界通过。
- 全部关键词参考加入后最新为 12 个文件，已重建内嵌并逐一校验同步 Codex/agents；备份 `~/.codex/backups/contract-search-keyword-20260920-184534`。本轮未修改 CLI 运行代码。
- 申请人组合负例补充轮：搜索场景/引用/安装回归、26 段 JSON 和独立 Agent 事实边界/路由检查通过；当时 12 个文件重建内嵌后同步两处并校验，备份 `~/.codex/backups/contract-search-applicant-20260920-185253`。该轮之后已获姓名确认并完成本页第 12、13 项正例。
- 第 3、4 项确认补充轮：搜索场景/引用/安装回归、29 段 JSON 及 Skill 格式通过；3 份新增请求与生产成功请求脱敏后完全一致，独立复核通过。13 个文件已重建内嵌并同步校验两处，备份 `~/.codex/backups/contract-search-status-20260920-190205`；本轮未修改 CLI 运行代码。
- 第 5 项需求人补充轮：搜索场景/引用/安装回归、30 段 JSON、Skill 格式及独立事实边界复核通过；新增 JSON 与真实请求脱敏后完全一致。14 个文件已重建内嵌并同步校验两处，备份 `~/.codex/backups/contract-search-demand-20260920-192642`；本轮未修改 CLI 运行代码。
- 第 6 项交易方补充轮：搜索场景/引用/安装回归、32 段 JSON、Skill 格式及独立事实边界复核通过；两份新增模板与真实成功请求脱敏后一致。15 个文件已重建内嵌并同步校验两处，保留并行新增人员/部门路由；备份 `~/.codex/backups/contract-search-trading-20260920-195233`，本轮未修改 CLI 运行代码。
- 总纲及我方 user 补充轮：按身份→角色/字段→真实值与组合→响应顺序整理主入口，保留九个三段式 JSON；11 项离线路由复核、32 段 JSON、Skill/引用/场景与安全文案回归通过。搜索 Skill 通用格式通过；shared/legal 原有 version 元数据不被通用校验器接受，在临时副本去除此 CLI 专用字段后其余格式通过，仓库版本约定保留。三项 Skill 共 25 文件重建内嵌并同步两处，备份 `~/.codex/backups/contract-search-guideline-20260920-202253`；生产补充仅为本页第 16 项列表/详情，不将我方精确合同搜索标记为已验证。
- 本机安装轮：`dist/local-install-1.9.1-beta.1-20260921/qfeius-contract-cli-1.9.1-beta.1.tgz` 为当前 macOS arm64 离线包；临时前缀安装及6项命令检查通过后更新本机。备份 `~/.codex/backups/contract-cli-1.9.1-beta.1-20260921-095609`；本轮只验证安装和命令入口，未新增生产业务请求。
