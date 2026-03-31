# tiandi_hands_rules 测试接入与验证计划

## 0. 当前状态

- [x] Step 1：制定 Plan，明确范围、上下文、任务拆分、测试策略与协作方式
- [x] Step 2：固化自然语言完整规则内容并形成实现输入
- [x] Step 3：后端接入测试模式与规则承载
- [x] Step 4：前端接入规则内容与测试入口
- [x] Step 5：完成模块化测试与集成测试
- [x] Step 6：更新架构文档 `docs/tiandi_arch_and_api.md`

## 1. User 原始提示

> 现在需要对 [tiandi_hands_rules.md](docs/tiandi_hands_rules.md) 在游戏的后端和前端进行测试  
> 方便起见，测试期间，我们总是默认P0是地主，其它则仍然正常随机模式  
> 请你制定一个Plan md在docs/plan/任务名称.md, 包含user原始提示，任务背景（也就是完成plan所需要的完整的上下文），打勾的步骤  
> 首先，我们需要确定自然语言的完整内容  
> 然后是如何将这些内容，插入到后端和前端模块  
> 由于可以让subagents (gpt 5.4 high)来派遣任务，所以任务中需要写明白并行和串行的任务，并且要明确每个任务的目标和输出。  
> 还要制定好test plan, 确保每个模块的功能都能被正确测试，并且接口的协作能够顺利进行。  
> 测试Plan最好也是模块化的，能够针对每个模块进行独立测试，同时也要有集成测试来确保整个系统的协作性。  
> 后端测试鼓励使用带详细注释和mock数据的，可复用性强的测试用例，前端测试则需要结合代码与playwright interacive mcp工具进行测试，最好是结合前端实际回传图片进行测试  
>   
> 制定完成Plan之后，请你作为协调者，分发任务给subagents，并监督任务的执行，确保每个任务都能按时完成，并且输出符合预期。  
> 每个任务完成后，需要打勾。  
> 全部任务完成后，更新架构文档docs/tiandi_arch_and_api.md  
>   
> 【step 1, 制定plan, 无需code】

## 2. 任务背景

### 2.1 当前仓库状态

- 后端当前已实现从洗牌、扣底、选天地赖、发牌、叫地主、抢地主、我抢，到进入 `PLAY` 阶段的状态机。
- 当前核心后端入口：
  - `backend/internal/tiandi/fsm/machine.go`
  - `backend/internal/tiandi/demo/service.go`
  - `backend/cmd/tiandi-server/main.go`
- 当前前端是后端状态驱动页面：
  - `web/src/App.tsx` 负责并行拉取局面和规则目录
  - `web/src/components/RulesPanel.tsx` 负责规则展示
  - `web/src/components/ActionBar.tsx` 负责动作提交
- 当前规则目录来自后端静态结构：
  - `backend/internal/tiandi/rules/catalog.go`
  - `web/src/types.ts`
  - `web/src/api/rules.ts`
- 当前已有测试：
  - 后端规则目录测试：`backend/test/backend/rules_test.go`
  - 后端状态机测试：`backend/test/backend/fsm_test.go`
  - HTTP 接口测试：`backend/cmd/tiandi-server/main_test.go`
  - 前端测试：`test/web/app.test.tsx`

### 2.2 本任务的真实目标

- 不是直接实现完整出牌引擎。
- 而是先把 `docs/tiandi_hands_rules.md` 变成“可被后端和前端稳定消费与测试”的规则内容。
- 同时引入一个测试期专用模式：默认 `P0` 为地主，其余逻辑尽量保持随机或沿用现有行为。
- 最终需要得到：
  - 一份可追踪的自然语言完整规则稿
  - 一组后端可消费的规则承载结构/测试入口
  - 一组前端可展示、可交互验证的规则与局面展示
  - 一套模块测试 + 集成测试 + Playwright 交互验证
  - 一次架构文档回写

### 2.3 关键约束与假设

- 本轮计划阶段不写业务代码，只做执行设计。
- 测试期“默认 P0 是地主”应优先作为“测试模式/测试配置”处理，而不是在前端硬编码。
- 现有后端 `fsm.Machine` 默认 `StartingBidder` 随机，地主归属由叫抢流程决定；如果要稳定测试，需要额外的测试注入点或 deterministic 模式。
- 前端当前不做牌序和规则判断，主要承担展示与接口调用，因此规则一致性应以后端为单一事实源。
- `docs/tiandi_hands_rules.md` 当前是草案，存在“实现前需要补全/澄清”的自然语言边界，需要先收敛再接模块。

## 3. 自然语言完整内容范围

### 3.1 需要收敛成完整文本的内容

- 牌点序、王与普通牌限制
- 基础牌型：单、对、三
- 组合牌型：三带一、三带二、四带二/两对
- 连续牌型：顺子、连对、三连
- 飞机：不带、带单、带对、比较规则
- 炸弹：四炸、5+ 炸、王炸、纯赖子炸、赖子替代炸
- 常规优先级与炸弹优先级
- 同轮比较规则与异型不可比规则
- 赖子替代规则、禁止替代边界、替代明细回传建议
- 与当前项目关系：当前阶段仅先做“规则内容接入与测试”，不等于完整出牌判定引擎

