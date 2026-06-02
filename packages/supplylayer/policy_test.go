package supplylayer

import (
	"strings"
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
		Brand: "codex",
		PoLPDecision: PoLPDecision{
			Decision: DecisionDeny,
			Reason:   "no scoped grant",
		},
	})
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	joined := strings.Join(result.Reasons, "|")
	for _, want := range []string{
		"missing agent id",
		"missing human id",
		"missing issue url",
		"action journal sink is required",
		"polp denied: no scoped grant",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing reason %q in %v", want, result.Reasons)
		}
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
	if got := strings.Join(result.Reasons, "|"); !strings.Contains(got, "tos conditional brand cannot receive customer data") {
		t.Fatalf("expected tos customer-data reason, got %v", result.Reasons)
	}
}

func TestEvaluateDeniesUnknownToSAndCapabilityDrift(t *testing.T) {
	registry := DefaultWave1Registry()
	req := validLaunch("cursor-cli", []string{"agent.spawn", "tool.write"})
	result := registry.Evaluate(req)
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	got := strings.Join(result.Reasons, "|")
	for _, want := range []string{"tos gate not approved", "capability not allowed: tool.write"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing reason %q in %v", want, result.Reasons)
		}
	}
}

func TestEvaluateDeniesEmptyCapabilities(t *testing.T) {
	registry := DefaultWave1Registry()
	req := validLaunch("codex", nil)
	result := registry.Evaluate(req)
	if result.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", result.Decision)
	}
	if got := strings.Join(result.Reasons, "|"); !strings.Contains(got, "capabilities are required") {
		t.Fatalf("expected capabilities reason, got %v", result.Reasons)
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
