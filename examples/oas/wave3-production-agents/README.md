# Wave 3 production agent onboarding

This fixture records the five production agent types onboarded to the Spawn
supply layer after the Wave 2 demo agent:

- `narrator-ai`
- `GridFastAgent`
- `hire-app`
- `closure-bot`
- `agency-agents`

Each agent registers through `POST /v1/spawn/dispatch`, declares Adapter Spec
`1.0`, and carries production dispatch evidence in `spwn.yaml`. Legacy
direct-invoke is marked deprecated for every entry.

Validation:

```bash
cd packages/supplylayer
go test ./...
```
