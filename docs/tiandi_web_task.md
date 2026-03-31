# 天地癞子测试前端与自动理牌任务文档

## 1. 任务目标

本次任务分两部分：

1. 增加一个极简测试前端，使用 `React + Vite`，放在 `web/` 目录下。
2. 后端增加自动理牌模块，把每位玩家的手牌按“从左到右、从大到小”排序，且 `laizi` 自动放在最前面。
3. 前端必须能测试“出牌前”的完整流程，不是只读展示结果，而是要能操作：
   - `叫地主`
   - `不叫`
   - `抢地主`
   - `不抢`
   - `我抢`

前端仍然只覆盖“出牌前”的流程，不进入正式出牌阶段。

## 2. 前端范围

前端展示并驱动一局“准备完成后到正式出牌前”的流程。

页面要求保持极简：

1. 页面上只展示三个玩家矩形区域。
2. 每个矩形区域显示：
   - 玩家座位
   - 是否为地主
   - 该玩家当前手牌
3. 当前轮到谁操作，要明确显示。
4. 当前阶段允许的按钮，只显示合法动作。
5. 地主确定后，地主的手牌中插入底牌。
6. 顶部可补充少量辅助信息：
   - 天赖子
   - 地赖子
   - 倍数
   - 当前阶段

## 3. 前端模块树

建议采用如下结构：

```text
web/
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── styles.css
    ├── types.ts
    ├── api/
    │   └── game.ts
    └── components/
        ├── ActionBar.tsx
        ├── PlayerPanel.tsx
        ├── CardStrip.tsx
        └── BottomPanel.tsx

test/
├── package.json
├── vitest.config.ts
├── backend/
│   ├── fsm_test.go
│   └── sortx_test.go
└── web/
    └── app.test.tsx
```

## 4. 前端 display 机制

### 4.1 数据来源

前端通过一组简单接口读取并驱动后端状态：

- `GET /api/game/state`
- `POST /api/game/reset`
- `POST /api/game/action`

接口返回字段建议包括：

1. `phase`
2. `currentActor`
3. `availableActions`
4. `players`
5. `landlord`
6. `laizi`
7. `multiplier`
8. `bottom`

其中每个玩家对象始终带有当前时刻手牌，且这些手牌都由后端完成理牌。

### 4.2 展示机制

1. 页面初始化时请求 `/api/game/state`。
2. 如需重新开局，点击“新开一局”并调用 `/api/game/reset`。
3. 根据 `availableActions` 渲染当前合法动作按钮。
4. 点击动作按钮后，调用 `/api/game/action`。
5. 收到新状态后重新渲染三个 `PlayerPanel` 和底牌区。
6. `CardStrip` 只负责按数组顺序平铺显示，不在前端做排序。
7. 排序规则完全由后端负责，前端只做展示。

### 4.3 视觉限制

1. 整体样式保持测试工具级别，不做复杂皮肤。
2. 每位玩家就是一个矩形面板。
3. 牌也只做成小矩形标签，不做真实扑克牌贴图。
4. 地主面板用明显但简单的边框或底色区分。
5. `laizi` 牌用单独颜色标记，方便肉眼检查理牌是否正确。

## 5. 后端模块设计

后端补一个独立的自动理牌模块，职责只做排序，不做牌型判断。

建议增加：

```text
internal/tiandi/sortx/
└── hand.go
```

说明：

1. `sortx` 避免与标准库 `sort` 包同名冲突。
2. 提供 `SortHand(cards, laiziPair)` 一类方法。
3. 排序规则：
   - `laizi` 最前
   - 非 `laizi` 再按点数从大到小
   - 同点数时按花色做稳定次序

## 6. 后端 demo 输出机制

为了给前端测试，后端补一个极简 HTTP server 和内存态单局服务：

1. 提供 `/api/game/state`
2. 提供 `/api/game/reset`
3. 提供 `/api/game/action`
4. 服务启动时自动初始化一局
5. 每次状态输出前，对三家手牌分别调用自动理牌模块
6. 地主确定后，将底牌并入地主手牌并返回最终状态

## 7. 执行顺序

1. 先新增本任务文档。
2. 读文档并确认模块边界。
3. 实现后端 `lipai` 模块。
4. 实现 `state/reset/action` API 和 server。
5. 创建 `web/` 下的 React + Vite 极简前端。
6. 加入最小测试，并统一放进 `test/`。
7. 联调：确保前端能请求后端并把“出牌前”流程完整走通。
7. 验证：
   - `go test ./...`
   - `cd test && npm install && npm test`
   - `cd web && npm run build`

## 8. 本次实现边界

本次不做以下内容：

1. 不做正式出牌阶段的出牌交互
2. 不做复杂动画
3. 不做完整牌型分析 UI
4. 不做数据库与持久化
