---
title: Skill marketplace / plugin ecosystem(spawn 扩展)
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: true
---

# 06a · Skill marketplace / plugin ecosystem

> 06 dev integration 的姐妹文档。**只**讲 spawn skill 系统:发现 / 安装 / 写自己 / 发布。

## Skill 是什么

Skill = agent 可调用的能力封装。一个 skill = 一个目录 + `SKILL.md`(declarative manifest)+ 实现(shell / Go / Python / WASM)。

例:`github` skill 让 agent 通过 `gh` CLI 跑 GH operation;`firecrawl` skill 让 agent 抓网页。

## 发现

```bash
spawn skill list                        # 看本地 + marketplace
spawn skill search "<keyword>"
spawn skill info <skill-id>             # 看 SKILL.md
```

Marketplace:`spawn.gridltd.com/marketplace`(若 host 自家)或 GH topic `spawn-skill`。

## 安装

```bash
spawn skill install <skill-id>          # 从 marketplace
spawn skill install github.com/<owner>/<repo>   # 从 GH
spawn skill install ./local-dir         # 本地 dev
```

安装位置:`~/.spawn/skills/<id>/`。

## SKILL.md 格式

```yaml
---
id: my-skill
version: 1.0.0
description: ...
inputs: [text, image]
outputs: [text, json]
runtime: shell        # shell / go / python / wasm
entry: main.sh
permissions:
  - network
  - filesystem:rw:/tmp
---

# My Skill

操作说明,agent 在 prompt 里看 SKILL.md 学怎么调。
```

## 写一个 skill

```bash
spawn skill init my-skill
cd my-skill/
# 编辑 main.sh + SKILL.md
spawn skill test                        # local 测
spawn skill publish                     # 发到 marketplace
```

## 发布到 marketplace

需要 GH App 装好 + maintainer approve。`spawn skill publish` 触发 PR 到 `marketplace-skills` repo。

## 安全 / permissions

Skill 默认 sandboxed:
- 默认无网络
- 默认无文件读写(只 `/tmp`)
- 必须在 SKILL.md `permissions` 显式声明

agent 加载 skill 时,system_prompt 自动列出 skill 权限边界。

## 治理

- 第三方 skill 安装前必须 audit(GH App 自动 license / CVE scan)
- platform 维护 "trusted skills" whitelist
- 撤销:remove `spawn.gridltd.com/marketplace/<id>` 或 yank tag

## 历史

详 [`docs/spawn-integration-retrospective.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-integration-retrospective.md)。

## 相关链接

- mission-control `/spawn` panel 看 active spawn child
- AgentTeam [`02-power-user.md`](https://github.com/OpenAgentSystem/AgentTeam/blob/main/docs/standard/02-power-user.md) "skill" 节
