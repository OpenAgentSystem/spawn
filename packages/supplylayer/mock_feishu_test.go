package supplylayer

import "testing"

func TestMockFeishuDemoGate(t *testing.T) {
	req := validLaunch("codex", []string{"agent.spawn", "agent.talk"})
	req.AgentID = "agent:mock-feishu-router"
	req.IssueURL = "https://github.com/Gridltd-DevOps/architecture-decisions/issues/29"

	result := DefaultWave1Registry().Evaluate(req)
	if result.Decision != DecisionAllow {
		t.Fatalf("mock Feishu demo should pass pre-spawn gates: %v", result.Reasons)
	}
}
