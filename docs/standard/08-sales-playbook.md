---
title: 销售作战手册 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 08 · 销售作战手册

> spawn 是 OAS 平台的 agent compose 工具,对外卖点是"compose agents as code"开发体验。

## 对外卖点

| 卖点 | 一句话 |
|---|---|
| 单一 YAML 定义,多端部署 | "K8s / Docker / local 一份 manifest,5 行 YAML 起 agent" |
| Skill marketplace 即拉即用 | "20+ pre-built skills(github / shell / firecrawl / mermaid / etc),即装即用" |
| Sandboxed | "skill 默认无网络/文件权限,显式声明,合规友好" |
| GitOps native | "YAML in GH,PR review,CI deploy — 跟普通代码同 workflow" |
| Cost guardrail | "自动 cleanup + 限额 + alert,不怕跑飞" |

## ICP

- 想 ship agent system 但缺 ops 团队的 SaaS 公司
- 大企业 platform team:想给内部 dev 一个统一的 agent compose 工具

## Demo(30 分钟)

1. 5 分钟:`brew install spawn` + 起第一个 agent
2. 10 分钟:加 skill(github + firecrawl)+ K8s deploy
3. 5 分钟:Spawn child agent + cleanup
4. 10 分钟:GitOps workflow + skill marketplace

## 反对意见

| 客户说 | 处置 |
|---|---|
| "用 LangChain 就行" | LangChain 是库,spawn 是 platform-level compose + 部署。补不是替 |
| "自家写 YAML schema 就行" | spawn 已 ship DSL + multi-target deploy + skill ecosystem,自己造 3 月 |
| "想 cloud 兼容" | spawn 输出兼容 K8s / Docker / local;Cloud Run / Lambda support 在 roadmap |

## Pricing

> 必查 [pricing@gridltd.com](mailto:pricing@gridltd.com)。

spawn binary 跟 OAS 平台 license 卖,部分 OSS 模块免费(基本 CLI + 一些 skill)。Enterprise 加 skill marketplace + spawn server。

详 AgentTeam [`08-sales-playbook.md`](https://github.com/OpenAgentSystem/AgentTeam/blob/main/docs/standard/08-sales-playbook.md)。
