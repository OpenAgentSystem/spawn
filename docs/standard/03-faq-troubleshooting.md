---
title: FAQ & 故障排查 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 03 · FAQ & 故障排查

## Top FAQ

### Q1. `spawn compose` 报错 `unknown skill: <X>`

skill 不在 marketplace 或本地仓库。
- `spawn skill list` 看可用
- 本地 dev skill 路径加到 `spawn.yml` `local_skill_paths`

### Q2. agent 起来但 `/readyz` 不通

- model API key 错(`OPENROUTER_API_KEY` 没注入)
- skill bootstrap fail(看 spawn log `skill init failed`)
- runtime resource quota 不够(K8s `kubectl describe pod`)

### Q3. K8s deploy 失败

`spawn deploy` 失败常见:
- kubeconfig 错 → `kubectl config current-context`
- ns 不存在 → spawn 自动创建?(check spec)
- image registry 没白名单 → 看 infra [`09a-base-image-governance.md`](https://github.com/OpenAgentSystem/infra/blob/main/docs/standard/09a-base-image-governance.md)

### Q4. spawn binary 跟最新 DSL 不兼容

binary release lag — `spawn version` 看 + `brew upgrade openagentsystem/tap/spawn`

## 错误信息

| log | 处置 |
|---|---|
| `unknown skill: <X>` | skill marketplace lookup / 本地路径 |
| `OPENROUTER_API_KEY missing` | 注入 env |
| `kubeconfig context not found` | `doctl kubernetes cluster kubeconfig save` |
| `image pull failed: not in allowlist` | infra 加 Kyverno 白名单 |
| `compose timeout` | DSL 复杂或 skill 慢 — `--timeout` 调长 |

## 哪里求助

- bug → spawn GH issue + `bug` label
- 紧急 → Lark "spawn 报障" 群