### 3.2 需要明确的歧义点

- [x] “四带二/两对”允许 `AAAA + B + C` 扩展
- [x] 飞机比较时主比点以最高三张为准
- [x] 赖子替代允许多种同优结果，测试与展示阶段应返回所有可能
- [x] “测试期默认 P0 为地主”直接跳到 `PLAY`，但发牌、底牌、赖子仍保持随机
- [x] 前端需要专门展示“当前为测试模式 / 固定地主为 P0”

### 3.3 该阶段产物

- 一份结构化规则确认稿，建议直接沉淀为：
  - 规则文档补充节
  - 或后端规则元数据生成说明
  - 或 `docs/plan` 内的“规则确认清单”

### 3.4 已确定规则决议

- “四带二/两对”拆成两个可测试结构：
  - `AAAA + BB + CC`
  - `AAAA + B + C`
- 飞机主比点统一取最高三张点数。
  - 示例：`JJJ QQQ` 与 `101010 JJJ` 比较时，前者主比点为 `Q`，后者为 `J`，因此前者更大。
- 赖子替代结果不做单解压缩。
  - 对同一手牌，如存在多个等价或同优可行替代，后端测试结果与前端展示都应支持“列出全部可能”。
- 测试模式下：
  - 发牌仍随机
  - 底牌仍随机
  - 天赖/地赖仍随机
  - 状态机直接进入 `PLAY`
  - 地主固定为 `P0`
  - 前端页面显式展示当前为测试模式

### 3.5 新增规则承载要求

- 后端规则模型后续需要支持以下信息，供测试和前端消费：
  - 牌型唯一 key
  - 牌型结构 pattern
  - 最小张数
  - 比较主键定义
  - 是否可与同 key 不同变体互比
  - 是否允许赖子替代
  - 赖子替代结果是否多解
  - 不可比较原因说明
- 对“四带二/两对”建议拆分为两个 rule item，而不是只保留一个文案项。
- 对飞机建议显式增加 `compareBy: highest_triplet_rank` 一类的元数据说明。

## 4. 插入点设计

### 4.1 后端插入点

- 规则内容承载：
  - `backend/internal/tiandi/rules/catalog.go`
  - 如需要，可新增更细的规则源文件，例如 hand 定义、比较说明、测试 fixture
- 测试模式插入：
  - `backend/internal/tiandi/fsm/machine.go`
  - `backend/internal/tiandi/demo/service.go`
  - 目标是支持“固定 `P0` 地主并直接进入 `PLAY`”的测试入口，同时保留随机发牌/底牌/赖子
- 接口暴露层：
  - `backend/cmd/tiandi-server/main.go`
  - 如有必要，可通过 query / env / test-only constructor 暴露测试模式，但默认优先选择不污染正式接口的方案

### 4.2 前端插入点

- 类型与数据消费：
  - `web/src/types.ts`
  - `web/src/api/game.ts`
  - `web/src/api/rules.ts`
- 规则展示：
  - `web/src/components/RulesPanel.tsx`
- 测试态展示与行为验证：
  - `web/src/App.tsx`
  - 增加测试模式提示区域，明确“当前为测试模式 / 固定地主为 P0”
- 前端测试：
  - `test/web/app.test.tsx`
  - 如需要，可增加更细粒度组件测试

### 4.3 推荐实现方向

- 后端优先增加“测试模式构造入口”或“测试模式选项”，避免污染正式主流程。
- 测试模式在后端直接产出 `PLAY` 态快照，前端只消费状态，不在前端推导地主归属。
- 规则比较测试不要混在 UI 层完成。
  - 规则比较正确性以后端测试为主。
  - 前端重点验证规则展示、测试态提示、以及接口协作。

## 5. 执行策略

### 5.1 串行主链

- [x] S1. 规则自然语言收敛
  - 目标：把 `docs/tiandi_hands_rules.md` 转成“实现和测试可直接消费”的完整规则说明
  - 输出：规则确认稿、歧义处理决议、规则字段映射表、比较关系测试清单
- [x] S2. 测试模式方案定稿
  - 目标：确定“固定 `P0` 地主并直接进入 `PLAY`”的注入点、开关边界、前端展示方式
  - 输出：后端测试模式设计说明、接口影响面清单、前端测试态展示方案
- [ ] S3. 后端与前端按统一规则模型接入
  - 目标：确保规则文本、后端目录、前端展示三者一致
  - 输出：代码改动、测试 fixture、接口契约、比较结果承载结构
- [ ] S4. 模块测试与联调
  - 目标：分别验证后端、前端、接口集成、preorder 比较关系与真实页面交互
  - 输出：测试结果、截图证据、异常记录、比较矩阵验证结果
