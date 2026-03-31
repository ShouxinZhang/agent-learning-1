# 天地癞子

一个用于演示“天地癞子斗地主”流程、规则目录和前后端联调的练习仓库。

- 后端：Go，提供内存态对局服务与规则接口
- 前端：React + Vite，提供简单的牌桌界面与规则展示
- 测试：Go test + Vitest

## 快速开始

```bash
./restart.sh
```

## 目录说明

```text
backend/   Go 后端与核心规则实现
web/       React 前端
test/      前端测试配置与测试文件
docs/      规则说明、接口说明与任务文档
```

后端入口：`backend/cmd/tiandi-server/main.go`

前端入口：`web/src/App.tsx`

## 参考文档

- `docs/tiandi_arch_and_api.md`：架构与接口说明
- `docs/tiandi_hands_rules.md`：牌型规则
- `docs/tiandi_laizi_rules.md`：天地癞子规则
- `docs/tiandi_web_task.md`：前端任务说明

## 环境说明

- Node.js：用于前端开发、构建与测试
- Go 1.23.4：用于后端服务与测试
