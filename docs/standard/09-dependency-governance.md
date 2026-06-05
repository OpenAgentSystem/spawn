---
title: 第三方依赖治理 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 09 · 第三方依赖治理

## 依赖清单

| 类别 | 来源 | scan |
|---|---|---|
| Go modules | `go.mod` + `go.sum` | per PR Dependabot + `golangci-lint` |
| Docker base | `Dockerfile` | 每月 EOL check |
| GH Actions | `.github/workflows/*.yml` | per PR SHA Pin + Dependabot |
| Skill marketplace 第三方 skill | curated allowlist | 每月 audit |

## License

允许:MIT / BSD / Apache-2.0 / ISC
禁止:GPL-3 / AGPL / SSPL

CI gate `license-check`(待加)`go-licenses check ./...`。

## CVE 监控

- Dependabot security alerts auto-PR
- `govulncheck ./...` per PR
- `trivy image` per Docker build

每周一 batch merge。月度 review。即时 0-day P1。

## 更新节奏

| 类型 | 节奏 |
|---|---|
| Security patch | 24h |
| Minor Go module | 周 batch |
| Major Go module / Go runtime | 季度 |
| Skill DSL spec | 跟 spawn major version |

## EOL

| 依赖 | EOL | 替换 |
|---|---|---|
| Go 1.22 | 2026 Q4 | Go 1.24 |
| ... | | |

## 供应链

- GH Actions SHA-pin 强制
- Docker image 必须 Kyverno 白名单(infra)
- `go mod download` 必须 `go.sum` commit
- `go install` 第三方包必须 review

## Skill marketplace 治理

详 [06a-skill-marketplace.md](06a-skill-marketplace.md)。