- [ ] S5. 架构文档更新
  - 目标：把实际落地方案回写到 `docs/tiandi_arch_and_api.md`
  - 输出：更新后的架构与 API 协作文档

### 5.2 可并行任务

- [x] P1. 自然语言规则整理 subagent
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：从 `docs/tiandi_hands_rules.md` 提炼完整规则、吸收本轮新增决议、输出字段映射和比较关系清单
  - 输入：规则文档、当前规则目录结构、现有测试
  - 输出：规则确认稿 + 决议清单 + 建议字段模型 + preorder 测试对象列表
- [x] P2. 后端接入设计 subagent
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：给出后端测试模式、规则承载、比较结果多解展示的最小侵入改动方案
  - 输入：`fsm/machine.go`、`demo/service.go`、`rules/catalog.go`
  - 输出：后端改动计划、测试清单、风险点、比较测试夹具方案
- [x] P3. 前端接入设计 subagent
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：给出前端规则展示、测试模式提示、多解替代展示预留、测试夹具与 Playwright 方案
  - 输入：`App.tsx`、`RulesPanel.tsx`、`types.ts`、现有前端测试
  - 输出：前端改动计划、测试夹具设计、Playwright 验证脚本思路
- [x] P4. 测试计划 subagent
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：输出模块化 test plan，包括 mock 数据策略、复用策略、preorder 比较矩阵与集成矩阵
  - 输入：后端/前端/接口现状与脚本
  - 输出：测试矩阵、测试数据设计、验收标准

### 5.3 串并行依赖关系

- `P1` 必须先完成，才能锁定自然语言规则基线。
- `P2`、`P3`、`P4` 可在 `P1` 初稿完成后并行开展。
- 最终需要由协调者汇总 `P1/P2/P3/P4` 输出，形成统一实施方案后才能进入实际编码与执行。

## 6. Subagent 分发计划

### 6.1 协调者职责

- 统一上下文，避免后端、前端、测试对规则理解不一致
- 在每个 subagent 完成后做交叉审阅
- 发现冲突时优先维护“后端规则为单一事实源”
- 维护本计划中的勾选状态与完成证据

### 6.2 任务卡

- [x] T1. 规则确认任务
  - 负责人：Subagent A
  - 目标：给出可实现的自然语言规则完整稿
  - 交付：完整规则清单、已确认决议、建议补充文案、比较关系对象清单
- [x] T2. 后端方案任务
  - 负责人：Subagent B
  - 目标：明确测试期 `P0` 固定地主并直接进入 `PLAY` 如何进入现有状态机/服务层
  - 交付：后端最小改动方案、接口影响说明、后端测试建议、比较夹具结构建议
- [x] T3. 前端方案任务
  - 负责人：Subagent C
  - 目标：明确规则展示、测试态提示、前端消费结构和组件测试策略
  - 交付：前端改动点、组件/页面测试建议、Playwright 验证步骤、截图点位建议
- [x] T4. 测试设计任务
  - 负责人：Subagent D
  - 目标：制定可复用、注释完整、数据明确的测试计划，覆盖 preorder 比较关系
  - 交付：测试矩阵、mock 数据模型、验收步骤、截图证据要求、不可比较断言清单

## 7. Test Plan

### 7.1 后端模块测试

- [ ] 规则目录测试
  - 验证 `Catalog` 与自然语言确认稿一致
  - 验证牌型、优先级、比较说明、赖子说明字段齐全
  - 验证“四带二/两对”包含两个结构变体
  - 验证飞机比较主键声明为“最高三张”
  - 验证赖子替代支持“多解展示”
  - 测试用例要求附带注释，说明每条规则对应文档出处
- [ ] 状态机测试
  - 验证测试模式下直接进入 `PLAY`
  - 验证测试模式下地主固定为 `P0`
  - 验证测试模式下 `P0` 手牌为 20 张
  - 验证测试模式下底牌归入 `P0`
  - 验证非测试模式仍保留原随机/正常流程
  - 验证进入 `PLAY` 后底牌归属、倍数、赖子展示状态不异常
- [ ] 服务层测试
  - 验证 `State/Reset/Apply/Rules` 在测试模式下输出稳定
  - 验证测试模式下 `availableActions` 与 `currentActor` 的约定符合预期
  - 验证错误输入仍按原契约返回错误
- [ ] HTTP 接口测试
  - 验证 `/api/game/state`、`/api/game/reset`、`/api/game/action`、`/api/game/rules`
  - 验证测试模式的入口方式不会破坏现有接口契约

### 7.2 牌型比较测试：preorder 设计

- [ ] 比较测试拆成两层：
  - 可比较性测试
  - 可比较对象上的偏序/全序链测试
- [ ] 可比较性测试目标
  - 同结构且规则允许比较的牌型，应返回“可比较”
  - 异结构且规则不允许比较的牌型，应返回“不可比较”
  - 特殊压制链如炸弹/王炸，应按特殊规则可比较
