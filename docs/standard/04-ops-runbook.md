---
title: 运维管理员手册 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 04 · 运维管理员手册

## 关键服务

spawn 是 CLI binary,无 server / 长 service。运维核心是:

- **release pipeline**:GH release → brew tap + multi-arch binary
- **skill marketplace**:`spawn.gridltd.com/marketplace` REST(若 host 自家 marketplace)
- **deploy target validation**:K8s / DOKS / Docker 兼容性

## Runbook 索引

- [`docs/admin-spawn-sandbox.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/admin-spawn-sandbox.md) — spawn sandbox 隔离机制
- [`docs/spawn-saas-security.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-saas-security.md) — spawn SaaS 安全
- [`docs/spawn-cost-guardrails.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-cost-guardrails.md) — cost 保护
- [`docs/spawn-integration-poc-criteria.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-integration-poc-criteria.md)
- [`docs/spawn-integration-retrospective.md`](https://github.com/OpenAgentSystem/mission-control/blob/main/docs/spawn-integration-retrospective.md)

## 常见 incident

### RB-01 brew tap 推送失败

- 看 GH Actions workflow `release.yml` log
- 通常 `homebrew-tap` repo 写权限缺(GH App token scope)

### RB-02 skill marketplace 502

`spawn.gridltd.com/marketplace` 不通:
- 看 marketplace pod(若 host 自家)
- fall back 到 GH 直接拉(skill 源也在 GH)

### RB-03 K8s deploy target 不通

参考 mission-control / infra runbook。

## SLO

- `spawn compose` p95 < 5s
- `spawn deploy` p95 < 30s
- skill marketplace API p99 < 200ms
