---
title: 物料成本对账 FinOps — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 07 · 物料成本对账 FinOps

## spawn 自己运营成本

| 项 | 月预算(USD) | 月实际 |
|---|---|---|
| GH Actions runner-minutes(build / test) | 30 | TBD |
| ghcr.io 镜像存储 | 20 | TBD |
| brew tap(免费) | 0 | 0 |
| Marketplace API host(若有)| 50 | TBD |
| **Subtotal spawn** | **100** | **TBD** |

整体 GridLtd FinOps → AgentTeam [`07-finops-reconciliation.md`](https://github.com/OpenAgentSystem/AgentTeam/blob/main/docs/standard/07-finops-reconciliation.md)。

## spawn 触发的下游 cost

spawn 起 agent → 跑 LLM call → 上游 OpenRouter spend。

跨 OAS aggregate → mission-control [`07a-agent-cost-attribution.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/standard/07a-agent-cost-attribution.md)。

## 关键 guardrail

[`docs/spawn-cost-guardrails.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-cost-guardrails.md):
- 默认 `SPAWN_AUTO_CLEANUP_MINUTES=240`
- per-spawn child 限额
- ASMA(自动 spawn monitor alert)

## 对账 cadence

跟 GridLtd 月度 cadence 同步,详 AgentTeam 文档。

## 链接

- mission-control [`07a-agent-cost-attribution.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/standard/07a-agent-cost-attribution.md)
- mission-control [`docs/spawn-cost-guardrails.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-cost-guardrails.md)
