# Mock Feishu demo agent

This fixture demonstrates the first OAS supply-layer path without touching real Feishu credentials.

Flow:

1. A mock Feishu message is normalized into a `LaunchRequest`.
2. `packages/supplylayer` checks PoLP, ToS, Adapter Spec, capability, and Action Journal gates in that order.
3. The allowed request can be handed to a spwn runtime such as Codex.

Failure paths covered by `TestMockFeishuDemoFailurePaths`:

- missing PoLP scoped grant
- customer data on a conditional-ToS brand
- adapter spec version drift
- unsupported capability request
- missing Action Journal sink

Local gate test:

```bash
cd packages/supplylayer
go test ./...
```

The fixture intentionally does not call real Feishu APIs and does not implement Action Journal storage.
