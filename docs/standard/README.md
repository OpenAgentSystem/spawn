# 9 大标准文档(spawn · 实时维护)

> RFC: [OpenAgentSystem/AgentTeam#4925](https://github.com/OpenAgentSystem/AgentTeam/issues/4925)
> Pilot: [OpenAgentSystem/AgentTeam#4926](https://github.com/OpenAgentSystem/AgentTeam/pull/4926)
> Cohort: [infra#637](https://github.com/OpenAgentSystem/infra/pull/637) · [gateway#665](https://github.com/OpenAgentSystem/gateway/pull/665) · [mission-control#452](https://github.com/OpenAgentSystem/mission-control/pull/452)

spawn = "Compose AI agents as code. One source, run anywhere." Go-based agent compose 工具。9 文档 + 1 spawn-specific 扩展(共 11)。

## 文档清单

| # | 文件 | 标题 |
|---|---|---|
| 01 | [`01-user-guide.md`](01-user-guide.md) | 用户使用指南(`spawn compose` / `spawn run`) |
| 02 | [`02-power-user.md`](02-power-user.md) | 进阶用户技巧 |
| 03 | [`03-faq-troubleshooting.md`](03-faq-troubleshooting.md) | FAQ & 故障排查 |
| 04 | [`04-ops-runbook.md`](04-ops-runbook.md) | 运维管理员手册(spawn server) |
| 05 | [`05-high-concurrency-batch.md`](05-high-concurrency-batch.md) | 高并发批量指南 |
| 06 | [`06-developer-integration-release.md`](06-developer-integration-release.md) | 开发者集成 & 发版 |
| 06a | [`06a-skill-marketplace.md`](06a-skill-marketplace.md) | **扩展** Skill marketplace / plugin ecosystem |
| 07 | [`07-finops-reconciliation.md`](07-finops-reconciliation.md) | 物料成本对账 FinOps |
| 08 | [`08-sales-playbook.md`](08-sales-playbook.md) | 销售作战手册 |
| 09 | [`09-dependency-governance.md`](09-dependency-governance.md) | 第三方依赖治理 |

## 实时维护协议

详 [AgentTeam pilot README](https://github.com/OpenAgentSystem/AgentTeam/blob/main/docs/standard/README.md)。要点:

- frontmatter 必填(owner / last_reviewed / next_review_due / feishu_node_id / project_specific)
- freshness gate >90 天 warn / >180 天 fail
- spawn compose DSL 改动同 PR 必须更新 `01-user-guide.md` 跟 `06a-skill-marketplace.md`

## 飞书同步

`.github/workflows/sync-docs-to-feishu.yml` push to main 触发(spawn 无 develop)。

## branch model

spawn 仓 `main` 唯一长期分支。feature/chore branch from main + PR target main。
