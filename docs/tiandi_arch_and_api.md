# 天地癞子：后端/前端模块树与接口协作说明

## 1. 仓库后端模块树（核心源代码）

```text
backend/
├── go.mod
├── cmd/
│   ├── tiandi-server/
│   │   └── main.go                # HTTP Server 入口
│   └── tiandi-demo/
│       └── main.go                # 命令行演示入口
└── internal/
    └── tiandi/
        ├── agent/
        │   ├── types.go            # prompt / decision / trace DTO
        │   ├── prompt.go           # LLM prompt builder
        │   ├── match.go            # 整局 orchestrator、记牌器与单轮记忆跟踪
        │   ├── decision.go         # 决策 JSON 解析、校验与 mock runner
        │   └── openrouter.go       # OpenRouter Chat Completions 调用封装
        ├── demo/
        │   └── service.go          # 游戏服务层（状态/动作转换、testMode 输出）
        ├── domain/
        │   ├── cards.go            # 牌面、花色、点数与序列化
        │   ├── laizi.go            # 天赖/地赖选择与赖子判断
        │   └── player.go           # 座位与玩家数量
        ├── fsm/
        │   └── machine.go          # 状态机：洗牌/发牌/选赖子/叫地主/抢地主/正式 PLAY 回合流转
        ├── game/
        │   └── session.go          # 手牌分配与摸牌逻辑
        ├── play/
        │   ├── types.go            # 出牌阶段 DTO/解析结果/桌面牌结构
        │   ├── analyze.go          # 牌型识别与候选解生成
        │   └── compare.go          # 同型比较、炸弹压制、pass 约束
        ├── rules/
        │   ├── book.go             # 规则主源：牌型 key、优先级、比较元数据
        │   └── catalog.go          # 牌型规则目录、优先级与前端展示 DTO
        └── sortx/
            └── hand.go             # 自动理牌（赖子前置+排序）
├── test/
    └── backend/
        ├── agent_test.go
        ├── fsm_test.go
        ├── play_test.go
        ├── rules_test.go
        ├── service_test.go
        └── sortx_test.go
scripts/
└── tiandi-agent.sh                 # bash 直玩脚本：单步 agent + 整局 match 的 state/run/trace/reset
```

- `tiandi-server` 负责运行可调用 API 的内存态一局服务。
- `tiandi-demo` 负责本地命令行输出状态，帮助验证状态机。
- `internal/tiandi` 采用分层：
  - `domain`（实体与规则）
  - `game`（会话数据构建）
  - `fsm`（游戏流程）
  - `play`（出牌阶段判型、比较、桌面主牌、回合推进）
  - `demo`（服务层组装与对外状态转换）
  - `agent`（LLM prompt、决策解析、mock/OpenRouter runner）
  - `rules`（规则主源与前端展示 DTO）
  - `sortx`（展示层用的手牌排序）

## 2. 仓库前端模块树（核心源代码）

```text
web/
├── index.html
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
└── src/
    ├── main.tsx                  # React 启动入口
    ├── App.tsx                   # 状态拉取、动作提交与页面编排
    ├── styles.css                # 样式
    ├── types.ts                  # 前后端状态类型定义
    ├── api/
    │   └── game.ts               # 三个后端接口的调用封装
    │   └── rules.ts              # 规则目录接口调用封装
    │   └── agent.ts              # agent prompt / trace / run + match 接口封装
    └── components/
        ├── ActionBar.tsx         # 操作按钮（叫抢地主 + PLAY 阶段出牌/不出/清空选择）
        ├── AgentDebugPanel.tsx   # 单步 trace + 整局 match / 记牌器 / 单轮记忆展示
        ├── PlayerPanel.tsx       # 玩家展示块
        ├── CardStrip.tsx         # 手牌/底牌卡片列表，可选中当前行动位手牌
        ├── BottomPanel.tsx       # 底牌展示
        ├── CurrentTrickPanel.tsx # 当前桌面牌 / 最近解析 / 候选解 / 错误与胜者信息
        └── RulesPanel.tsx        # 牌型规则展示面板
```

测试目录（与前端功能协作）在：
- `test/web/app.test.tsx`（集成式交互测试：首次加载、规则目录渲染、agent 调试面板、PLAY 阶段选牌与 mock agent 执行）

## 3. 接口清单（前后端协作入口）