- [ ] 不可比较断言必须覆盖
  - 单 vs 对
  - 三带一 vs 三带二
  - 顺子 vs 连对
  - 飞机不带 vs 飞机带单
  - `AAAA + BB + CC` vs `AAAA + B + C` 是否互比，需在规则模型中明确
  - 普通牌型 vs 非同层特殊压制牌型
- [ ] 可比较断言必须覆盖
  - 同牌型不同主键
  - 同牌型同主键
  - 炸弹内部优先级
  - 王炸对任意普通炸弹
  - 飞机按最高三张比较
  - 含赖子与不含赖子的同结构比较
- [ ] preorder 测试目标
  - 先验证“不可比较”的边界是正确的
  - 再验证“可比较集合内”的顺序链是正确的
  - 如果设计上存在等价类，则验证：
    - 自反性
    - 可传递性
    - 同优不同解的等价展示
- [ ] 推荐测试组织形式
  - `comparability_test.go`
  - `preorder_test.go`
  - `fixtures/` 下按牌型组织样例
- [ ] 推荐断言结构
  - `comparable: true/false`
  - `relation: less/equal/greater/incomparable`
  - `reason`
  - `matchedType`
  - `mainKey`
  - `allCandidateInterpretations`

### 7.3 前端模块测试

- [ ] 类型与 API 层测试
  - 验证 `RulesCatalog`、`GameState` 结构兼容新增字段
  - 验证规则接口失败时降级仍成立
- [ ] 组件测试
  - `RulesPanel`：规则章节、优先级、说明文本正确渲染
  - `ActionBar`：测试模式下的动作展示符合预期
  - 测试模式提示组件或提示区：明确显示“当前为测试模式 / 固定地主为 P0”
  - 如新增“测试模式”提示组件，应独立测试
- [ ] 页面测试
  - `App`：并行加载局面与规则目录
  - 固定 `P0` 地主后的主页面渲染稳定
  - 页面明确展示测试模式
  - 规则展示与局面展示不会互相阻塞

### 7.4 Playwright 交互测试

- [ ] 启动真实后端与前端联调环境
- [ ] 使用 Playwright interactive MCP 进入页面，验证：
  - 首屏规则内容可见
  - 页面明确显示“当前为测试模式 / 固定地主为 P0”
  - 局面展示与 `P0` 地主测试态符合预期
  - `P0` 的地主标识与底牌归属可见
  - 动作按钮与规则面板共存时页面仍正常
- [ ] 关键页面截图回传
  - 初始页
  - 规则面板可见状态
  - 测试期 `P0` 地主局面
  - 测试模式提示区
- [ ] 截图需作为前端联调证据保留在任务输出中

### 7.5 集成测试

- [ ] 命令级验证
  - `npm run backend:test`
  - `npm test`
  - `npm run build`
- [ ] 端到端验证
  - 后端真实服务 + 前端真实页面 + Playwright 交互
- [ ] 回归验证
  - 现有规则目录展示不能退化
  - 现有叫地主/抢地主基础行为在非测试模式下不能被破坏

## 8. Mock 数据与复用策略

- 后端测试优先使用可复用 fixture：
  - 固定牌组
  - 固定赖子
  - 固定地主
  - 固定倍数
  - 固定比较对
  - 固定不可比较对
  - 固定多解赖子替代样例
- 每个 mock 数据块必须写清楚：
  - 为什么这样构造
  - 对应验证哪个规则或状态
  - 是否可在多个测试中复用
- 前端测试优先复用：
  - `GameState` fixture
  - `RulesCatalog` fixture
  - Playwright 截图场景定义

## 9. 风险与决策点

- 如果“固定 `P0` 为地主”直接写死在主流程，会污染正式逻辑；优先改为测试模式或测试构造入口。
- 如果自然语言规则不先收敛，后端目录、前端展示、测试断言会出现三套口径。
- 如果不先定义“可比较 / 不可比较”的边界，后续顺序测试会失真，因为 preorder 的第一层不是排序，而是定义域。
- 如果不显式支持“赖子替代多解”，测试会错误地把“展示一个结果”当成“只有一个结果”。
- 如果前端只做 mock 测试而不做真实页面交互，规则面板与实际局面联动问题可能遗漏。
- 如果 Playwright 不保留截图证据，前端联调结果不易复核。

## 10. 完成定义

- [x] 自然语言规则确认稿完成，并被协调者确认
- [x] 后端测试模式方案完成，并确认不污染正式逻辑
- [x] 前端接入方案完成，并确认以后端为单一事实源
- [x] preorder 比较测试方案完成，并确认覆盖“可比较/不可比较 + 偏序链”
- [x] 模块化测试全部通过
- [x] Playwright 真实交互验证完成，并产出截图证据
- [x] `docs/tiandi_arch_and_api.md` 更新完成

## 11. 下一步

- 下一轮进入实际实现阶段，按第 14.7 节执行 `E1 -> E6`。
- 实现阶段的第一个目标是后端测试模式与 `testMode` 状态字段，而不是先动前端。

## 12. T1/P1 输出：规则基线说明

### 12.1 A. 完整规则确认稿

