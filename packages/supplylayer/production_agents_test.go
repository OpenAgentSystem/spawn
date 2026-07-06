package supplylayer

import (
	"strings"
	"testing"
)

func TestDefaultWave3ProductionAgentsValidateFiveRegistrations(t *testing.T) {
	agents := DefaultWave3ProductionAgents()
	if len(agents) != 5 {
		t.Fatalf("expected 5 production agents, got %d", len(agents))
	}

	seen := map[string]bool{}
	for _, agent := range agents {
		if err := ValidateProductionAgentRegistration(agent); err != nil {
			t.Fatalf("%s registration should validate: %v", agent.AgentType, err)
		}
		if seen[agent.AgentType] {
			t.Fatalf("duplicate agent type %q", agent.AgentType)
		}
		seen[agent.AgentType] = true
	}

	for _, want := range []string{
		"narrator-ai",
		"GridFastAgent",
		"hire-app",
		"closure-bot",
		"agency-agents",
	} {
		if !seen[want] {
			t.Fatalf("missing Wave 3 agent type %q", want)
		}
	}
}

func TestWave3ProductionAgentsDispatchThroughSupplyLayer(t *testing.T) {
	registry := DefaultWave3Registry()
	for _, agent := range DefaultWave3ProductionAgents() {
		req := validLaunch(agent.AgentType, agent.Capabilities)
		req.AgentID = "agent:" + normalize(agent.AgentType)
		req.IssueURL = agent.IssueURL
		req.PoLPDecision.CapabilityID = "agent.dispatch"

		result := registry.Evaluate(req)
		if result.Decision != DecisionAllow {
			t.Fatalf("%s should dispatch through Spawn supply layer: %v", agent.AgentType, result.Reasons)
		}
	}
}

func TestProductionAgentRegistrationRequiresSpawnAndDirectInvokeDeprecation(t *testing.T) {
	agent := DefaultWave3ProductionAgents()[0]
	agent.SpawnAPIPath = "/legacy/direct-invoke"
	agent.DirectInvokeDeprecated = false

	err := ValidateProductionAgentRegistration(agent)
	if err == nil {
		t.Fatal("expected invalid direct-invoke registration")
	}
	got := err.Error()
	for _, want := range []string{
		"spawn dispatch api path mismatch",
		"legacy direct-invoke must be deprecated",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing reason %q in %q", want, got)
		}
	}
}
