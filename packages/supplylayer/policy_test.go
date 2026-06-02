package supplylayer

import (
	"testing"
	"time"
)

func TestEvaluateAllowsWave1InternalLaunch(t *testing.T) {
	registry := DefaultWave1Registry()
	result := registry.Evaluate(LaunchRequest{
		Brand:        "codex",
		AgentID:      "mock-feishu-agent",
		HumanID:      "user:paul",
		IssueURL:     "https://github.com/Gridltd-DevOps/architecture-decisions/issues/29",
		Capabilities: []string{"agent.spawn", "agent.talk"},
		ActionJournal: ActionJournalConfig{
			Endpoint: "https://action-journal.internal/events",
			TenantID: "gridltd",
		},
		PoLPDecision: PoLPDecision{
			Decision:     DecisionAllow,
			SubjectID:    "user:paul",
			ScopeID:      "issue:29",
			CapabilityID: "agent.spawn",
		},
		RequestedAt:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		AdapterSpecWant: "1.0",
	})
	if result.Decision != DecisionAllow {
		t.Fatalf("expected allow, got %s: %v", result.Decision, result.Reasons)
	}
}

func TestEvaluateDeniesMissingAuditAndPoLP(t *testing.T) {
	registry := DefaultWave1Registry()
	result := registry.Evaluate(LaunchRequest{
		Brand:        "codex",
		AgentID:      "agent:mock-feishu",
		HumanID:      "user:paul",
		IssueURL:     "https://github.com/Gridltd-DevOps/architecture-decisions/issues/29",
		Capabilities: []string{"agent.spawn"},
		PoLPDecision: PoLPDecision{
			Decision: DecisionDeny,
			Reason:   "no scoped grant",
		},
	})
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	if result.Gate != GatePoLP {
		t.Fatalf("expected PoLP gate, got %s", result.Gate)
	}
	if got, want := result.Reasons[0], "polp denied: no scoped grant"; got != want {
		t.Fatalf("reason mismatch: got %q want %q", got, want)
	}
}

func TestEvaluateDeniesCustomerDataForConditionalToS(t *testing.T) {
	registry := DefaultWave1Registry()
	result := registry.Evaluate(validLaunch("claude-code", []string{"agent.spawn"}))
	if result.Decision != DecisionAllow {
		t.Fatalf("sanity allow failed: %v", result.Reasons)
	}

	req := validLaunch("claude-code", []string{"agent.spawn"})
	req.CustomerData = true
	result = registry.Evaluate(req)
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	if result.Gate != GateToS {
		t.Fatalf("expected ToS gate, got %s", result.Gate)
	}
	if got, want := result.Reasons[0], "tos conditional brand cannot receive customer data"; got != want {
		t.Fatalf("reason mismatch: got %q want %q", got, want)
	}
}

func TestEvaluateDeniesUnknownToSBeforeCapabilityDrift(t *testing.T) {
	registry := DefaultWave1Registry()
	req := validLaunch("cursor-cli", []string{"agent.spawn", "tool.write"})
	result := registry.Evaluate(req)
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	if result.Gate != GateToS {
		t.Fatalf("expected ToS gate before capability gate, got %s", result.Gate)
	}
	if got, want := result.Reasons[0], "tos gate not approved"; got != want {
		t.Fatalf("reason mismatch: got %q want %q", got, want)
	}
}

func TestEvaluateGateOrder(t *testing.T) {
	tests := []struct {
		name string
		req  LaunchRequest
		gate Gate
	}{
		{
			name: "polp first",
			req: func() LaunchRequest {
				req := validLaunch("codex", []string{"agent.spawn"})
				req.PoLPDecision.Decision = DecisionDeny
				req.PoLPDecision.Reason = "missing grant"
				req.CustomerData = true
				req.AdapterSpecWant = "9.9"
				req.ActionJournal = ActionJournalConfig{}
				return req
			}(),
			gate: GatePoLP,
		},
		{
			name: "tos second",
			req: func() LaunchRequest {
				req := validLaunch("codex", []string{"not.allowed"})
				req.CustomerData = true
				req.AdapterSpecWant = "9.9"
				req.ActionJournal = ActionJournalConfig{}
				return req
			}(),
			gate: GateToS,
		},
		{
			name: "adapter third",
			req: func() LaunchRequest {
				req := validLaunch("codex", []string{"not.allowed"})
				req.AdapterSpecWant = "9.9"
				req.ActionJournal = ActionJournalConfig{}
				return req
			}(),
			gate: GateAdapter,
		},
		{
			name: "capability fourth",
			req: func() LaunchRequest {
				req := validLaunch("codex", []string{"not.allowed"})
				req.ActionJournal = ActionJournalConfig{}
				return req
			}(),
			gate: GateCapability,
		},
		{
			name: "action journal last",
			req: func() LaunchRequest {
				req := validLaunch("codex", []string{"agent.spawn"})
				req.ActionJournal = ActionJournalConfig{}
				return req
			}(),
			gate: GateActionJournal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultWave1Registry().Evaluate(tt.req)
			if result.Decision != DecisionDeny {
				t.Fatalf("expected deny, got %s", result.Decision)
			}
			if result.Gate != tt.gate {
				t.Fatalf("expected gate %s, got %s with reasons %v", tt.gate, result.Gate, result.Reasons)
			}
		})
	}
}

func TestEvaluateDeniesEmptyCapabilities(t *testing.T) {
	result := DefaultWave1Registry().Evaluate(validLaunch("codex", nil))
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	if result.Gate != GateCapability {
		t.Fatalf("expected capability gate, got %s", result.Gate)
	}
	if got, want := result.Reasons[0], "capabilities are required"; got != want {
		t.Fatalf("reason mismatch: got %q want %q", got, want)
	}
}

func validLaunch(brand string, caps []string) LaunchRequest {
	return LaunchRequest{
		Brand:        brand,
		AgentID:      "agent:mock-feishu",
		HumanID:      "user:paul",
		IssueURL:     "https://github.com/Gridltd-DevOps/architecture-decisions/issues/29",
		Capabilities: caps,
		ActionJournal: ActionJournalConfig{
			Endpoint: "https://action-journal.internal/events",
			TenantID: "gridltd",
		},
		PoLPDecision: PoLPDecision{
			Decision:     DecisionAllow,
			SubjectID:    "user:paul",
			ScopeID:      "issue:29",
			CapabilityID: "agent.spawn",
		},
		AdapterSpecWant: "1.0",
	}
}