#### 12.1.1 基本约定

- 牌点序从小到大为：
  - `3, 4, 5, 6, 7, 8, 9, 10, J, Q, K, A, 2, BlackJoker, RedJoker`
- 比较默认只比较主序列或关键牌，不比较花色。
- 大小王不能作为顺子、连对、三连、飞机的核心点数。
- `天赖`、`地赖` 是非王赖子，允许替代非王点数，不替代大小王。
- 在“测试模式”下：
  - 发牌随机
  - 底牌随机
  - 天赖/地赖随机
  - 后端直接产出 `PLAY` 态
  - 地主固定为 `P0`
  - 前端必须显式展示“当前为测试模式 / 固定地主为 P0”

#### 12.1.2 基础牌型

- `single` / 单张
  - 结构：`A`
  - 说明：任意 1 张牌，包括普通牌或王
- `pair` / 对子
  - 结构：`AA`
  - 说明：2 张同点数牌
- `triple` / 三张
  - 结构：`AAA`
  - 说明：3 张同点数牌

#### 12.1.3 组合牌型

- `triple_with_single` / 三带一
  - 结构：`AAA + B`
  - 主比较键：三张点数
- `triple_with_pair` / 三带二
  - 结构：`AAA + BB`
  - 主比较键：三张点数
- `four_with_two_pairs` / 四带两对
  - 结构：`AAAA + BB + CC`
  - 主比较键：四张点数
- `four_with_two_singles` / 四带两单
  - 结构：`AAAA + B + C`
  - 主比较键：四张点数

#### 12.1.4 连续牌型

- `straight` / 顺子
  - 结构：`ABCDE...`
  - 至少 5 张连续单牌
  - 最高到 `A`
  - 不允许包含 `2`、`BlackJoker`、`RedJoker`
- `serial_pairs` / 连对
  - 结构：`AA BB CC...`
  - 至少 3 组连续对子
  - 不允许包含 `2`、`BlackJoker`、`RedJoker`
- `plane_base` / 三连 / 飞机基座
  - 结构：`AAA BBB ...`
  - 至少 2 组连续三张
  - 不允许包含 `2`、`BlackJoker`、`RedJoker`

#### 12.1.5 飞机

- `plane_base` / 飞机不带
  - 结构：`AAA BBB ...`
  - 主比较键：最高三张点数
- `plane_with_singles` / 飞机带单
  - 结构：`AAA BBB + X + Y + ...`
  - 带牌数量必须与三连组数一致
  - 主比较键：最高三张点数
- `plane_with_pairs` / 飞机带对
  - 结构：`AAA BBB + CC + DD + ...`
  - 带牌对数必须与三连组数一致
  - 主比较键：最高三张点数
- 飞机比较规则
  - 先比较三连组数
  - 组数相同，再比较最高三张点数
  - 带牌部分不参与大小比较，只参与结构校验

#### 12.1.6 炸弹与特殊压制

- `bomb_four` / 四张炸弹
  - 结构：`AAAA`
- `bomb_five_plus` / 5+ 炸弹
  - 结构：`AAAAA...`
  - 张数更多优先级更高；同张数时再比较点数
- `rocket` / 王炸
  - 结构：`BlackJoker + RedJoker`
  - 最高牌型
- `pure_laizi_bomb` / 纯赖子炸弹
  - 结构：`LLLL`
  - 4 张来自同一赖子点数
  - 不允许天赖与地赖混合
- `laizi_substitute_bomb` / 赖子替代炸弹
  - 结构：4 张同点数，其中至少 1 张由赖子替代形成

#### 12.1.7 炸弹优先级

- `rocket`
- `bomb_five_plus`
- `pure_laizi_bomb`
- `bomb_four`
- `laizi_substitute_bomb`

#### 12.1.8 常规牌型优先级

- `rocket`
- `bomb`
- `plane`
- `straight`
- `serial_pairs`
- `triple_with_pair`
- `triple_with_single`
- `triple`
- `pair`
- `single`

#### 12.1.9 可比较与不可比较规则

- 默认仅同结构牌型可比较。
- 同结构的含义应精确到“带牌结构”级别，而不是仅大类名称。
  - `triple_with_single` 不可与 `triple_with_pair` 比较。
  - `plane_base` 不可与 `plane_with_singles` 比较。
  - `plane_with_singles` 不可与 `plane_with_pairs` 比较。
  - `four_with_two_pairs` 与 `four_with_two_singles` 默认视为不同结构，不可直接比较。
- 特殊压制例外：
  - `rocket` 可压所有普通炸弹和普通牌型。
  - 炸弹可压非炸弹普通牌型。
  - 炸弹内部按炸弹优先级比较。

#### 12.1.10 赖子规则

- 赖子仅能替代非王点数。
- 单独打出单张赖子时，按其原始牌处理。
- 组合判定时，赖子可参与替代形成合法牌型。
- 若同一手牌存在多个可行替代结果：
  - 必须保留全部候选解释
  - 测试阶段必须验证候选解释集合
  - 前端展示阶段应预留多解展示能力
