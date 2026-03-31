# tiandi_bash_skill_llm 完整测试计划

## 0. 当前状态

- [x] Step 1：梳理现有前后端、测试模式与文档现状
- [x] Step 2：制定 Plan，明确自然语言内容、插入点、串并行任务与 test plan
- [x] Step 3：完成后端 bash 直玩能力与 LLM 测试接入
- [x] Step 4：完成前端调试展示与真实页面验证链路
- [x] Step 5：完成模块化测试、集成测试与截图证据沉淀
- [x] Step 6：汇总 subagent 交付并更新 `docs/tiandi_arch_and_api.md`

## 1. User 原始提示

> 仓库目前前端测试比较简陋
> 有鉴于此，我决定开发一套后端bash里，直接玩游戏的skills
> 然后接入这个免费的LLM, 进行完整测试
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

- 后端已经具备：
  - Go 内存态对局服务
  - `PLAY` 阶段的状态表示、动作提交与规则目录接口
  - 测试模式 `fixed_p0_play_test`
- 前端已经具备：
  - React + Vite 牌桌页面
  - 规则目录展示
  - `PLAY` 阶段选牌与出牌基础交互
- 当前不足：
  - 没有给终端/脚本直接消费的“bash skill 层”
  - 没有 LLM 可直接使用的 prompt 规范、动作协议和回合驱动器
  - 前端测试仍以 Vitest 为主，缺少一条“后端状态 -> LLM 决策 -> 前端真实页面截图验证”的完整链路

### 2.2 本轮真实目标

- 补一套后端 bash 直玩能力，让人或脚本不经过浏览器也能直接读取状态、执行动作、跑整局测试。
- 接入基于 OpenRouter 的免费模型测试链路。
  - 本计划默认按用户截图中的模型 ID：`qwen/qwen3.6-plus-preview:free`
  - 鉴权不写入仓库，运行时读取 `OPENROUTER_API_KEY`
- 固化给 LLM 的自然语言内容，避免每次测试都手写 prompt。
- 增加前端调试展示，使同一份后端 prompt / agent 决策能够被可视化检查。
- 用模块化测试 + 集成测试 + Playwright 真实截图收口，确保后端、前端和 agent 协作可回归。

### 2.3 关键约束

- 后端仍然是唯一规则事实源。
- bash skill 不复制规则判断，只消费后端状态并提交动作。
- LLM 只能输出约束内的结构化动作，不直接操作 HTTP 细节。
- 免费模型接入必须允许离线 mock，避免测试强依赖真实外部调用。
- 前端只展示和辅助验证，不在前端重写判型逻辑。

## 3. 自然语言完整内容

### 3.1 必须固化的 prompt 组成

- 系统提示：
  - 你的身份是“天地癞子斗地主测试代理”
  - 你的目标是从当前局面中选择一个合法动作
  - 你不能发明规则，必须遵循服务端给出的可执行动作和候选信息
- 规则摘要：
  - 当前阶段
  - 当前操作者
  - 可用动作列表
  - 当前桌面主牌与最近解析结果
  - 选牌时只允许使用当前玩家手牌 id
  - 非首出时仅当 `availableActions` 含 `pass` 才允许不出
- 输出协议：
  - 只输出 JSON
  - 字段固定为 `seat`、`kind`、`cards`、`reason`
  - `cards` 必须是牌 id 数组
  - `kind` 必须来自服务端返回的 `availableActions`
- 容错约束：
  - 若没有足够信息完成出牌，优先返回最保守的合法动作
  - 若 `play` 可用但无明确牌组策略，可选择最简单合法候选
  - 禁止输出解释性散文包裹 JSON

### 3.2 建议落地的 prompt 产物

- `system prompt`
  - 固定测试代理职责、输出格式和禁止事项
- `user prompt`
  - 由后端当前局面、可用动作、规则摘要和玩家手牌动态生成