| 接口 | 方法 | 请求 | 响应 | 说明 |
|---|---|---|---|---|
| `/api/game/state` | `GET` | 无 | `GameState` | 获取当前局面状态 |
| `/api/game/reset` | `POST` | 无 | `GameState` | 重开一局，并返回新状态 |
| `/api/game/action` | `POST` | `ActionRequest` | `GameState` | 提交一条动作并返回执行后的状态 |
| `/api/game/rules` | `GET` | 无 | `RulesCatalog` | 获取牌型规则目录与优先级定义 |
| `/api/game/agent/prompt` | `GET` | 无 | `PromptResponse` | 获取当前局面对应的 system/user prompt 与动作协议 |
| `/api/game/agent/trace` | `GET` | 无 | `TraceEnvelope` | 获取最近一次 agent 执行轨迹 |
| `/api/game/agent/run` | `POST` | `RunRequest` | `RunResponse` | 使用 `mock` 或 `openrouter` 运行一次 agent 决策并尝试提交动作 |
| `/api/game/agent/match` | `GET/POST` | `MatchRunRequest?` | `MatchStateResponse` / `MatchRunResponse` | 兼容入口，支持查询整局状态或直接运行整局 |
| `/api/game/agent/match/state` | `GET` | 无 | `MatchStateResponse` | 获取当前牌局状态和最近一次整局 trace |
| `/api/game/agent/match/run` | `POST` | `MatchRunRequest` | `MatchRunResponse` | 让三个 seat 连续决策直到 `winner` 或超出 `maxSteps` |
| `/api/game/agent/match/trace` | `GET` | 无 | `MatchEnvelope` | 获取最近一次整局执行 trace |
| `/api/game/agent/match/reset` | `POST` | 无 | `MatchStateResponse` | 重开一局并清空最近一次整局 trace |

- `ActionRequest`（后端 `demo.Service.Apply` 输入）
  - `seat: string`（`P0/P1/P2`）
  - `kind: string`（`jiaodizhu | bujiao | qiangdizhu | buqiang | woqiang | play | pass`）
  - `cards?: string[]`（`kind = play` 时提交当前选中的原始牌 id）
  - `resolutionId?: string`（保留给多候选解析结果选择）
- `GameState`（前端 `web/src/types.ts`）核心字段
  - `phase`, `currentActor`, `availableActions`, `message`
  - `players[]`（`seat/isLandlord/isCurrent/cards`）
  - `landlord`, `multiplier`
  - `laizi`（`tian/di`, `tianVisible/diVisible`）
  - `bottom`（`visible/count/cards`）
  - `currentTrick?`
    - `leadingSeat`
    - `lastPlaySeat`
    - `passCount`
    - `cards`
    - `resolvedHand`
  - `resolvedHand?`（最近一次动作的标准解析）
  - `resolutionCandidates[]`（最近一次动作的全部候选解）
  - `playError`（最近一次出牌失败原因）
  - `winner`（胜者 seat，未结束为空串）
  - `testMode?`
    - `enabled`
    - `label`
    - `fixedLandlord`
    - `directPlay`
- `RulesCatalog`（前端 `web/src/types.ts`）核心字段
  - `version`
  - `rankOrder`
  - `sequenceHigh`（当前为 `A`，表示顺子最高到 `A`）
  - `notes[]`
  - `comparisonNotes[]`
  - `laiziResolutionNotes[]`
  - `sections[]`：按“基础牌型 / 组合牌型 / 连续牌型 / 飞机扩展 / 炸弹与特殊压制”分组
  - `bombPriority[]`
  - `handPriority[]`
- `PromptResponse`
  - `modelHint`
  - `state`
  - `phase`, `currentActor`, `availableActions`
  - `playerSeat`, `playerRole`, `playerHand`
  - `cardCounter`
  - `roundMemory`
  - `systemPrompt`, `userPrompt`
  - `actionSchema`
- `CardCounter`
  - `playedCardsBySeat`
  - `playedRankCounts`
  - `remainingUnknown`
  - `blackJokerPlayed`, `redJokerPlayed`
  - `bombSignals`
  - `totalPlayedCardCount`
- `SeatRoundMemory`
  - `seat`, `seatRole`
  - `roundIndex`, `trickIndex`
  - `lastSelfPlay`
  - `lastTeammatePlay`
  - `lastOpponentPlay`
- `TraceEnvelope`
  - `trace?`
- `Trace`
  - `runId`, `createdAt`
  - `mode`, `model`
  - `prompt`
  - `rawResponse`
  - `decision`
  - `applied`
  - `error`
  - `resultMessage`
  - `resultState?`
