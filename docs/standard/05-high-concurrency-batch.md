---
title: 高并发批量指南 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 05 · 高并发批量指南

## Spawn 容量

spawn binary 是 client-side 工具,主要瓶颈在:
- LLM API rate(由 provider 决定)
- K8s API rate(deploy 时)
- skill marketplace API(下载 skill artifact)

## 批量 deploy

```bash
spawn deploy agents/*.yml --concurrency 5
```

5 并发 deploy。fail-fast 默认;`--continue-on-error` 跳过失败继续。

## Rate limit

- skill marketplace:per-IP 100 rpm
- K8s API:per-cluster default(DOKS quota)
- OpenRouter:per-sub-key

## Hot path

`spawn compose` 是 CPU-bound(YAML parse + DSL eval + skill bundle)。大 YAML(>1MB)可能慢 — 拆 stage。

`spawn deploy` 是 IO-bound,K8s API 单 cluster 一般 ok。
