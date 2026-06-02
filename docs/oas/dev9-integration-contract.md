# Dev #9 integration contract

This fork exposes supply-layer checks as a pure Go package so the multi-agent collaboration engine can call it before dispatching an agent into a Saga step.

## Pre-spawn request

Required fields:

- `brand`
- `agent_id`
- `human_id`
- `issue_url`
- `capabilities`
- `action_journal.endpoint`
- `action_journal.tenant_id`
- `polp.decision`
- `polp.subject_id`
- `polp.scope_id`
- `polp.capability_id`

## Saga behavior

Dev #9 should treat `DecisionDeny` as a non-retryable precondition failure. It should not start compensation for a step that never spawned an agent. Compensation should start only after an allowed spawn has emitted Action Journal evidence and then fails during execution.

Gate evaluation is fail-closed and short-circuits in this order:

1. PoLP
2. ToS
3. Adapter
4. Capability
5. Action Journal

Dev #9 should persist the failed `Gate` value in its Saga step result so Mission Control can distinguish authorization denial, legal/data-use denial, adapter drift, capability drift, and missing audit plumbing.

## Deadline behavior

The collaboration engine can run this gate synchronously. It has no network calls in this package; external Action Journal and PoLP lookups happen before constructing the request.