- `few-shot examples`
  - 首出单张
  - 跟牌压制
  - 桌面已有主牌时选择 `pass`
  - 非法动作时回退到结构化重试

### 3.3 建议的结构化动作协议

```json
{
  "seat": "P0",
  "kind": "play",
  "cards": ["heart-A", "club-A"],
  "reason": "使用当前玩家手牌中的最小合法压制。"
}
```

- `seat`
  - 必须等于当前操作者
- `kind`
  - 必须属于后端返回的 `availableActions`
- `cards`
  - 仅 `kind = play` 时必填
  - 只能出现当前玩家手牌 id
- `reason`
  - 用于日志和前端调试展示

### 3.4 自然语言内容插入策略

- 后端：
  - 新增 prompt builder，把规则摘要和当前局面拼成 prompt
  - 新增 agent runner / bash 工具消费该 prompt
- 前端：
  - 增加调试面板，展示“当前 prompt 摘要 / 最近一次 agent 决策 / 最近一次执行结果”
  - 允许人工对照真实页面验证 prompt 是否和当前 UI 一致

## 4. 模块插入点

### 4.1 后端插入点

- `backend/internal/tiandi/demo/service.go`
  - 复用现有 `StateResponse` 作为 agent 状态源
- `backend/cmd/tiandi-server/main.go`
  - 增加 agent prompt / agent action 相关 HTTP 接口
- `backend/internal/tiandi/...`
  - 新增 `agent` 或 `prompt` 模块，负责：
    - prompt 生成
    - 动作 JSON 解析
    - agent trace 结构
- `scripts/` 或 `backend/cmd/...`
  - 提供 bash 直玩脚本和 OpenRouter 调用脚本

### 4.2 前端插入点

- `web/src/types.ts`
  - 增加 prompt / agent trace 调试类型
- `web/src/api/`
  - 新增 agent 调试接口调用封装
- `web/src/App.tsx`
  - 拉取并展示 agent 调试信息
- `web/src/components/`
  - 增加调试面板组件
- `test/web/`
  - 增加前端调试展示与交互测试

## 5. 执行策略

### 5.1 串行主链

- [x] S1. 锁定自然语言与动作协议
  - 目标：定义 LLM 可复用 prompt、JSON 输出和失败回退规范
  - 输出：本计划第 3 节 + 后续 prompt 模块
- [x] S2. 锁定后端与前端插入点
  - 目标：明确哪些模块产出 prompt，哪些模块消费和展示
  - 输出：本计划第 4 节
- [x] S3. 实现后端 bash 直玩与 LLM runner
  - 目标：让命令行能读取状态、驱动动作、调用 OpenRouter 并记录 trace
  - 输出：bash 工具、后端 prompt 模块、接口与测试
- [x] S4. 实现前端调试展示
  - 目标：让页面能验证 agent prompt 和决策结果
  - 输出：前端面板、接口封装、Vitest 用例
- [x] S5. 完成端到端验证
  - 目标：让后端、LLM、前端协作形成可回归路径
  - 输出：集成测试、Playwright 截图、日志说明
- [x] S6. 回写架构文档
  - 目标：将实际落地结构同步到 `docs/tiandi_arch_and_api.md`
  - 输出：更新后的架构说明

### 5.2 可并行任务

- [x] P1. 后端 agent/prompt/bash 技术方案与实现
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：补齐 prompt builder、HTTP 接口、bash runner 或 CLI
  - 输出：后端代码、脚本、单测
- [x] P2. 前端调试面板与测试
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：补齐 agent prompt / trace 展示、测试和页面可见性
  - 输出：前端代码、Vitest、Playwright 关注点
- [x] P3. 集成测试与验收
  - 模型：`gpt-5.4`，reasoning `high`
  - 目标：验证 mock LLM、真实调用入口、bash 命令与页面联调
  - 输出：测试矩阵、脚本、验收记录

### 5.3 串并行依赖