- `RunRequest`
  - `mode: "mock" | "openrouter"`
  - `model?: string`
- `RunResponse`
  - `state`
  - `trace`
- `MatchRunRequest`
  - `mode: "mock" | "openrouter"`
  - `model?: string`
  - `fallbackMode?: string`
  - `maxSteps?: number`
  - `resetGame?: boolean`
- `MatchTrace`
  - `matchId`, `startedAt`, `finishedAt`
  - `status`, `mode`, `fallbackMode`, `model`
  - `winner`
  - `steps[]`
  - `finalState?`
  - `error?`
  - `stepCount`
- `MatchStep`
  - `stepIndex`, `seat`
  - `attemptMode`, `effectiveMode`, `model`
  - `prompt`, `decision`, `applied`, `error`, `resultMessage`
  - `roundIndex`, `trickIndex`
  - `cardCounter`, `roundMemory`
  - `stateBefore`, `stateAfter`
- `MatchStateResponse`
  - `state`
  - `match?`

## 3.1 测试模式约定

- 后端支持通过 `TIANDI_TEST_MODE=fixed_p0_play_test` 启用测试模式。
- 后端支持通过 `TIANDI_SERVER_ADDR` 或 `PORT` 指定监听地址，便于并行跑多份本地服务做集成测试。
- 测试模式下：
  - 洗牌仍随机
  - 底牌仍随机
  - 天赖/地赖仍随机
  - 状态机在发牌后直接进入 `PLAY`
  - 地主固定为 `P0`
  - `P0` 直接吃到底牌，因此手牌为 `20` 张
  - 前端通过 `GameState.testMode` 显式展示“当前为测试模式 / 固定地主为 P0 / 直接进入 PLAY”
  - 当前行动位固定为 `P0`
  - 初始 `availableActions = ["play"]`

## 4. 协作流程（请求流）

### 4.1 页面初始化
1. 前端 `App.tsx` `useEffect` 中并行调用 `fetchGameState()`、`fetchRulesCatalog()`、`fetchAgentPrompt()`、`fetchAgentTrace()` 与 `fetchAgentMatchTrace()`。
2. `api/game.ts` 发起 `GET /api/game/state`；`api/rules.ts` 发起 `GET /api/game/rules`；`api/agent.ts` 发起 `GET /api/game/agent/prompt`、`GET /api/game/agent/trace` 与 `GET /api/game/agent/match/trace`。
3. 后端 `tiandi-server/main.go` 分别路由到 `handleState`、`handleRules`、`handleAgentPrompt`、`handleAgentTrace`、`handleAgentMatchTrace`。
4. `handleState` 返回当前局面；`handleRules` 返回静态规则目录；`handleAgentPrompt` 基于当前局面构造 prompt；`handleAgentTrace` 返回最近一次单步 trace；`handleAgentMatchTrace` 返回最近一次整局 trace。
5. 前端根据局面渲染 Header、ActionBar、玩家区与底牌区；规则目录成功时渲染 `RulesPanel`；agent 调试接口成功时渲染 `AgentDebugPanel`，失败时只在调试区域降级提示，不阻塞主界面。

### 4.2 刷新局面 / 新开局
1. 用户点击“新开一局”。
2. 前端调用 `resetGame()` -> `POST /api/game/reset`。
3. 后端 `handleReset -> service.Reset()`。
4. `service.Reset()` 重建 `fsm.Machine` 并 `Start()`。
5. `buildState(snapshot)` 返回标准化状态给前端。

### 4.2.1 测试模式刷新局面
1. 后端通过 `demo.NewServiceWithOptions(...)` 持有测试模式配置。
2. `service.Reset()` 用 `fsm.NewMachineWithOptions(...)` 创建状态机。
3. `fsm.Machine` 在 `PhaseDeal` 分支检测测试模式后，直接进入 `PLAY`。
4. `buildState(snapshot)` 输出 `testMode` 元数据给前端。

### 4.3 玩家动作提交
1. 用户点击合法动作按钮（由 `data.availableActions` 动态渲染）。
2. 前端调用 `applyGameAction({ seat: data.currentActor, kind, cards? })`。
3. 后端 `handleAction` 校验请求 JSON，构造 `demo.ActionRequest`。
4. `service.Apply` 做：
   - `ParseSeat(req.Seat)`
   - `cards` id 映射回当前玩家真实手牌
   - `machine.Apply(fsm.Action{Seat, Kind, Cards, ResolutionID})`
   - `buildState(snapshot)`
5. 状态返回后前端覆盖本地 `ready` 数据并重渲染。