- 对比较函数而言：
  - 若多种解释都合法，应返回所有候选解释
  - 上层测试可据此验证“是否存在可赢解释”“是否存在多个同优解释”

#### 12.1.11 与当前项目关系

- 当前项目尚未实现完整出牌阶段引擎。
- 当前阶段的目标是：
  - 固化规则文本
  - 固化规则元数据
  - 为后端测试和前端展示提供统一事实源
  - 为未来的出牌判定与比较实现预留明确接口

### 12.2 B. 后端规则模型字段映射建议

#### 12.2.1 `Catalog`

- `version`
- `rankOrder`
- `sequenceHigh`
- `notes`
- `sections`
- `bombPriority`
- `handPriority`
- 建议新增：
  - `testModeNotes`
  - `comparisonNotes`
  - `laiziResolutionNotes`

#### 12.2.2 `HandRule`

- 已有字段可继续保留：
  - `key`
  - `name`
  - `pattern`
  - `description`
  - `minCards`
  - `notes`
- 建议新增字段：
  - `category`
  - `compareBy`
  - `groupSize`
  - `kickerShape`
  - `comparableGroup`
  - `allowsLaiziSubstitution`
  - `returnsAllInterpretations`
  - `disallowRanks`
  - `specialCompareNotes`

#### 12.2.3 `BombPriority`

- 保留：
  - `rank`
  - `key`
  - `name`
  - `description`
  - `notes`
- 建议新增：
  - `compareTier`
  - `intraTierCompareBy`

#### 12.2.4 比较结果承载结构建议

- 建议未来比较测试或判定模块返回统一结构，例如：
  - `matchedType`
  - `comparable`
  - `relation`
  - `reason`
  - `mainKey`
  - `tieBreakKeys`
  - `laiziUsed`
  - `allCandidateInterpretations`

### 12.3 C. preorder 测试对象清单

#### 12.3.1 可比较对象分类

- 同型单张：`3 < 4 < ... < RedJoker`
- 同型对子：`33 < 44 < ...`
- 同型三张：`333 < 444 < ...`
- 同型三带一
- 同型三带二
- 同型顺子：同长度比较
- 同型连对：同组数比较
- 同型飞机不带：同组数比较最高三张
- 同型飞机带单：同组数比较最高三张
- 同型飞机带对：同组数比较最高三张
- 同型四带两对
- 同型四带两单
- 炸弹内部比较

#### 12.3.2 不可比较对象分类

- `single` vs `pair`
- `pair` vs `triple`
- `triple_with_single` vs `triple_with_pair`
- `straight` vs `serial_pairs`
- `plane_base` vs `plane_with_singles`
- `plane_with_singles` vs `plane_with_pairs`
- `four_with_two_pairs` vs `four_with_two_singles`
- 非炸弹普通牌型 vs 不同结构普通牌型

#### 12.3.3 同型比较链

- 单张链：
  - `3 < 4 < ... < A < 2 < BlackJoker < RedJoker`
- 对子链：
  - `33 < 44 < ... < AA < 22`
- 三张链：
  - `333 < 444 < ... < AAA < 222`
- 顺子链：
  - `34567 < 45678 < ... < 10JQKA`
- 连对链：
  - `334455 < 445566 < ...`
- 飞机链：
  - `333444 < 444555 < ...`
  - `333444 + 5 + 6 < 444555 + 6 + 7`
  - `333444 + 55 + 66 < 444555 + 66 + 77`
- 四带链：
  - `4444 + 5 + 6 < 5555 + 6 + 7`
  - `4444 + 55 + 66 < 5555 + 66 + 77`

#### 12.3.4 炸弹特殊压制链

- `laizi_substitute_bomb < bomb_four < pure_laizi_bomb < bomb_five_plus < rocket`
- 同一炸弹层级内：
  - 先按层级
  - 对 `bomb_five_plus` 先按张数，再按点数
  - 对同层普通四炸按点数

#### 12.3.5 赖子多解展示对象

- 单赖子可补成多个不同对子候选
- 双赖子可补成多个不同三张候选
- 含赖子的顺子可能存在多个连续解释
- 含赖子的飞机可能存在多个主三连解释
- 含赖子的炸弹可能同时形成普通炸弹和更高层炸弹候选

### 12.4 仍需协调者最终确认的点

- 是否需要在 `HandPriority` 中把 `four_with_two_pairs` 与 `four_with_two_singles` 拆开单列。
  - 建议默认：在 `HandRule` 中拆开，在 `HandPriority` 中继续归并到“四带”大类或保留在组合牌型下，不强制单列。
- `bomb_five_plus` 同张数时的比较键是否只比较点数，还是还要比较赖子使用情况。
  - 建议默认：先比较张数，再比较主点数，不额外按赖子使用情况排序。
- 赖子多解返回时，是否需要排序。
  - 建议默认：需要稳定排序，优先按 `matchedType`、`mainKey`、`laiziUsed` 排序，保证测试可重复。
