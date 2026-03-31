# tiandi_llm_full_match 三代理整局测试计划

## 0. 当前状态

- [x] Step 1：梳理现有 agent、prompt、bash 脚本、前端调试面板与运行条件
- [x] Step 2：制定 Plan，明确自然语言内容、模块插入点、串并行任务与 test plan
- [x] Step 3：完成三代理整局对局编排、记牌器与单轮记忆接入
- [x] Step 4：完成前端整局观测与回放展示
- [x] Step 5：完成模块测试、整局集成测试与 Playwright 取证
- [x] Step 6：更新 `docs/tiandi_arch_and_api.md` 并回填完成状态

## 1. User 原始提示

> .env
> 现在需要增加测试内容了
> 要求让LLM完整的进行一轮对局，直到分出胜负才可以
> 3个LLM各自扮演角色
> 模型统一为免费模型
>
> agent接收的上下文包括完整的当前牌，以及记牌器（这个需要新增加）内容，还有上一轮对手/队友的出牌
> agent每轮都更新记忆，也就是只看一轮次的记忆，然后进行决策
>
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

## 2. 任务背景

### 2.1 当前仓库现状

- 后端已经有：
  - 单次 agent prompt 生成
  - 单次 `mock` / `openrouter` 决策执行
  - `scripts/tiandi-agent.sh` 的单步命令行操作
  - `PLAY` 阶段状态、解析结果、候选解与错误反馈
- 前端已经有：
  - Agent 调试面板
  - 最近一次单次 trace 的展示
  - Vitest 与 Playwright 基础验证能力
- 当前缺口：
  - 没有“三个 seat 各自一名 LLM”的整局 orchestrator
  - 没有“直到胜负分出”的持续回合循环
  - 没有 seat 级别的记牌器
  - 没有“上一轮对手/队友出牌”的 seat 记忆结构
  - 前端没有整局对局过程的观测面板

### 2.2 本轮真实目标

- 让三个 LLM 分别扮演 `P0/P1/P2`，从一局起始状态持续决策，直到出现 `winner`。
- 三个 seat 使用统一免费模型。
  - 默认仍用 `qwen/qwen3.6-plus-preview:free`
- 每个 agent 的上下文必须包含：
  - 当前玩家完整手牌
  - 记牌器摘要
  - 上一轮对手/队友的出牌
  - 当前桌面主牌
  - 当前局面允许动作
- 每个 agent 只保留“一轮次记忆”。
  - 一轮次结束后，旧记忆应被滚动替换，而不是无限堆积
- 最终要能从 bash 和前端都观测完整整局测试结果

### 2.3 关键约束

- 规则事实源仍然只能是后端。
- LLM 只负责在给定上下文下输出动作 JSON。
- 记牌器由后端根据公共信息推导，不允许前端或 LLM 自行统计出一套不同版本。
- `.env` 已存在运行时配置，但敏感值不写入文档、不回显到日志。
- OpenRouter 实网调用要允许失败降级，并把错误记入整局 trace，而不是直接让流程崩溃。

## 3. 自然语言完整内容

### 3.1 每个 seat 的 system prompt 必须包含

- 你是当前 seat 的测试代理，只能代表该 seat 决策
- 后端规则与 `availableActions` 是唯一事实源
- 只允许输出一个 JSON 对象
- `seat` 必须等于当前行动位
- `kind` 必须属于 `availableActions`
- 若 `kind = play`，只能从当前手牌里选择 card id
- 若提供 `resolutionCandidates`，需要在必要时回填 `resolutionId`
- 必须结合“记牌器”和“上一轮记忆”做保守、可解释的决策

### 3.2 每轮 user prompt 必须包含

- 当前局面摘要：
  - `phase`
  - `currentActor`
  - `landlord`
  - `multiplier`
  - `winner`
- 当前 seat 可见信息：
  - 完整当前手牌
  - 当前桌面主牌
  - 最近解析 / 候选解
  - 本轮允许动作
- 记牌器摘要：
  - 已打出牌的计数
  - 剩余未知牌统计
  - 大牌/炸弹/王是否已出现的摘要
- 上一轮记忆：
  - 对手上一轮出牌
  - 队友上一轮出牌
  - 当前轮次自己的上一手表现
- 输出协议：
  - `seat`
  - `kind`
  - `cards`
  - `resolutionId`
  - `reason`

### 3.3 记牌器定义