### 4.3.1 PLAY 阶段动作流
1. 当前行动位选择若干张自己的手牌。
2. 前端提交：
   - `kind = "play"`，并带 `cards`
   - 或 `kind = "pass"`，仅在桌面已有主牌时允许
3. 后端 `play.AnalyzeSelection(...)` 生成 `resolutionCandidates`。
4. 后端选定标准解：
   - 默认取排序第一项
   - 若候选中存在炸弹解释，炸弹优先
5. 若桌面已有主牌，则 `play.Beats(...)` 校验是否压过。
6. 合法出牌后：
   - 从当前玩家手牌移除所选牌
   - 更新 `currentTrick`
   - 更新 `lastResolvedHand` 与 `lastCandidates`
   - `currentActor` 切到下一家
7. 若执行 `pass`：
   - `passCount += 1`
   - 两家连续 `pass` 后清空桌面主牌
   - 由最近一次成功出牌者重新首出
8. 若某玩家手牌清空：
   - `winner` 固定
   - `availableActions` 变为空
   - 局面进入结束态

### 4.3.2 Agent Prompt / Trace 协作
1. 后端 `agent.BuildPrompt(state, rules)` 生成固定的 `systemPrompt`、动态 `userPrompt`、动作 schema 和当前玩家手牌 id 列表。
2. `GET /api/game/agent/prompt` 直接返回 `PromptResponse`，供 bash 脚本、前端调试面板或外部 LLM runner 消费。
3. `POST /api/game/agent/run` 支持两种模式：
   - `mock`：使用当前玩家第一张牌或第一条非出牌动作作为保守合法决策
   - `openrouter`：向 OpenRouter `chat/completions` 发送 system/user prompt，并解析返回 JSON 动作
4. 后端先保存 `Trace`，再校验 decision：
   - `seat` 必须等于 `currentActor`
   - `kind` 必须属于 `availableActions`
   - `cards` 只能来自当前玩家手牌
5. 校验通过后调用现有 `service.Apply`，把结果局面写入 `RunResponse.state`，并把执行结果回填到 `Trace`。
6. 前端 `AgentDebugPanel` 和 `scripts/tiandi-agent.sh trace` 都通过 `GET /api/game/agent/trace` 读取最近一次 trace。

### 4.3.3 整局 Agent Match 协作
1. 前端或 bash 调用 `POST /api/game/agent/match/run`，或兼容地调用 `POST /api/game/agent/match`。
2. 后端 `agent.Service.RunMatch(...)` 进入循环：
   - 读取当前 `state`
   - 为当前 `currentActor` 组装 seat-specific prompt
   - 把完整当前手牌、`CardCounter`、`SeatRoundMemory` 注入 prompt
   - 运行 `mock` 或 `openrouter`
   - 决策合法后调用现有 `demo.Service.Apply(...)`
3. 每一步都记录到 `MatchStep`：
   - seat
   - 决策前后状态
   - 当前轮次/手次
   - seat 对应的记牌器摘要
   - seat 对应的单轮记忆
4. 记牌器由后端根据公共已出牌信息增量更新；单轮记忆只保留当前轮的 `lastSelfPlay / lastTeammatePlay / lastOpponentPlay`，清轮后重置。
5. 若 `openrouter` 调用失败且配置了 `fallbackMode`，本步会降级到 `mock`，但整局 trace 会保留 `attemptMode`、`effectiveMode` 和错误信息。
6. 实际联调中，免费模型 `qwen/qwen3.6-plus-preview:free` 已验证可成功返回合法 JSON 并推进多手，但中后盘仍可能出现非法动作；因此整局自动化建议保留 `fallbackMode=mock` 作为兜底。
7. 前端 `AgentDebugPanel` 用整局 trace 展示：
   - winner
   - stepCount
   - 最近动作/结果
   - 3 seat 剩余手牌
   - 记牌器摘要
   - 上一轮记忆
8. bash 可通过 `scripts/tiandi-agent.sh match-state|match-run|match-run-mock|match-trace|match-reset` 观察同一份整局结果。

### 4.4 规则目录协作
1. 后端规则主源在 `rules/book.go`。
2. `rules/catalog.go` 作为前端展示 DTO，负责把规则主源映射成 `RulesCatalog`。
3. 当前目录中已显式包含：
   - `four_with_two_pairs`
   - `four_with_two_singles`
   - 飞机比较主键 `highest_triplet_rank`
   - 赖子多解说明

## 5. 关键数据流职责边界（为什么前端不排序）

