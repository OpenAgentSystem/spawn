# Mock Feishu demo agent

This fixture demonstrates the first OAS supply-layer path without touching real Feishu credentials.

Flow:

1. A mock Feishu message is normalized into a `LaunchRequest`.
2. `packages/supplylayer` checks Action Journal, PoLP, ToS, Adapter Spec, and capability gates.
3. The allowed request can be handed to a spwn runtime such as Codex.

Local gate test:

```bash
cd packages/supplylayer
go test ./...
```

The fixture intentionally does not call real Feishu APIs and does not implement Action Journal storage.