- 测试模式下前端是否还需要保留动作按钮。
  - 建议默认：如果直接进入 `PLAY` 且尚未实现出牌动作，可不显示现有叫抢动作；但页面需展示测试模式状态。

## 13. T2/T3/T4 执行 Brief

### 13.0 协调者审阅结论

- `T1/P1` 的规则边界与测试思路可接受，但命名层存在与现有代码不一致的问题，后续任务必须优先消除。
- 当前仓库已存在两套接近但不完全一致的命名来源：
  - `catalog.go`
  - `book.go`
- `T1/P1` 输出中又引入了第三套候选命名，例如：
  - `four_with_two_singles`
  - `serial_pairs`
  - `plane_base`
- 后续 `T2/T3/T4` 必须基于“现有代码命名优先、必要时最小范围统一重命名”的原则推进，避免规则 key 再次分叉。
- 当前明确需要重点协调的命名差异：
  - `four_with_two_pairs` vs `four_with_two_pair`
  - `serial_pairs` vs `consecutive_pairs`
  - `plane_base` / `plane_with_singles` / `plane_with_pairs` vs `airplane` / `airplane_with_solo` / `airplane_with_pairs`
  - `bomb_four` / `laizi_substitute_bomb` vs `bomb` / `mixed_laizi_bomb`
- 默认协调原则：
  - 对外展示层可以保留用户更易懂的文案命名
  - 后端内部 key 先尽量复用已有常量名
  - 如必须重命名，应由 `T2` 给出统一迁移方案，并同步影响 `T3/T4`

### 13.1 T2 / P2 后端方案任务 Brief

- 任务目标
  - 设计最小侵入的后端接入方案，使服务支持测试模式：
    - 随机发牌
    - 随机底牌
    - 随机天地赖
    - 直接进入 `PLAY`
    - 地主固定为 `P0`
  - 同时扩展规则承载结构，支持后续比较测试与赖子多解展示。
- 必读上下文
  - 本计划第 3 节、第 4 节、第 12 节
  - `backend/internal/tiandi/fsm/machine.go`
  - `backend/internal/tiandi/demo/service.go`
  - `backend/internal/tiandi/rules/catalog.go`
  - `backend/test/backend/fsm_test.go`
- 必须回答的问题
  - 测试模式开关放在 `Machine`、`Service` 还是 server 层
  - 如何在不污染默认流程的情况下直达 `PLAY`
  - 如何保证 `P0` 获得底牌并维持 20 张手牌
  - `StateResponse` 是否需要新增测试模式字段
  - 比较结果与赖子多解承载结构应如何预留
- 输出要求
  - 后端改动点清单
  - 建议新增或修改的数据结构
  - 后端测试矩阵
  - 风险点与推荐默认方案
- 验收标准
  - 不要求实现代码
  - 必须明确指出最小改动路径
  - 必须覆盖测试模式、规则承载、多解展示、接口影响

### 13.2 T3 / P3 前端方案任务 Brief

- 任务目标
  - 设计前端规则展示与测试模式提示接入方案。
  - 设计真实页面验证点和 Playwright 截图点位。
- 必读上下文
  - 本计划第 4 节、第 7 节、第 12 节
  - `web/src/App.tsx`
  - `web/src/components/RulesPanel.tsx`
  - `web/src/types.ts`
  - `test/web/app.test.tsx`
- 必须回答的问题
  - 测试模式提示区放在哪里最清晰
  - `GameState` 是否需要新增 `testMode` / `testModeLabel` 等字段
  - 规则面板是否需要展示新增比较元数据
  - 多解赖子替代目前是否只预留结构，不立即完整渲染
  - 进入 `PLAY` 后现有动作条是否隐藏或降级
- 输出要求
  - 前端改动点清单
  - 类型层与组件层字段建议
  - 页面测试方案
  - Playwright 交互步骤与截图点位
- 验收标准
  - 不要求实现代码
  - 必须明确页面上哪些元素证明测试模式已生效
  - 必须给出适配现有 `App` 数据流的最小改动路径

### 13.3 T4 / P4 测试设计任务 Brief

- 任务目标
  - 设计完整测试矩阵，覆盖：
    - 后端规则目录
    - 测试模式状态机
    - 服务/API
    - 前端展示
    - Playwright 真实交互
    - preorder 比较关系
- 必读上下文
  - 本计划第 7 节、第 8 节、第 12 节
  - `backend/test/backend/*.go`
  - `backend/cmd/tiandi-server/main_test.go`
  - `test/web/app.test.tsx`
  - `package.json`
- 必须回答的问题
  - 如何拆分“可比较性”与“偏序链”测试
  - 哪些样例应作为复用 fixture
  - 哪些断言必须落到 HTTP 层，哪些只应在纯规则层完成
  - Playwright 截图证据如何命名和组织
  - 哪些回归项必须防止误伤现有叫抢地主流程
- 输出要求
  - 测试矩阵表
  - fixture 设计
  - 断言结构建议
  - 命令级验证和页面级验证步骤
