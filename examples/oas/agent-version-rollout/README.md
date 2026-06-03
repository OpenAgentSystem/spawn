# Agent version rollout demo (W3-25)

This fixture demonstrates the W3-25 acceptance criterion: one agent migrating from
`v1` to `v2` via a 5% → 50% → 100% canary rollout.

The demo uses a single agent type, `versioned-bot`, captured at three points in
its rollout lifecycle:

| Stage | `agent.yaml` | `v_active` | `v_canary` | `canary_percent` |
| --- | --- | --- | --- | --- |
| Stage 1 (5%)   | `spwn/agents/versioned-bot/stage-1-5pct/agent.yaml`   | `v1` | `v2` | `5`   |
| Stage 2 (50%)  | `spwn/agents/versioned-bot/stage-2-50pct/agent.yaml`  | `v1` | `v2` | `50`  |
| Stage 3 (100%) | `spwn/agents/versioned-bot/stage-3-100pct/agent.yaml` | `v2` | `""` | `0`   |

Stages 1–2 keep `v1` as `v_active` and shift more traffic to the `v_canary` `v2`
each step. Stage 3 promotes `v2` to `v_active` and clears the canary fields —
this is the "rollout complete" terminal state.

How it works inside Spawn:

- `packages/architect.resolveAgentRollout` hashes `(worldID, agentName)` into a
  stable 0–99 bucket; when `bucket < canary_percent` and `v_canary` is set, the
  agent is routed to the canary version. The cohort is recorded on the world's
  `AgentRecord` (`Version`, `RolloutCohort`, `CanaryPercent`).
- Spawn-time and hot-deploy (`Architect.DeployAgent`) both call into the same
  resolver, so newly-attached agents respect the active stage immediately.
- The runtime exec wires the resolved cohort into the container environment
  (`SPWN_AGENT_VERSION`, `SPWN_ROLLOUT_COHORT`) and journals one
  `(version, cohort, success)` triple per session.

Acceptance evidence:

- `packages/architect.TestW3_25_VersionedBotV1ToV2Migration` drives one agent
  through the three stages over a 1000-world cohort sample and asserts each
  stage's canary share is within rollout tolerance.
- `packages/architect.SideBySideSuccessRates` aggregates the journal triples into
  the `v1/active` vs `v2/canary` (and eventually `v2/active`) metric the
  operator reads to decide promote vs roll-back.

See `docs/oas/agent-version-rollout.md` for the operator runbook.
