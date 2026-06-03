package architect

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"spwn.sh/packages/agent"
)

// TestW3_25_VersionedBotV1ToV2Migration is the W3-25 acceptance test.
// It loads the three stage manifests from the
// examples/oas/agent-version-rollout fixture, drives one agent through
// the 5% -> 50% -> 100% canary progression over a deterministic
// world-id cohort, and verifies:
//   - the cohort share routed to v_canary matches canary_percent
//     within tolerance at each pre-promotion stage;
//   - the terminal stage routes 100% of cohorts to v2/active (no canary
//     traffic remains);
//   - SideBySideSuccessRates groups outcomes by (version, cohort) so the
//     operator sees v1/active and v2/canary in parallel during the
//     ramp.
func TestW3_25_VersionedBotV1ToV2Migration(t *testing.T) {
	root := repoRootForTest(t)
	fixtureBase := filepath.Join(
		root,
		"examples", "oas", "agent-version-rollout",
		"spwn", "agents", "versioned-bot",
	)

	type stage struct {
		dir             string
		wantActive      string
		wantCanary      string
		wantPercent     int
		wantVersionTag  string
		wantTerminalAll string // when set, every world must resolve to this version
	}

	stages := []stage{
		{dir: "stage-1-5pct", wantActive: "v1", wantCanary: "v2", wantPercent: 5, wantVersionTag: "v1"},
		{dir: "stage-2-50pct", wantActive: "v1", wantCanary: "v2", wantPercent: 50, wantVersionTag: "v1"},
		{dir: "stage-3-100pct", wantActive: "v2", wantCanary: "", wantPercent: 0, wantVersionTag: "v2", wantTerminalAll: "v2"},
	}

	const sampleSize = 1000

	// Aggregate one synthetic outcome per (worldID, stage) into a
	// single side-by-side run so the metric reports v1/active and
	// v2/canary in parallel — what the operator reads during the ramp.
	var outcomes []RolloutOutcome

	for _, st := range stages {
		t.Run(st.dir, func(t *testing.T) {
			m, err := agent.LoadManifestPath(filepath.Join(fixtureBase, st.dir))
			if err != nil {
				t.Fatalf("load manifest %s: %v", st.dir, err)
			}
			if m == nil {
				t.Fatalf("manifest %s not found", st.dir)
			}
			if m.Version != st.wantVersionTag {
				t.Fatalf("manifest version = %q, want %q", m.Version, st.wantVersionTag)
			}
			if m.Rollout.ActiveVersion != st.wantActive {
				t.Fatalf("rollout.v_active = %q, want %q", m.Rollout.ActiveVersion, st.wantActive)
			}
			if m.Rollout.CanaryVersion != st.wantCanary {
				t.Fatalf("rollout.v_canary = %q, want %q", m.Rollout.CanaryVersion, st.wantCanary)
			}
			if m.Rollout.CanaryPercent != st.wantPercent {
				t.Fatalf("rollout.canary_percent = %d, want %d", m.Rollout.CanaryPercent, st.wantPercent)
			}

			canaryHits := 0
			for i := 0; i < sampleSize; i++ {
				worldID := fmt.Sprintf("w3-25-world-%04d", i)
				decision, err := resolveAgentRollout("versioned-bot", worldID, m)
				if err != nil {
					t.Fatalf("resolveAgentRollout: %v", err)
				}
				if st.wantTerminalAll != "" {
					if decision.Version != st.wantTerminalAll || decision.Cohort != rolloutCohortActive {
						t.Fatalf("terminal stage: world %s routed to %s/%s, want %s/active",
							worldID, decision.Version, decision.Cohort, st.wantTerminalAll)
					}
				}
				if decision.Cohort == rolloutCohortCanary {
					canaryHits++
				}
				outcomes = append(outcomes, RolloutOutcome{
					AgentName: "versioned-bot",
					Version:   decision.Version,
					Cohort:    decision.Cohort,
					Success:   true,
				})
			}

			if st.wantTerminalAll == "" {
				// Pre-promotion stages: canary share must match
				// canary_percent within +/- 4 percentage points over
				// 1000 worlds — fnv32a hashing is uniform enough for
				// this band to be reliable without flakes.
				gotPercent := float64(canaryHits) * 100.0 / float64(sampleSize)
				lo := float64(st.wantPercent) - 4
				hi := float64(st.wantPercent) + 4
				if gotPercent < lo || gotPercent > hi {
					t.Fatalf("stage %s canary share = %.2f%%, want %d%% +/- 4pp", st.dir, gotPercent, st.wantPercent)
				}
				if canaryHits == 0 {
					t.Fatalf("stage %s produced zero canary cohorts; rollout would never observe v2", st.dir)
				}
			}
		})
	}

	metrics := SideBySideSuccessRates(outcomes)
	if len(metrics) == 0 {
		t.Fatal("SideBySideSuccessRates returned no rows")
	}
	want := map[string]bool{
		"v1/active": true,
		"v2/canary": true,
		"v2/active": true, // terminal stage promotes v2 to active
	}
	saw := map[string]bool{}
	for _, m := range metrics {
		saw[m.Version+"/"+m.Cohort] = true
		if m.Total == 0 {
			t.Fatalf("metric row %s/%s has Total=0", m.Version, m.Cohort)
		}
		if m.SuccessRate != 1 {
			t.Fatalf("metric row %s/%s success rate = %v, want 1 (all synthetic outcomes succeed)",
				m.Version, m.Cohort, m.SuccessRate)
		}
	}
	for key := range want {
		if !saw[key] {
			t.Fatalf("side-by-side metric missing %s row; got %+v", key, metrics)
		}
	}
}

// repoRootForTest walks up from the test file location to the repo
// root (identified by the go.work file). Lets the test reference the
// examples/ fixture without depending on the caller's cwd.
func repoRootForTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 10; i++ {
		if fileExists(filepath.Join(dir, "go.work")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate repo root (go.work) from test file")
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
