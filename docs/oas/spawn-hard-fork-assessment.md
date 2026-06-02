# Spawn hard fork assessment

## Source

- Upstream: `jterrazz/spwn`
- Internal hard fork: `OpenAgentSystem/spawn`
- Upstream default branch: `main`
- Upstream latest checked commit: `b525faa9c0fb616493bbb46525f16b269d7f3166`
- Commit date: 2026-05-04
- Latest upstream release checked: `v0.20.0`, published 2026-05-01
- License: MIT

## Fit for OAS agent supply layer

Spwn already has the primitives OAS needs for the supply layer:

- Runtime adapter registry for agent brands such as Claude Code and Codex.
- Per-agent workspace and HOME isolation inside Docker worlds.
- Project manifests for declarative agent/world composition.
- CLI and API surfaces that can be wrapped by policy gates before launch.

The main gap is that upstream launch paths assume local operator trust. OAS needs every spawn request to be checked against:

- Action Journal sink presence and tenant identity.
- PoLP allow decision with subject, scope, and capability identifiers.
- ToS status from the 21-brand Agent Adapter Spec v1.0 legal review.
- Explicit issue URL for traceability.

The first fork commit adds those gates in `packages/supplylayer` without taking ownership of Action Journal storage, PoLP calculation, or legal review.

## Wave 1 compatibility

Tracked against `Gridltd-DevOps/architecture-decisions#27`.

| Brand | Upstream support | OAS gate state |
|---|---:|---|
| Claude Code | runtime adapter exists | conditional ToS, internal/controlled PoC only |
| Codex | runtime adapter exists | conditional ToS, internal/controlled PoC only |
| Cursor CLI | pending adapter | ToS unknown, denied by default |
| OpenClaw | pending adapter | ToS unknown, denied by default |
| Cline | pending adapter | ToS unknown, denied by default |

## Out of scope in this fork commit

- Action Journal implementation.
- PoLP engine implementation.
- Final legal whitelist/blacklist decisions.
- Real Feishu credential handling.