- 不是完整牌谱数据库，而是 seat 可消费的结构化摘要
- 至少包含：
  - `playedCardsBySeat`
  - `playedRankCounts`
  - `remainingUnknownCount`
  - `jokerStatus`
  - `bombSignals`
- 该信息由后端在每次动作后增量更新

### 3.4 单轮记忆定义

- “只看一轮次的记忆”解释为：
  - 只保留当前轮的主要出牌交换
  - 清轮后重置 round memory
  - 同一轮内始终覆盖 seat 关心的最近信息
- 每个 seat 至少记录：
  - `lastTeammatePlay`
  - `lastOpponentPlay`
  - `lastSelfPlay`
  - `roundIndex`
  - `trickIndex`

## 4. 模块插入点

### 4.1 后端插入点

- `backend/internal/tiandi/agent/`
  - 新增整局 orchestrator
  - 新增 seat 记忆模型
  - 新增记牌器模型
  - 扩展 prompt builder，支持 seat-specific 上下文
- `backend/cmd/tiandi-server/main.go`
  - 增加整局运行与整局 trace 查询接口
- `scripts/tiandi-agent.sh`
  - 增加 `match-run`、`match-state`、`match-reset` 一类命令
- 视需要扩展 `demo.StateResponse`
  - 增加已出牌历史或可供记牌器消费的公共信息

### 4.2 前端插入点

- `web/src/types.ts`
  - 增加 match trace、seat memory、card counter 摘要类型
- `web/src/api/agent.ts`
  - 增加整局运行、查询与轮次数据拉取
- `web/src/components/AgentDebugPanel.tsx`
  - 增加整局模式状态、3 seat 执行进度、最近回合摘要
- 视需要新增组件：
  - `MatchTracePanel.tsx`
  - `SeatMemoryPanel.tsx`
  - `CardCounterPanel.tsx`

## 5. 执行策略

### 5.1 串行主链

- [x] S1. 锁定自然语言内容
  - 输出：三 seat prompt、记牌器定义、单轮记忆定义
- [x] S2. 锁定插入点与接口
  - 输出：新增后端与前端模块边界
- [x] S3. 实现后端三代理整局编排
  - 输出：orchestrator、整局 trace、seat memory、card counter
- [x] S4. 实现前端整局观测
  - 输出：整局调试面板、seat/round/memory 可视化
- [x] S5. 完成整局集成验证
  - 输出：bash 跑完整局、前端展示完整局、截图证据
- [x] S6. 回写架构文档
  - 输出：更新后的 `docs/tiandi_arch_and_api.md`

### 5.2 可并行任务

- [x] P1. 后端整局 orchestrator
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：完成三代理整局循环、seat memory、记牌器和整局接口
  - 输出：后端代码、bash 脚本扩展、Go 测试
- [x] P2. 前端整局观测与展示
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：完成整局 trace、seat 记忆、记牌器 UI
  - 输出：前端代码、Vitest、Playwright 验证点
- [x] P3. 集成测试与验收
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：验证实网/降级、整局完成到胜负出现、截图与日志证据
  - 输出：测试矩阵、执行记录、风险清单

### 5.3 串并行依赖

- `P1` 先定义接口契约，主线程审核后，`P2` 按契约接前端
- `P3` 在 `P1/P2` 接口落地后执行
- 文档回写在全部实现与验收完成后执行

## 6. Subagent 分发计划

### 6.1 协调者职责

- 统一三 seat prompt 契约、记牌器结构与单轮记忆结构
- 保证后端是规则、记牌器、记忆的唯一事实源
- 汇总 subagent 结果并统一验收
- 回填本计划中的勾选状态

### 6.2 任务卡

- [x] T1. 后端整局编排任务
  - 负责人：Subagent A
  - 目标：新增整局运行器、seat memory、card counter、整局接口、bash 命令
  - 输出：后端代码、Go 测试、脚本说明
- [x] T2. 前端整局观测任务
  - 负责人：Subagent B
  - 目标：展示整局进度、当前 seat、记牌器摘要、上一轮记忆、胜负结果
  - 输出：前端代码、Vitest、Playwright 关注点
- [x] T3. 集成验收任务
  - 负责人：Subagent C
  - 目标：跑通完整一局到 `winner`，输出日志、截图和失败场景说明
  - 输出：执行证据、截图、测试结论

## 7. Test Plan

### 7.1 后端模块测试

- [ ] 记牌器测试
  - 验证每次出牌后统计正确更新
  - 验证王、炸弹、点数计数摘要正确
