---
title: 进阶用户技巧 — spawn
owner: KYBvWHxW
last_reviewed: 2026-06-05
next_review_due: 2026-09-03
feishu_node_id: ""
project_specific: false
---

# 02 · 进阶用户技巧

## DSL 高阶

### Multi-stage agent compose

```yaml
agent:
  id: my-pipeline
  stages:
    - id: collect
      skills: [shell, github]
      output: collected.json
    - id: analyze
      input: collected.json
      model: openrouter/anthropic/claude-opus-4
```

### Skill 局部覆盖

```yaml
skills:
  - github
  - github_overrides:
      timeout_seconds: 30
```

### Resource limits

```yaml
runtime:
  cpu_limit: 500m
  memory_limit: 1Gi
  timeout_minutes: 30
```

## CLI shortcuts

| flag | 用 |
|---|---|
| `-v` | verbose log |
| `--dry-run` | compose 但不部署 |
| `--target k8s|local|docker` | 选 deploy target |
| `--watch` | 文件变 auto-recompose |

## CI 集成

GH Action `OpenAgentSystem/spawn-action@<sha>`:

```yaml
- uses: OpenAgentSystem/spawn-action@<sha>
  with:
    file: agents/my-agent.yml
    target: k8s
```

## 与本项目特有扩展

- Skill marketplace → [06a-skill-marketplace.md](06a-skill-marketplace.md)
