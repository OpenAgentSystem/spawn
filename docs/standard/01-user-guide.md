---
title: 用户使用指南 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 01 · 用户使用指南

## 这是什么 / 给谁看

spawn = "Compose AI agents as code. One source, run anywhere."

Go binary 把 agent 定义(persona / model / skills / tools)编译成可在多端跑的 artifact:K8s pod / docker container / local process。

给 **OAS 平台团队 + 自家 agent 系统的 builder** 用。

## 5 分钟上手

```bash
# 装 binary
brew install openagentsystem/tap/spawn

# 起一个 agent
cat > my-agent.yml <<YAML
agent:
  id: my-helper
  model: openrouter/anthropic/claude-haiku-4.5
  skills: [shell, github]
  system_prompt: |
    You are a helper.
YAML

spawn run my-agent.yml
# → 起 local process, listening :18789
```

## 核心命令

| cmd | 用途 |
|---|---|
| `spawn compose <file>` | 编译 agent yml → artifact |
| `spawn run <file>` | local 跑 agent |
| `spawn deploy <file>` | 推到 K8s / DOKS |
| `spawn list` | 当前 fleet 看 running agent |
| `spawn logs <agent-id>` | 看 agent 日志 |
| `spawn stop <agent-id>` | 停 |
| `spawn skill list` | 看 skill marketplace |

详细 → `spawn --help` + 主 [`README.md`](../../README.md)。

## 哪里求助

- 集成 → [06-developer-integration-release.md](06-developer-integration-release.md)
- Skill 写 / 找 → [06a-skill-marketplace.md](06a-skill-marketplace.md)
- 报障 → Lark "spawn 报障" 群
