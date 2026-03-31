# tiandi_play 实现设计稿

## 1. 后端设计

### 1.1 新增模块建议

建议新增目录：

```text
backend/internal/tiandi/play/
├── types.go
├── analyze.go
├── compare.go
├── laizi_resolve.go
├── engine.go
├── turn.go
└── result.go
```

职责建议：

- `types.go`
  - `PlayIntent`
  - `ResolvedHand`
  - `ResolutionCandidate`
  - `CurrentTrick`
  - `PlayError`
- `analyze.go`
  - 非赖子/含赖子判型入口
- `compare.go`
  - 主干比较
  - 炸弹 / 王炸跨型压制
  - 三带类与四带类按主牌点数直接互比
- `laizi_resolve.go`
  - 赖子替代枚举
  - 候选排序
  - 炸弹优先
- `engine.go`
  - 对外统一入口
  - `AnalyzeSelection`
  - `ApplyPlay`
  - `CanPass`
- `turn.go`
  - 当前轮推进
  - `pass` 链路
  - 重置首出权
- `result.go`
  - 胜负判定
  - 结束态冻结

### 1.2 对现有模块的改动建议

- `backend/internal/tiandi/fsm/machine.go`
  - 保留现有叫抢地主流程
  - `PLAY` 阶段改为可接收 `play/pass`
  - `Snapshot` 新增出牌阶段状态字段
- `backend/internal/tiandi/demo/service.go`
  - `ActionRequest` 新增 `cards`、`resolutionId`
  - `StateResponse` 新增 `currentTrick`、`winner`、`playError`、`resolvedHand`、`resolutionCandidates`
  - `availableActions()` 在 `PLAY` 阶段依赖桌面状态输出 `play/pass`
- `backend/cmd/tiandi-server/main.go`
  - 路由不变
  - `POST /api/game/action` 保持统一入口

### 1.3 推荐状态模型

在 `fsm.Snapshot` 中新增：

- `LeadingSeat`
- `LastPlaySeat`
- `HasCurrentTrick`
- `CurrentTrickCards []domain.Card`
- `CurrentTrickKind string`
- `CurrentTrickMainRank string`
- `PassCount int`
- `Winner domain.Seat`
- `HasWinner bool`
- `LastResolvedHand *ResolvedHand`
- `LastCandidates []ResolutionCandidate`
- `LastPlayError string`

说明：

- `LeadingSeat`
  - 当前轮首出权所属玩家
- `LastPlaySeat`
  - 最近一次成功出牌的人
- `PassCount`
  - 从最近一次成功出牌后累计的连续 `pass` 数
- `HasCurrentTrick`
  - 桌面是否已有主牌

### 1.4 DTO 草案

`ActionRequest`

```json
{
  "seat": "P0",
  "kind": "play",
  "cards": ["heart-A", "club-A", "spade-A"],
  "resolutionId": "candidate-1"
}
```

`StateResponse` 新增字段建议

```json
{
  "currentTrick": {
    "seat": "P0",
    "label": "三张",
    "kind": "triple",
    "cards": ["heart-A", "club-A", "spade-A"],
    "mainRank": "A",
    "isBomb": false
  },
  "resolvedHand": {
    "kind": "triple",
    "label": "三张",
    "pattern": "AAA",
    "cards": ["heart-A", "club-A", "spade-A"],
    "resolvedCards": ["heart-A", "club-A", "spade-A"],
    "mainRank": "A",
    "usesLaizi": false,
    "isBomb": false
  },
  "resolutionCandidates": [],
  "playError": "",
  "winner": "P0"
}
```

### 1.5 回合流转

`PLAY` 阶段建议流程：

1. 当前行动位提交 `play`。
2. 后端校验：
   - 是否轮到该玩家
   - 所选牌是否都在手牌中
   - 能否解析为合法牌型
   - 如桌面已有主牌，是否能压过
   - 若为三带类 / 四带类互比，只按主牌点数比较，不引入主干张数权重
3. 校验成功后：
   - 从手牌中移除
   - 更新 `currentTrick`
   - `LastPlaySeat = 当前玩家`
   - `PassCount = 0`
   - 行动位切到下一家
4. 若提交 `pass`：
   - 必须 `HasCurrentTrick = true`
   - `PassCount += 1`
   - 若 `PassCount < 2`，行动位切到下一家
   - 若 `PassCount == 2`：
     - 清空 `currentTrick`
     - `CurrentActor = LastPlaySeat`
     - `LeadingSeat = LastPlaySeat`
     - `PassCount = 0`
5. 若当前玩家手牌清空：
   - `HasWinner = true`
   - `Winner = 当前玩家`
   - `availableActions = []`

### 1.6 候选解排序策略

建议统一排序键：

1. 是否王炸
2. 是否 5+ 炸
3. 是否纯赖子炸
4. 是否实心四炸
5. 是否赖子替代炸
6. 主干长度 / 主牌大小
7. 赖子消耗更少优先

