package supplylayer

import "testing"

func TestMockFeishuDemoGate(t *testing.T) {
	req := mockFeishuLaunch()

	result := DefaultWave1Registry().Evaluate(req)
	if result.Decision != DecisionAllow {
		t.Fatalf("mock Feishu demo should pass pre-spawn gates: %v", result.Reasons)
	}
}

func TestMockFeishuDemoFailurePaths(t *testing.T) {
	tests := []struct {
		name string
		req  LaunchRequest
		gate Gate
	}{
		{
			name: "missing scoped grant",
			req: func() LaunchRequest {
				req := mockFeishuLaunch()
				req.PoLPDecision.Decision = DecisionDeny
				req.PoLPDecision.Reason = "mock Feishu thread lacks scoped spawn grant"
				return req
			}(),
			gate: GatePoLP,
		},
		{
			name: "customer data blocked by conditional tos",
			req: func() LaunchRequest {
				req := mockFeishuLaunch()
				req.CustomerData = true
				return req
			}(),
			gate: GateToS,
		},
		{
			name: "adapter spec drift",
			req: func() LaunchRequest {
				req := mockFeishuLaunch()
				req.AdapterSpecWant = "2.0"
				return req
			}(),
			gate: GateAdapter,
		},
		{
			name: "unsupported tool request",
			req: func() LaunchRequest {
				req := mockFeishuLaunch()
				req.Capabilities = []string{"agent.spawn", "tool.admin"}
				return req
			}(),
			gate: GateCapability,
		},
		{
			name: "missing audit sink",
			req: func() LaunchRequest {
				req := mockFeishuLaunch()
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

func mockFeishuLaunch() LaunchRequest {
	req := validLaunch("codex", []string{"agent.spawn", "agent.talk"})
	req.AgentID = "agent:mock-feishu-router"
	req.IssueURL = "https://github.com/Gridltd-DevOps/architecture-decisions/issues/29"
	req.PoLPDecision.ScopeID = "mock-feishu:thread:internal-sandbox"
	req.PoLPDecision.CapabilityID = "agent.spawn"
	return req
}
