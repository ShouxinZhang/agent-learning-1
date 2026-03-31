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
        ├── demo/
        │   └── service.go          # 游戏服务层（状态/动作转换、testMode 输出）
        ├── domain/
        │   ├── cards.go            # 牌面、花色、点数与序列化
        │   ├── laizi.go            # 天赖/地赖选择与赖子判断
        │   └── player.go           # 座位与玩家数量
        ├── fsm/
        │   └── machine.go          # 状态机：洗牌/发牌/选赖子/叫地主/抢地主/我抢/测试模式直达 PLAY
        ├── game/
        │   └── session.go          # 手牌分配与摸牌逻辑
        ├── rules/
        │   ├── book.go             # 规则主源：牌型 key、优先级、比较元数据
        │   └── catalog.go          # 牌型规则目录、优先级与前端展示 DTO
        └── sortx/
            └── hand.go             # 自动理牌（赖子前置+排序）
├── test/
    └── backend/
        ├── fsm_test.go
        ├── rules_test.go
        ├── service_test.go
        └── sortx_test.go
```

- `tiandi-server` 负责运行可调用 API 的内存态一局服务。
- `tiandi-demo` 负责本地命令行输出状态，帮助验证状态机。
- `internal/tiandi` 采用分层：
  - `domain`（实体与规则）
  - `game`（会话数据构建）
  - `fsm`（游戏流程）
  - `demo`（服务层组装与对外状态转换）
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
    └── components/
        ├── ActionBar.tsx         # 操作按钮（叫地主/不叫/抢地主/不抢/我抢）
        ├── PlayerPanel.tsx       # 玩家展示块
        ├── CardStrip.tsx         # 手牌/底牌卡片列表
        └── BottomPanel.tsx       # 底牌展示
        └── RulesPanel.tsx        # 牌型规则展示面板
```

测试目录（与前端功能协作）在：
- `test/web/app.test.tsx`（集成式交互测试：首次加载、规则目录渲染、规则接口失败降级、测试模式 `PLAY` 渲染）

## 3. 接口清单（前后端协作入口）

| 接口 | 方法 | 请求 | 响应 | 说明 |
|---|---|---|---|---|
| `/api/game/state` | `GET` | 无 | `GameState` | 获取当前局面状态 |
| `/api/game/reset` | `POST` | 无 | `GameState` | 重开一局，并返回新状态 |
| `/api/game/action` | `POST` | `ActionRequest` | `GameState` | 提交一条动作并返回执行后的状态 |
| `/api/game/rules` | `GET` | 无 | `RulesCatalog` | 获取牌型规则目录与优先级定义 |

- `ActionRequest`（后端 `demo.Service.Apply` 输入）
  - `seat: string`（`P0/P1/P2`）
  - `kind: string`（`jiaodizhu | bujiao | qiangdizhu | buqiang | woqiang`）
- `GameState`（前端 `web/src/types.ts`）核心字段
  - `phase`, `currentActor`, `availableActions`, `message`
  - `players[]`（`seat/isLandlord/isCurrent/cards`）
  - `landlord`, `multiplier`
  - `laizi`（`tian/di`, `tianVisible/diVisible`）
  - `bottom`（`visible/count/cards`）
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

## 3.1 测试模式约定

- 后端支持通过 `TIANDI_TEST_MODE=fixed_p0_play_test` 启用测试模式。
- 测试模式下：
  - 洗牌仍随机
  - 底牌仍随机
  - 天赖/地赖仍随机
  - 状态机在发牌后直接进入 `PLAY`
  - 地主固定为 `P0`
  - `P0` 直接吃到底牌，因此手牌为 `20` 张
  - 前端通过 `GameState.testMode` 显式展示“当前为测试模式 / 固定地主为 P0 / 直接进入 PLAY”

## 4. 协作流程（请求流）

### 4.1 页面初始化
1. 前端 `App.tsx` `useEffect` 中并行调用 `fetchGameState()` 与 `fetchRulesCatalog()`。
2. `api/game.ts` 发起 `GET /api/game/state`；`api/rules.ts` 发起 `GET /api/game/rules`。
3. 后端 `tiandi-server/main.go` 分别路由到 `handleState` 与 `handleRules`。
4. `handleState` 返回当前局面；`handleRules` 返回静态规则目录 `rules.Catalog`。
5. 前端根据局面渲染 Header、testMode 提示区、ActionBar、玩家区与底牌区；规则目录成功时渲染 `RulesPanel`，失败时只在规则区域降级提示，不阻塞主界面。

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
2. 前端调用 `applyGameAction({ seat: data.currentActor, kind })`。
3. 后端 `handleAction` 校验请求 JSON，构造 `demo.ActionRequest`。
4. `service.Apply` 做：
   - `ParseSeat(req.Seat)`
   - `machine.Apply(fsm.Action{Seat, Kind})`
   - `buildState(snapshot)`
5. 状态返回后前端覆盖本地 `ready` 数据并重渲染。

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
- 因此 `CardStrip` 只做展示组件，不再包含排序、规则判断。
- 赖子标注也来源于后端标注后的 `CardView.isLaizi`。
- 新增的牌型规则目录同样由后端统一定义，前端只负责渲染，不在前端重复维护炸弹优先级和牌型名称。
- `testMode` 同样以后端为准，前端不从 `phase === PLAY && landlord === P0` 反推是否为测试模式。

## 6. 测试策略

- 后端规则模块单测：
  - `backend/test/backend/rules_test.go`
  - 验证必需牌型是否全部存在。
  - 验证炸弹优先级顺序是否与规则文档一致。
  - 验证顺子最高点 `sequenceHigh` 是否为 `A`。
  - 验证比较说明与赖子多解说明字段存在。
- 后端状态机 / service 测试：
  - `backend/test/backend/fsm_test.go`
  - `backend/test/backend/service_test.go`
  - 验证测试模式直接进入 `PLAY`。
  - 验证地主固定为 `P0`，且 `P0` 手牌为 `20` 张。
  - 验证 `testMode` 元数据输出完整。
- 后端 HTTP 接口测试：
  - `backend/cmd/tiandi-server/main_test.go`
  - 验证 `GET /api/game/rules` 返回成功且 JSON 结构可解析。
  - 验证 `POST /api/game/rules` 被正确拒绝。
  - 验证测试模式下 `GET /api/game/state` 返回 `PLAY + P0 + testMode`。
- 前端模块测试：
  - `test/web/app.test.tsx`
  - 验证游戏状态与规则目录并行加载后的正常渲染。
  - 验证规则目录中的代表性牌型可见。
  - 验证规则接口失败时主界面仍可继续使用。
  - 验证动作提交仍按既有 `/api/game/action` 路径工作。
  - 验证测试模式下显示 banner、`P0` 为地主、`ActionBar` 无动作态。
- 浏览器联调验证：
  - 使用 Playwright interactive 流程验证真实页面。
  - 验证首屏为 `PLAY`、地主 `P0`、测试模式提示可见、底牌公开、规则面板正常。
  - 保留截图证据。
- 集成验证：
  - `go test ./...`
  - `npm test`
  - `npm run build`

## 7. 运行与联调环境说明

- 前端默认通过 `import.meta.env.VITE_API_BASE_URL` 组装接口。
- 开发模式下 `web/vite.config.ts` 已配置：
  - Vite 开发服务器端口：`5173`
  - 代理：`/api -> http://localhost:8080`
- 后端监听端口：`:8080`。
- 推荐联调方式（本仓库脚本）：
  - 后端：`TIANDI_TEST_MODE=fixed_p0_play_test npm run backend:dev`
  - 前端：`dev`
