package architect

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"spwn.sh/packages/agent"
)

const (
	rolloutCohortActive = "active"
	rolloutCohortCanary = "canary"
)

// RolloutDecision is the resolved version routing decision for one
// agent in one world. Bucket is 0-99 and stable for the same
// world/agent pair so repeated spawns do not flap between cohorts.
type RolloutDecision struct {
	Version       string
	Cohort        string
	ActiveVersion string
	CanaryVersion string
	CanaryPercent int
	Bucket        int
}

// RolloutOutcome is one completed run used by SideBySideSuccessRates.
type RolloutOutcome struct {
	AgentName string
	Version   string
	Cohort    string
	Success   bool
}

// RolloutSuccessRate is the side-by-side success metric for one
// version/cohort pair.
type RolloutSuccessRate struct {
	Version     string
	Cohort      string
	Successes   int
	Total       int
	SuccessRate float64
}

func resolveAgentRollout(agentName, worldID string, m *agent.Manifest) (RolloutDecision, error) {
	if m == nil {
		return RolloutDecision{}, nil
	}
	active := strings.TrimSpace(m.Rollout.ActiveVersion)
	if active == "" {
		active = strings.TrimSpace(m.Version)
	}
	canary := strings.TrimSpace(m.Rollout.CanaryVersion)
	percent := m.Rollout.CanaryPercent
	if percent < 0 || percent > 100 {
		return RolloutDecision{}, fmt.Errorf("agent %q rollout canary_percent must be between 0 and 100, got %d", agentName, percent)
	}

	decision := RolloutDecision{
		Version:       active,
		Cohort:        rolloutCohortActive,
		ActiveVersion: active,
		CanaryVersion: canary,
		CanaryPercent: percent,
		Bucket:        rolloutBucket(worldID, agentName),
	}
	if canary != "" && percent > 0 && decision.Bucket < percent {
		decision.Version = canary
		decision.Cohort = rolloutCohortCanary
	}
	return decision, nil
}

func rolloutBucket(worldID, agentName string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(worldID))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(agentName))
	return int(h.Sum32() % 100)
}

// SideBySideSuccessRates groups completed outcomes by version/cohort
// and returns deterministic, UI-friendly success metrics.
func SideBySideSuccessRates(outcomes []RolloutOutcome) []RolloutSuccessRate {
	byKey := map[string]*RolloutSuccessRate{}
	for _, outcome := range outcomes {
		version := strings.TrimSpace(outcome.Version)
		if version == "" {
			version = "unknown"
		}
		cohort := strings.TrimSpace(outcome.Cohort)
		if cohort == "" {
			cohort = rolloutCohortActive
		}
		key := version + "\x00" + cohort
		rec := byKey[key]
		if rec == nil {
			rec = &RolloutSuccessRate{Version: version, Cohort: cohort}
			byKey[key] = rec
		}
		rec.Total++
		if outcome.Success {
			rec.Successes++
		}
		rec.SuccessRate = float64(rec.Successes) / float64(rec.Total)
	}

	out := make([]RolloutSuccessRate, 0, len(byKey))
	for _, rec := range byKey {
		out = append(out, *rec)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Version == out[j].Version {
			return out[i].Cohort < out[j].Cohort
		}
		return out[i].Version < out[j].Version
	})
	return out
}
