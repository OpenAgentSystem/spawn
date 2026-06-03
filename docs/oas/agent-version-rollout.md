# Agent version rollout (W3-25)

Spawn supports staged version migration for any agent: declare a `version:` on
the manifest, then carve out a percentage of `(world, agent)` cohorts to a
`v_canary` candidate before promoting it to `v_active`. The same fields drive
the cold-spawn path (`Architect.Spawn`) and hot-deploy (`Architect.DeployAgent`).

This page is the W3-25 operator runbook for migrating one agent from `v1` to
`v2` via a 5% → 50% → 100% canary.

## Manifest fields

```yaml
# spwn/agents/<name>/agent.yaml
name: versioned-bot
version: v1           # current deployable tag for this agent
rollout:
  v_active: v1        # stable version receiving baseline traffic
  v_canary: v2        # candidate version under test (may be empty)
  canary_percent: 5   # share of (world, agent) cohorts routed to v_canary
```

`canary_percent` is bounded `[0, 100]` and validated at spawn time. Empty
`v_canary` (or `canary_percent: 0`) collapses to "100% active" — the legacy
behaviour for agents that have not opted into versioning.

## Cohort assignment

`packages/architect.resolveAgentRollout` deterministically hashes
`(worldID, agentName)` into a 0–99 bucket. When `bucket < canary_percent` and
`v_canary` is non-empty, the agent runs the canary version; otherwise the
active version. The bucket is stable for the same `(world, agent)` pair, so:

- Restarts of the same world keep the same cohort — no flapping.
- Different worlds running the same agent split into the configured share.

The resolved `(Version, RolloutCohort, CanaryPercent)` is stamped onto the
world's `AgentRecord` and surfaces through:

- Container env: `SPWN_AGENT_VERSION` and `SPWN_ROLLOUT_COHORT`.
- Journal events (one triple per session, see `agent.AppendJournalWithRollout`).
- Side-by-side metric (`architect.SideBySideSuccessRates`).

## Staged migration: v1 → v2

The fixture in `examples/oas/agent-version-rollout/` captures one agent at the
three checkpoints. The operator's commit history would look like:

### Stage 1 — open a 5% canary

```yaml
version: v1
rollout:
  v_active: v1
  v_canary: v2
  canary_percent: 5
```

Spawn (or hot-deploy) starts routing ~5% of worlds to `v2`. The remaining ~95%
stay on `v1`. Read the side-by-side metric for at least one full traffic cycle
before advancing; if canary success-rate trails active, revert to a single-line
`version: v1` manifest.

### Stage 2 — broaden to 50%

```yaml
version: v1
rollout:
  v_active: v1
  v_canary: v2
  canary_percent: 50
```

Half the cohorts now run `v2`. The bucket assignment is sticky — every world
that was on canary at 5% is still on canary at 50%, plus net-new canary
worlds. This keeps per-world behaviour observable across the ramp.

### Stage 3 — promote `v2` to active

```yaml
version: v2
rollout:
  v_active: v2
  # v_canary and canary_percent cleared — rollout complete
```

`v2` becomes the unconditional version. New worlds and existing worlds (next
spawn) cohort-resolve to `v2/active`. The journal continues emitting
`(v2, active)` rows so the metric still shows a clean v1→v2 cut-over window.

## Reading the side-by-side metric

`architect.SideBySideSuccessRates` groups journal outcomes by
`(version, cohort)` and reports `successes / total` per group. The expected
shape during each stage:

| Stage | Rows reported |
| --- | --- |
| Stage 1 (5%)   | `v1/active`, `v2/canary` |
| Stage 2 (50%)  | `v1/active`, `v2/canary` |
| Stage 3 (100%) | `v2/active` (older `v1/active` and `v2/canary` rows fade as outcomes age out of the window) |

A safe promote requires the canary success-rate to be within an acceptable
delta of the active success-rate over the operator's chosen window (typically
matching the agent's natural request volume).

## Acceptance test

`packages/architect.TestW3_25_VersionedBotV1ToV2Migration` exercises the full
flow:

1. Loads each stage manifest from the fixture above.
2. Resolves cohort for 1000 deterministic world IDs at each stage.
3. Verifies the canary share matches `canary_percent` within tolerance.
4. Drives a synthetic outcome stream through `SideBySideSuccessRates` and
   asserts the `v1/active` vs `v2/canary` rows are reported.

Run it with:

```bash
go test ./packages/architect -run TestW3_25_VersionedBotV1ToV2Migration -v
```
