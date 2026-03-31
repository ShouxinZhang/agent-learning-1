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
        │   └── service.go          # 游戏服务层（状态/动作转换）
        ├── domain/
        │   ├── cards.go            # 牌面、花色、点数与序列化
        │   ├── laizi.go            # 天赖/地赖选择与赖子判断
        │   └── player.go           # 座位与玩家数量
        ├── fsm/
        │   └── machine.go          # 状态机：洗牌/发牌/选赖子/叫地主/抢地主/我抢
        ├── game/
        │   └── session.go          # 手牌分配与摸牌逻辑
        ├── rules/
        │   └── catalog.go          # 牌型规则目录、优先级与前端展示 DTO
        └── sortx/
            └── hand.go             # 自动理牌（赖子前置+排序）
├── test/
    └── backend/
        ├── fsm_test.go
        ├── rules_test.go
        └── sortx_test.go
```

- `tiandi-server` 负责运行可调用 API 的内存态一局服务。
- `tiandi-demo` 负责本地命令行输出状态，帮助验证状态机。
- `internal/tiandi` 采用分层：
  - `domain`（实体与规则）
  - `game`（会话数据构建）
  - `fsm`（游戏流程）
  - `demo`（服务层组装与对外状态转换）
  - `rules`（牌型规则目录与优先级元数据）
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
- `test/web/app.test.tsx`（集成式交互测试：首次加载、规则目录渲染、规则接口失败降级、动作提交流程）

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
- `RulesCatalog`（前端 `web/src/types.ts`）核心字段
  - `version`
  - `rankOrder`
  - `sequenceHigh`（当前为 `A`，表示顺子最高到 `A`）
  - `notes[]`
  - `sections[]`：按“基础牌型 / 组合牌型 / 连续牌型 / 飞机扩展 / 炸弹与特殊压制”分组
  - `bombPriority[]`
  - `handPriority[]`

## 4. 协作流程（请求流）

### 4.1 页面初始化
1. 前端 `App.tsx` `useEffect` 中并行调用 `fetchGameState()` 与 `fetchRulesCatalog()`。
2. `api/game.ts` 发起 `GET /api/game/state`；`api/rules.ts` 发起 `GET /api/game/rules`。
3. 后端 `tiandi-server/main.go` 分别路由到 `handleState` 与 `handleRules`。
4. `handleState` 返回当前局面；`handleRules` 返回静态规则目录 `rules.Catalog`。
5. 前端根据局面渲染 Header、ActionBar、玩家区与底牌区；规则目录成功时渲染 `RulesPanel`，失败时只在规则区域降级提示，不阻塞主界面。

### 4.2 刷新局面 / 新开局
1. 用户点击“新开一局”。
2. 前端调用 `resetGame()` -> `POST /api/game/reset`。
3. 后端 `handleReset -> service.Reset()`。
4. `service.Reset()` 重建 `fsm.Machine` 并 `Start()`。
5. `buildState(snapshot)` 返回标准化状态给前端。

### 4.3 玩家动作提交
1. 用户点击合法动作按钮（由 `data.availableActions` 动态渲染）。
2. 前端调用 `applyGameAction({ seat: data.currentActor, kind })`。
3. 后端 `handleAction` 校验请求 JSON，构造 `demo.ActionRequest`。
4. `service.Apply` 做：
   - `ParseSeat(req.Seat)`
   - `machine.Apply(fsm.Action{Seat, Kind})`
   - `buildState(snapshot)`
5. 状态返回后前端覆盖本地 `ready` 数据并重渲染。

## 5. 关键数据流职责边界（为什么前端不排序）

- 服务器侧在 `service.buildState` 中调用 `sortx.SortedHand` 做排序，前端直接按收到顺序渲染。
- 因此 `CardStrip` 只做展示组件，不再包含排序、规则判断。
- 赖子标注也来源于后端标注后的 `CardView.isLaizi`。
- 新增的牌型规则目录同样由后端统一定义，前端只负责渲染，不在前端重复维护炸弹优先级和牌型名称。

## 6. 测试策略

- 后端规则模块单测：
  - `backend/test/backend/rules_test.go`
  - 验证必需牌型是否全部存在。
  - 验证炸弹优先级顺序是否与规则文档一致。
  - 验证顺子最高点 `sequenceHigh` 是否为 `A`。
- 后端 HTTP 接口测试：
  - `backend/cmd/tiandi-server/main_test.go`
  - 验证 `GET /api/game/rules` 返回成功且 JSON 结构可解析。
  - 验证 `POST /api/game/rules` 被正确拒绝。
- 前端模块测试：
  - `test/web/app.test.tsx`
  - 验证游戏状态与规则目录并行加载后的正常渲染。
  - 验证规则目录中的代表性牌型可见。
  - 验证规则接口失败时主界面仍可继续使用。
  - 验证动作提交仍按既有 `/api/game/action` 路径工作。
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
  - 后端：`backend:dev`
  - 前端：`dev`
