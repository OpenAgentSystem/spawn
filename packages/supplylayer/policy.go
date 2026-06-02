package supplylayer

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Decision is the result of a supply-layer launch gate.
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

// ToSState records the legal/commercial status for a runtime brand.
type ToSState string

const (
	ToSAllowed     ToSState = "allowed"
	ToSConditional ToSState = "conditional"
	ToSBlocked     ToSState = "blocked"
	ToSUnknown     ToSState = "unknown"
)

// BrandPolicy is the per-brand supply policy derived from Agent Adapter Spec
// v1.0 and the legal review tracked in Gridltd-DevOps/architecture-decisions#27.
type BrandPolicy struct {
	Brand                string
	AdapterSpecVersion   string
	AllowedCapabilities  []string
	ToS                  ToSState
	ConditionalUseReason string
}

// Registry is the in-memory supply policy catalog used by the fork.
type Registry struct {
	Brands map[string]BrandPolicy
}

// LaunchRequest is the normalized intent checked before an agent is spawned.
type LaunchRequest struct {
	Brand           string
	AgentID         string
	HumanID         string
	IssueURL        string
	Capabilities    []string
	ActionJournal   ActionJournalConfig
	PoLPDecision    PoLPDecision
	CustomerData    bool
	RequestedAt     time.Time
	AdapterSpecWant string
}

// ActionJournalConfig points to the append-only audit sink owned outside this
// repository. This package validates presence and identity only; it does not
// implement Action Journal storage.
type ActionJournalConfig struct {
	Endpoint string
	TenantID string
}

// PoLPDecision is the output of the PoLP engine owned outside this repository.
type PoLPDecision struct {
	Decision     Decision
	SubjectID    string
	ScopeID      string
	CapabilityID string
	Reason       string
}

// GateResult explains an allow/deny outcome.
type GateResult struct {
	Decision Decision
	Reasons  []string
}

// DefaultWave1Registry returns the first five brands called out by #27.
func DefaultWave1Registry() Registry {
	return Registry{Brands: map[string]BrandPolicy{
		"claude-code": {
			Brand:              "claude-code",
			AdapterSpecVersion: "1.0",
			AllowedCapabilities: []string{
				"agent.spawn",
				"agent.talk",
				"tool.read",
				"tool.write",
			},
			ToS:                  ToSConditional,
			ConditionalUseReason: "internal研发/受控PoC only until legal review completes",
		},
		"codex": {
			Brand:              "codex",
			AdapterSpecVersion: "1.0",
			AllowedCapabilities: []string{
				"agent.spawn",
				"agent.talk",
				"tool.read",
				"tool.write",
			},
			ToS:                  ToSConditional,
			ConditionalUseReason: "internal研发/受控PoC only until legal review completes",
		},
		"cursor-cli": {
			Brand:              "cursor-cli",
			AdapterSpecVersion: "1.0",
			AllowedCapabilities: []string{
				"agent.spawn",
				"agent.talk",
			},
			ToS: ToSUnknown,
		},
		"openclaw": {
			Brand:              "openclaw",
			AdapterSpecVersion: "1.0",
			AllowedCapabilities: []string{
				"agent.spawn",
				"agent.talk",
			},
			ToS: ToSUnknown,
		},
		"cline": {
			Brand:              "cline",
			AdapterSpecVersion: "1.0",
			AllowedCapabilities: []string{
				"agent.spawn",
				"agent.talk",
			},
			ToS: ToSUnknown,
		},
	}}
}

// Evaluate enforces the OAS supply-layer pre-spawn contract.
func (r Registry) Evaluate(req LaunchRequest) GateResult {
	var reasons []string
	brand, ok := r.Brands[normalize(req.Brand)]
	if !ok {
		reasons = append(reasons, "unknown brand")
		return GateResult{Decision: DecisionDeny, Reasons: reasons}
	}
	if req.AgentID == "" {
		reasons = append(reasons, "missing agent id")
	}
	if req.HumanID == "" {
		reasons = append(reasons, "missing human id")
	}
	if req.IssueURL == "" {
		reasons = append(reasons, "missing issue url")
	}
	if err := validateActionJournal(req.ActionJournal); err != nil {
		reasons = append(reasons, err.Error())
	}
	if err := validatePoLP(req.PoLPDecision); err != nil {
		reasons = append(reasons, err.Error())
	}
	if req.AdapterSpecWant != "" && brand.AdapterSpecVersion != req.AdapterSpecWant {
		reasons = append(reasons, fmt.Sprintf("adapter spec mismatch: have %s want %s", brand.AdapterSpecVersion, req.AdapterSpecWant))
	}
	if brand.ToS == ToSBlocked || brand.ToS == ToSUnknown {
		reasons = append(reasons, "tos gate not approved")
	}
	if brand.ToS == ToSConditional && req.CustomerData {
		reasons = append(reasons, "tos conditional brand cannot receive customer data")
	}
	if len(req.Capabilities) == 0 {
		reasons = append(reasons, "capabilities are required")
	}
	if missing := missingCapabilities(req.Capabilities, brand.AllowedCapabilities); len(missing) > 0 {
		reasons = append(reasons, "capability not allowed: "+strings.Join(missing, ","))
	}
	if len(reasons) > 0 {
		return GateResult{Decision: DecisionDeny, Reasons: reasons}
	}
	return GateResult{Decision: DecisionAllow}
}

func validateActionJournal(cfg ActionJournalConfig) error {
	if cfg.Endpoint == "" || cfg.TenantID == "" {
		return errors.New("action journal sink is required")
	}
	return nil
}

func validatePoLP(decision PoLPDecision) error {
	if decision.Decision != DecisionAllow {
		if decision.Reason != "" {
			return fmt.Errorf("polp denied: %s", decision.Reason)
		}
		return errors.New("polp denied")
	}
	if decision.SubjectID == "" || decision.ScopeID == "" || decision.CapabilityID == "" {
		return errors.New("polp allow decision is missing subject, scope, or capability")
	}
	return nil
}

func missingCapabilities(requested, allowed []string) []string {
	allowedSet := map[string]bool{}
	for _, c := range allowed {
		allowedSet[c] = true
	}
	var missing []string
	for _, c := range requested {
		if !allowedSet[c] {
			missing = append(missing, c)
		}
	}
	sort.Strings(missing)
	return missing
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