同时：

- `resolutionCandidates` 返回全部候选
- `resolvedHand` 取排序第一项

## 2. 前端设计

### 2.1 组件改动

- `web/src/types.ts`
  - 扩展 `DemoCard`，增加稳定 `id`
  - 扩展 `GameState`
  - 新增：
    - `CurrentTrickView`
    - `ResolvedHandView`
    - `ResolutionCandidateView`
    - `WinnerView`
- `web/src/api/game.ts`
  - `ActionPayload` 支持 `cards?: string[]`
  - `resolutionId?: string`
- `web/src/components/CardStrip.tsx`
  - 支持：
    - `selectable`
    - `selectedIds`
    - `disabled`
    - `onToggle`
- `web/src/components/PlayerPanel.tsx`
  - 只允许当前行动位手牌可选
- `web/src/components/ActionBar.tsx`
  - 在 `PLAY` 阶段显示：
    - `出牌`
    - `不出`
    - `清空选择`
- 新增组件建议
  - `CurrentTrickPanel.tsx`
  - `ResolutionPanel.tsx`
  - `WinnerBanner.tsx`
  - `PlayErrorBanner.tsx`

### 2.2 状态管理

`App.tsx` 本地新增状态：

- `selectedCardIds: string[]`
- `lastLocalError: string | null`

设计原则：

- 选牌只存在于本地
- 出牌结果以后端返回为准
- 每次局面刷新后：
  - 清空已选牌
  - 清空本地错误

### 2.3 页面交互流程

1. 页面进入 `PLAY`
2. 当前行动位玩家点击自己手牌
3. `selectedCardIds` 更新
4. 点击 `出牌`
5. 前端调用：

```ts
applyGameAction({
  seat: data.currentActor,
  kind: "play",
  cards: selectedCardIds,
})
```

6. 后端返回新状态
7. 前端重渲染：
   - 当前桌面牌
   - 手牌数量
   - 下一行动位
   - 候选解 / 解析说明
8. 若点击 `不出`
   - 仅在 `currentTrick` 存在时调用 `kind = "pass"`

### 2.4 展示原则

- 主牌区展示原牌标签，不替换成解释牌。
- 单出赖子时，桌面牌和手牌都显示原牌本体。
- 若有 `resolvedCards`，仅在解析说明区域展示。
- 若 `winner` 存在：
  - 展示胜者 banner
  - 动作条进入只读态

### 2.5 错误与边界态

- 非法出牌
  - 展示后端 `playError`
  - 不清空当前手牌 UI
- 首出时不允许 `pass`
  - 不渲染 `不出`
- 非当前行动位
  - 手牌不可选
- 已结束
  - 禁止继续选牌与提交动作
- 规则候选多解
  - 主视图显示 `resolvedHand`
  - 侧边显示 `resolutionCandidates`

## 3. 测试设计

### 3.1 后端测试矩阵

- `play/analyze_test.go`
  - 表驱动覆盖全部牌型
- `play/compare_test.go`
  - 同型比较
  - 炸弹跨型压制
  - 四带两单与三带二按主干比较
- `play/laizi_resolve_test.go`
  - 多解返回
  - 炸弹优先
- `play/turn_test.go`
  - 首出
  - 合法跟牌
  - 非法跟牌
  - 首出阶段非法 `pass`
  - 两家 `pass` 后重置首出
  - 胜负结束
- `demo/service_test.go`
  - DTO 输出完整

### 3.2 fixture 组织方式

建议新增：

```text
backend/test/backend/fixtures/
├── hands.go
├── tricks.go
└── laizi_cases.go
```

要求：

- 每个 fixture 带注释说明来源规则
- 命名直接反映牌例
- 同一牌例可同时供 analyze / compare / turn 测试复用

### 3.3 前端测试矩阵

- `CardStrip` 选中 / 取消选中
- `ActionBar` 在不同阶段的按钮可见性
- `App` 发送 `play` payload
- `App` 发送 `pass` payload
- `App` 渲染 `playError`
- `App` 渲染 `currentTrick`
- `App` 渲染 `winner`
- 单出赖子时仍显示原牌本体

### 3.4 Playwright interactive 场景

1. 初始 `PLAY` 场景
2. 选牌高亮场景
3. 出牌成功场景
4. 非法出牌报错场景
5. `pass` 后轮转场景
6. 两家 `pass` 后上一手赢家重新首出场景
7. 结束胜利场景

每个场景至少保留 1 张截图。

## 4. 推荐实施顺序

1. 先实现后端 `types/analyze/compare`
2. 再实现 `turn/engine`
3. 再扩展 `demo/service.go` DTO
4. 然后接前端 `types/App/CardStrip/ActionBar`
5. 最后补齐 Go test / Vitest / Playwright
