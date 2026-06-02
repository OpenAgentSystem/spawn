# Mock Feishu demo agent

This fixture demonstrates the first OAS supply-layer path without touching real Feishu credentials.

Flow:

1. A mock Feishu message is normalized into a `LaunchRequest`.
2. `packages/supplylayer` checks Action Journal, PoLP, ToS, Adapter Spec, and capability gates.
3. The allowed request can be handed to a spwn runtime such as Codex.
4. `packages/agentplatform` exposes the Wave 3 HTTP path:
   - `POST /agents/register` accepts the Adapter Spec v1 mock Feishu manifest.
   - `POST /dispatch` gates the task and sends it to the registered mock agent.
   - `GET /tasks/:id/status` returns the completed dispatch status.

Local gate test:

```bash
cd packages/supplylayer
go test ./...

cd ../agentplatform
go test ./...
```

The fixture intentionally does not call real Feishu APIs and does not implement Action Journal storage.
