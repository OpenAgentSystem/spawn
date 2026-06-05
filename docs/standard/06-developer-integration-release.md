---
title: 开发者集成 & 发版说明 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 06 · 开发者集成 & 发版说明

## 集成方式

### 1. CLI(默认)

```bash
brew install openagentsystem/tap/spawn
spawn run agent.yml
```

### 2. Go SDK

```go
import "github.com/OpenAgentSystem/spawn/sdk"

func main() {
    a, _ := sdk.Compose("agent.yml")
    a.Run()
}
```

### 3. GH Action

```yaml
- uses: OpenAgentSystem/spawn-action@<40-char-sha>
  with:
    file: agents/foo.yml
    target: k8s
```

## DSL 文档

详 `docs/dsl-reference.md`(待添加)+ main [`README.md`](../../README.md) 跟 [`AGENTS.md`](../../AGENTS.md)。

## 发版

每 release 自动:
1. tag `vX.Y.Z` → GH Action `release.yml`
2. multi-arch binary build (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
3. release artifact 上 GH Releases
4. push 到 `homebrew-tap` repo
5. push docker image 到 `ghcr.io/openagentsystem/spawn:<version>`

## 版本

| 版本 | 日期 | 兼容 DSL |
|---|---|---|
| latest | TBD | DSL v1.x |
| ... | | |

## SDK 兼容

Go SDK 跟 spawn binary 配套发布,版本同步。

## 集成 Skill

详 [06a-skill-marketplace.md](06a-skill-marketplace.md)。

## CODEOWNERS

paul + spawn maintainer 团队。