- `P1` 与 `P2` 可并行，但共享的数据契约必须先由协调者固定。
- `P3` 依赖 `P1` 和 `P2` 的接口契约落地。
- 文档回写在所有实现和验收完成后统一执行。

## 6. Subagent 分发计划

### 6.1 协调者职责

- 统一 prompt 协议、HTTP 返回结构和前端展示字段
- 审核每个 subagent 的交付是否与本计划一致
- 只允许后端作为规则与动作合法性的最终事实源
- 在本计划中实时维护勾选状态

### 6.2 任务卡

- [x] T1. 后端 Bash Skill / LLM Runner
  - 负责人：Subagent A
  - 目标：提供 bash 直玩脚本、OpenRouter 请求脚本或 CLI、后端 prompt 生成与 mock 测试
  - 输出：后端代码、脚本、单元测试、运行说明
- [x] T2. 前端 Agent Debug 面板
  - 负责人：Subagent B
  - 目标：展示 prompt 摘要、最近一次决策、执行结果和失败信息
  - 输出：React 组件、API 封装、前端测试
- [x] T3. 联调与验收
  - 负责人：Subagent C
  - 目标：补齐端到端验收，输出模块测试、集成测试和截图证据
  - 输出：测试脚本、截图说明、验收结论

## 7. Test Plan

### 7.1 后端模块测试

- [x] Prompt builder 测试
  - 验证系统提示、用户提示、动作约束和当前局面摘要完整
  - 验证当前玩家、可执行动作和手牌 id 出现在 prompt 中
- [x] 动作解析测试
  - 验证合法 JSON 动作可解析
  - 验证非法 `kind`、错误 `seat`、越权 `cards` 被拒绝
- [x] OpenRouter client/mock 测试
  - 验证请求头、模型名、消息体组装正确
  - 验证 mock 响应可稳定回放
- [x] Bash skill 测试
  - 验证状态读取、执行动作、重置、trace 记录
  - 测试用例使用详细注释和 mock 数据，保证可复用

### 7.2 前端模块测试

- [x] 调试面板渲染测试
  - 验证 prompt 摘要、动作 JSON、执行状态和错误提示可见
- [x] 交互测试
  - 验证刷新状态后调试面板更新
  - 验证 agent 决策执行后页面状态联动变化
- [x] 降级测试
  - 验证 agent 调试接口失败时，主牌桌仍可继续使用

### 7.3 集成测试

- [x] 模拟 LLM 集成
  - 使用 mock 响应驱动完整一手动作
  - 校验后端状态和前端展示一致
- [x] 真实 LLM 入口冒烟（代码路径与缺密钥降级已验证，真实联网调用待提供 `OPENROUTER_API_KEY`）
  - 仅在存在 `OPENROUTER_API_KEY` 时启用
  - 至少跑通一次 prompt -> 决策 -> action 提交
- [ ] Bash 到前端协作
  - bash 提交动作后，前端刷新应看到同步局面

### 7.4 Playwright 验收

- [x] 使用 Playwright interactive MCP 验证真实页面
  - 打开首页并确认测试模式 banner / 调试面板可见
  - 触发一次 agent 决策或展示最近一次 trace
  - 截图保存到 `artifacts/playwright/`
- [x] 图片回传验证
  - 保留至少 2 张截图：
    - 调试面板初始态
    - agent 执行后页面态

## 8. 验收标准

- bash 中可以直接读取局面并提交动作
- 存在一条可选的 OpenRouter 免费模型调用链路
- prompt、动作协议、trace 结构固定，可被测试稳定消费
- 前端能展示调试信息且不破坏现有牌桌
- 后端、前端、集成、Playwright 验收全部通过
- `docs/tiandi_arch_and_api.md` 已更新

## 9. 完成记录

- [x] 计划文档创建完成
- [x] Subagent 已分发并完成交付
- [x] 主线程完成集成与验收
- [x] 架构文档已回写
