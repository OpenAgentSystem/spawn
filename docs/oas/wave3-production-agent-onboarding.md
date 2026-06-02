# Wave 3 production agent onboarding

Wave 3 extends the Spawn supply layer beyond the Wave 2 demo agent. The first
five production agent types are:

| Agent type | Repository | Adapter Spec | Spawn API | Legacy direct-invoke |
| --- | --- | --- | --- | --- |
| `narrator-ai` | `GridLtd-ProductDev/narrator-ai-py` | `1.0` | `/v1/spawn/dispatch` | deprecated |
| `GridFastAgent` | `OpenAgentSystem/agent-services` | `1.0` | `/v1/spawn/dispatch` | deprecated |
| `hire-app` | `OpenAgentSystem/hire-app` | `1.0` | `/v1/spawn/dispatch` | deprecated |
| `closure-bot` | `OpenAgentSystem/closure-bot` | `1.0` | `/v1/spawn/dispatch` | deprecated |
| `agency-agents` | `OpenAgentSystem/agency-agents` | `1.0` | `/v1/spawn/dispatch` | deprecated |

The canonical executable registry is `DefaultWave3ProductionAgents()` in
`packages/supplylayer`. The production fixture is
`examples/oas/wave3-production-agents`.

## Acceptance evidence

- Each agent has a production registration record.
- Each registration validates Adapter Spec `1.0`.
- Each registration points to `POST /v1/spawn/dispatch`.
- Each registration marks legacy direct-invoke deprecated.
- Each agent can build an allowed `LaunchRequest` through `DefaultWave3Registry`.

Run:

```bash
cd packages/supplylayer
go test ./...
```
