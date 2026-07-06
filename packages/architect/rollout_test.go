package architect

import (
	"fmt"
	"testing"

	"spwn.sh/packages/agent"
)

func TestResolveAgentRollout_CanaryProgression(t *testing.T) {
	m := &agent.Manifest{
		Version: "v1",
		Rollout: agent.RolloutConfig{
			ActiveVersion: "v1",
			CanaryVersion: "v2",
		},
	}

	stages := []struct {
		percent int
		want    bool
	}{
		{percent: 5, want: true},
		{percent: 5, want: false},
		{percent: 50, want: true},
		{percent: 50, want: false},
		{percent: 100, want: true},
	}

	for _, stage := range stages {
		m.Rollout.CanaryPercent = stage.percent
		worldID := findRolloutFixtureWorldID(t, "neo", stage.percent, stage.want)
		decision, err := resolveAgentRollout("neo", worldID, m)
		if err != nil {
			t.Fatalf("resolveAgentRollout: %v", err)
		}
		want := rolloutCohortActive
		if stage.want {
			want = rolloutCohortCanary
		}
		if decision.Cohort != want {
			t.Fatalf("%s at %d%% cohort = %q, want %q", worldID, stage.percent, decision.Cohort, want)
		}
		if decision.Cohort == rolloutCohortCanary && decision.Version != "v2" {
			t.Fatalf("canary version = %q, want v2", decision.Version)
		}
		if decision.Cohort == rolloutCohortActive && decision.Version != "v1" {
			t.Fatalf("active version = %q, want v1", decision.Version)
		}
	}
}

func findRolloutFixtureWorldID(t *testing.T, agentName string, percent int, wantCanary bool) string {
	t.Helper()
	for i := 0; i < 1000; i++ {
		worldID := fmt.Sprintf("world-%03d", i)
		if rolloutBucket(worldID, agentName) < percent == wantCanary {
			return worldID
		}
	}
	t.Fatalf("no rollout fixture found for %d%% canary=%t", percent, wantCanary)
	return ""
}

func TestResolveAgentRollout_RejectsInvalidPercent(t *testing.T) {
	_, err := resolveAgentRollout("neo", "world", &agent.Manifest{
		Version: "v1",
		Rollout: agent.RolloutConfig{
			CanaryVersion: "v2",
			CanaryPercent: 101,
		},
	})
	if err == nil {
		t.Fatal("expected invalid canary percent error")
	}
}

func TestSideBySideSuccessRates(t *testing.T) {
	got := SideBySideSuccessRates([]RolloutOutcome{
		{AgentName: "neo", Version: "v1", Cohort: "active", Success: true},
		{AgentName: "neo", Version: "v1", Cohort: "active", Success: false},
		{AgentName: "neo", Version: "v2", Cohort: "canary", Success: true},
		{AgentName: "neo", Version: "v2", Cohort: "canary", Success: true},
	})

	if len(got) != 2 {
		t.Fatalf("metrics count = %d, want 2: %+v", len(got), got)
	}
	if got[0].Version != "v1" || got[0].Successes != 1 || got[0].Total != 2 || got[0].SuccessRate != 0.5 {
		t.Fatalf("v1 metric drifted: %+v", got[0])
	}
	if got[1].Version != "v2" || got[1].Successes != 2 || got[1].Total != 2 || got[1].SuccessRate != 1 {
		t.Fatalf("v2 metric drifted: %+v", got[1])
	}
}