- 服务器侧在 `service.buildState` 中调用 `sortx.SortedHand` 做排序，前端直接按收到顺序渲染。
- 因此 `CardStrip` 只做展示/选中组件，不再包含排序、规则判断。
- 赖子标注也来源于后端标注后的 `CardView.isLaizi`。
- 新增的牌型规则目录同样由后端统一定义，前端只负责渲染，不在前端重复维护炸弹优先级和牌型名称。
- `testMode` 同样以后端为准，前端不从 `phase === PLAY && landlord === P0` 反推是否为测试模式。
- 前端提交的 `cards` 只是原始牌 id；所有判型、比较、赖子求解都由后端负责。
- agent prompt 也以后端为准，前端只展示 prompt 与 trace，不自行拼接规则摘要。
- bash 脚本不做规则判断，只是对现有 HTTP 契约的命令行封装。
- OpenRouter 缺少 `OPENROUTER_API_KEY` 时，后端保留 `openrouter` 模式入口，但会把错误写入 trace，而不是让前端或脚本崩溃。

## 6. 测试策略

- 后端规则模块单测：
  - `backend/test/backend/rules_test.go`
  - 验证必需牌型是否全部存在。
  - 验证炸弹优先级顺序是否与规则文档一致。
  - 验证顺子最高点 `sequenceHigh` 是否为 `A`。
  - 验证比较说明与赖子多解说明字段存在。
- 后端出牌模块单测：
  - `backend/test/backend/play_test.go`
  - 验证单张识别、三带类与四带类按主干比较、`play/pass` 链路、两家 `pass` 后清轮。
- 后端状态机 / service 测试：
  - `backend/test/backend/fsm_test.go`
  - `backend/test/backend/service_test.go`
  - 验证测试模式直接进入 `PLAY`。
  - 验证地主固定为 `P0`，且 `P0` 手牌为 `20` 张。
  - 验证 `testMode` 元数据输出完整。
  - 验证测试模式初始 `currentActor = P0`，且 `availableActions = ["play"]`。
- 后端 HTTP 接口测试：
  - `backend/cmd/tiandi-server/main_test.go`
  - 验证 `GET /api/game/rules` 返回成功且 JSON 结构可解析。
  - 验证 `POST /api/game/rules` 被正确拒绝。
  - 验证测试模式下 `GET /api/game/state` 返回 `PLAY + P0 + testMode + play`。
  - 验证 `GET /api/game/agent/prompt` 返回当前玩家、system prompt 和手牌摘要。
  - 验证 `POST /api/game/agent/run` 在 `mock` 模式下能推进局面并写入 trace。
- 前端模块测试：
  - `test/web/app.test.tsx`
  - 验证游戏状态与规则目录并行加载后的正常渲染。
  - 验证 agent 调试面板加载 prompt 与 trace。
  - 验证 agent 调试接口失败时主牌桌仍可继续使用。
  - 验证规则目录中的代表性牌型可见。
  - 验证动作提交仍按既有 `/api/game/action` 路径工作。
  - 验证 `PLAY` 阶段可选牌并提交 `play` 请求。
  - 验证桌面已有主牌时显示 `不出`。
  - 验证点击 `运行 Mock Agent` 后调试面板与牌桌同步刷新。
- 浏览器联调验证：
  - 使用 Playwright interactive 流程验证真实页面。
  - 验证 Agent 调试面板可见。
  - 验证点击 `运行 Mock Agent` 后页面可见最近 trace 与结果。
  - 保留截图证据：`artifacts/playwright/11-agent-panel-before.png`、`artifacts/playwright/12-agent-panel-after.png`
- 集成验证：
  - `go test ./...`
  - `npm test`
  - `npm run build`
  - `./scripts/tiandi-agent.sh prompt`
  - `./scripts/tiandi-agent.sh run-mock`
  - `./scripts/tiandi-agent.sh trace`

## 7. 运行与联调环境说明

- 前端默认通过 `import.meta.env.VITE_API_BASE_URL` 组装接口。
- 开发模式下 `web/vite.config.ts` 已配置：
  - Vite 开发服务器端口：`5173`
  - 代理：`/api -> http://localhost:8080`
- 后端监听端口：`:8080`。
- 推荐联调方式（本仓库脚本）：
  - 一键启动：`./restart.sh`
  - 后端：`TIANDI_TEST_MODE=fixed_p0_play_test npm run backend:dev`
  - 前端：`npm run dev -- --host 127.0.0.1`