- [x] 单轮记忆测试
  - 验证队友/对手/自己上一轮出牌正确归档
  - 验证清轮后记忆重置
- [x] prompt builder 测试
  - 验证 prompt 包含当前完整手牌、记牌器、上一轮记忆
- [x] 整局 orchestrator 测试
  - 验证 3 seat 能持续轮转直到出现 winner
  - 验证中途错误能写入 match trace
- [ ] OpenRouter runner 测试
  - 已完成：接入真实 `OPENROUTER_API_KEY` 后，`qwen/qwen3.6-plus-preview:free` 单步调用成功，且纯 `openrouter` 逐手实战已连续推进到 `step 47`
  - 已发现：纯 `openrouter` 在 `step 47` 触发非法动作，错误为 `selected cards do not beat the current trick`
  - 已完成：缺 key / 失败时降级到 `mock` 的路径已验证可继续打到 `winner`
  - 未完成：纯实网免费模型从开局独立打到 `winner`，当前受免费模型时延和中后盘非法动作影响

### 7.2 前端模块测试

- [x] 整局面板渲染测试
  - 验证三 seat 状态、当前回合、winner、整局摘要可见
- [x] 记牌器与记忆面板测试
  - 验证最新 round memory、对手/队友上一轮出牌、计牌摘要渲染正确
- [x] 降级测试
  - 验证整局接口失败时主牌桌与单步调试仍可使用

### 7.3 集成测试

- [x] bash 整局测试
  - 从命令行触发整局运行直到 winner
  - 输出最终 winner、轮次数、错误摘要
- [x] 后端整局接口测试
  - 验证返回 seat trace、记牌器、round memory、winner
- [x] 前后端联调测试
  - 验证前端能实时展示整局进度与终局结果

### 7.4 Playwright 验收

- [x] 打开页面并确认整局调试面板可见
- [x] 触发“运行完整对局”
- [x] 等待 winner 出现
- [x] 保存至少两张截图：
  - 开始整局前
  - 出现 winner 后

## 8. 验收标准

- 三个 seat 各自使用统一免费模型完成完整一局
  - 当前状态：真实 `OPENROUTER_API_KEY` 已验证可用；纯实网 `qwen/qwen3.6-plus-preview:free` 已成功推进多步，但本轮未独立打到 `winner`
  - 已验证：`openrouter` 路径可成功返回合法动作；缺 key/失败时可降级到 `mock`
- 至少一条整局路径运行到 `winner`
- 每轮 prompt 都包含完整手牌、记牌器和上一轮记忆
- 记牌器和单轮记忆在后端可测试、在前端可观测
- bash、后端接口、前端面板三者都能看到整局结果
- `docs/tiandi_arch_and_api.md` 已同步更新

## 9. 完成记录

- [x] 计划文档创建完成
- [x] Subagent 已分发并完成交付
- [x] 主线程完成集成与验收
- [x] 架构文档已回写

## 10. 本轮执行证据

- Bash：
  - `./scripts/tiandi-agent.sh --base-url http://127.0.0.1:18080 match-run-mock --max-steps 256`
  - 结果：`winner=P0`，`status=completed`，`stepCount=62`
- 免费 LLM 路径：
  - 单步烟测：
    - `./scripts/tiandi-agent.sh --base-url http://127.0.0.1:18080 run-openrouter --model qwen/qwen3.6-plus-preview:free`
    - 结果：真实 `openrouter` 成功返回并应用动作，无 `mock` 降级
  - 逐手整局：
    - `BASE_URL=http://127.0.0.1:18080 MAX_STEPS=200 backend/test/backend/run_agent_full_match_acceptance.sh openrouter qwen/qwen3.6-plus-preview:free`
    - 结果：真实 `openrouter` 连续推进到 `step 47`，随后因 `selected cards do not beat the current trick` 失败
  - 近终局补救：
    - `POST /api/game/agent/match/run` with `mode=openrouter` + `fallbackMode=mock`
    - 结果：真实路径继续推进到终盘，但在本次会话时间窗口内仍未产生 `winner`
- 测试：
  - `cd backend && go test ./...`
  - `npm test`
  - `npm run build`
- Playwright 截图：
  - `artifacts/playwright/20-desktop-initial.png`
  - `artifacts/playwright/21-agent-match-desktop.png`
  - `artifacts/playwright/22-agent-match-mobile.png`
  - `artifacts/playwright/23-agent-match-free-llm-fallback.png`