- 验收标准
  - 不要求实现代码
  - 必须可直接转化为测试任务列表
  - 必须覆盖“不可比较断言”与“赖子多解断言”

## 14. 协调者收敛结论

### 14.1 已完成的协调任务

- [x] T2/P2 后端方案已收齐并审阅
- [x] T3/P3 前端方案已收齐并审阅
- [x] T4/P4 测试方案已收齐并审阅
- [x] 三份方案的冲突点已识别并给出统一决策

### 14.2 统一决策

- 后端测试模式开关：
  - 语义在 `fsm` 层
  - 配置在 `demo.Service` 层
  - 环境接线在 `cmd/tiandi-server/main.go`
- 测试模式进入方式：
  - 直接进入 `PLAY`
  - 地主固定 `P0`
  - 发牌、底牌、天地赖均保持随机
- 前后端状态字段命名：
  - 统一采用 `testMode`
  - 不采用泛化的 `mode`
- 规则事实源：
  - `backend/internal/tiandi/rules/book.go` 为单一规则主源
  - `backend/internal/tiandi/rules/catalog.go` 为展示 DTO 投影层
- 多解赖子策略：
  - 后端比较/解释层返回全部候选解释
  - 前端本轮只预留结构，不在主页面完整展开
- preorder 测试顺序：
  - 先测试 `comparable / incomparable`
  - 再测试可比较集合内的 `less / equal / greater`

### 14.3 统一命名规范

- 四带结构：
  - `four_with_two_pairs`
  - `four_with_two_singles`
- 连对：
  - 统一使用 `serial_pairs`
- 飞机：
  - `plane_base`
  - `plane_with_singles`
  - `plane_with_pairs`
- 炸弹：
  - `bomb_four`
  - `bomb_five_plus`
  - `pure_laizi_bomb`
  - `laizi_substitute_bomb`
  - `rocket`

### 14.4 后端落地基线

- `backend/internal/tiandi/fsm/machine.go`
  - 增加 `Mode` / `Options`
  - 保留 `NewMachine(rng)`，新增 `NewMachineWithOptions(rng, opts)`
  - 在 `PhaseDeal` 分支支持测试模式直达 `PLAY`
  - 用独立 helper 完成 `P0` 吃底牌、公开天地赖、落地主状态
- `backend/internal/tiandi/demo/service.go`
  - 增加 `ServiceOptions`
  - 保留 `NewService()`，新增 `NewServiceWithOptions()`
  - `StateResponse` 增加：
    - `testMode.enabled`
    - `testMode.label`
    - `testMode.fixedLandlord`
    - `testMode.directPlay`
- `backend/cmd/tiandi-server/main.go`
  - 只负责环境配置到 `ServiceOptions` 的接线
  - 不在 handler 层硬编码测试模式逻辑
- `backend/internal/tiandi/rules/book.go`
  - 先完成 key 统一
  - 再补充 `CompareBy`、`ComparableGroup`、`ReturnsAllInterpretations`
  - 飞机统一 `CompareBy = highest_triplet_rank`
- `backend/internal/tiandi/rules/catalog.go`
  - 从 `DefaultRuleBook()` 派生关键 key、优先级和比较说明
  - 拆出“四带两对 / 四带两单”两个展示项

### 14.5 前端落地基线

- `web/src/types.ts`
  - `GameState` 新增可选 `testMode` 对象
- `web/src/App.tsx`
  - 在 `topbar` 与 `ActionBar` 之间新增测试模式提示区
  - 文案直接消费 `data.testMode.label`
- `web/src/components/ActionBar.tsx`
  - `PLAY` 且无动作时降级为：
    - 保留“新开一局”
    - 展示“当前无可执行动作”
  - 不再展示误导性的 `当前操作人：-`
- `web/src/components/RulesPanel.tsx`
  - 只展示人类可读的比较说明
  - 不直接暴露内部测试元数据

### 14.6 测试落地基线

- 纯规则层新增测试：
  - `comparability_test.go`
  - `preorder_test.go`
  - `laizi_interpretation_test.go`
- 流程层新增测试：
  - `service_test.go`
  - 扩展 `fsm_test.go`
  - 扩展 `main_test.go`
- 前端测试：
  - 扩展 `test/web/app.test.tsx`
  - 如需要，新增 `test-mode-banner.test.tsx`
- Playwright 证据：
  - 首页全图
  - 测试模式提示区
  - `P0` 地主 + 底牌区
  - 无动作态 `ActionBar`

### 14.7 下一阶段执行顺序

- [x] E1. 先实现后端测试模式与 `testMode` 状态字段
- [x] E2. 再统一 `book.go` / `catalog.go` 命名与规则投影
- [x] E3. 补后端状态机 / service / HTTP 测试
- [x] E4. 接前端测试模式提示与无动作态
- [x] E5. 补前端单测与 Playwright 联调测试
- [x] E6. 最后更新 `docs/tiandi_arch_and_api.md`
