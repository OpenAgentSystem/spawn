package supplylayer

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Decision is the result of a supply-layer launch gate.
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

// Gate names the fail-closed pre-spawn gate that produced a decision.
type Gate string

const (
	GatePoLP          Gate = "polp"
	GateToS           Gate = "tos"
	GateAdapter       Gate = "adapter"
	GateCapability    Gate = "capability"
	GateActionJournal Gate = "action_journal"
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
	Gate     Gate
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
	brand, adapterErr := r.resolveAdapter(req)
	ordered := []struct {
		gate Gate
		err  error
	}{
		{GatePoLP, validatePoLP(req.PoLPDecision)},
		{GateToS, validateToS(brand, req.CustomerData, adapterErr)},
		{GateAdapter, validateAdapter(brand, req, adapterErr)},
		{GateCapability, validateCapabilities(req.Capabilities, brand.AllowedCapabilities, adapterErr)},
		{GateActionJournal, validateActionJournal(req.ActionJournal)},
	}
	for _, step := range ordered {
		if step.err != nil {
			return GateResult{Decision: DecisionDeny, Gate: step.gate, Reasons: []string{step.err.Error()}}
		}
	}
	return GateResult{Decision: DecisionAllow}
}

func (r Registry) resolveAdapter(req LaunchRequest) (BrandPolicy, error) {
	if req.AgentID == "" {
		return BrandPolicy{}, errors.New("missing agent id")
	}
	if req.HumanID == "" {
		return BrandPolicy{}, errors.New("missing human id")
	}
	if req.IssueURL == "" {
		return BrandPolicy{}, errors.New("missing issue url")
	}
	brand, ok := r.Brands[normalize(req.Brand)]
	if !ok {
		return BrandPolicy{}, errors.New("unknown brand")
	}
	return brand, nil
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

func validateToS(brand BrandPolicy, customerData bool, adapterErr error) error {
	if adapterErr != nil {
		return nil
	}
	if brand.ToS == ToSBlocked || brand.ToS == ToSUnknown {
		return errors.New("tos gate not approved")
	}
	if brand.ToS == ToSConditional && customerData {
		return errors.New("tos conditional brand cannot receive customer data")
	}
	return nil
}

func validateAdapter(brand BrandPolicy, req LaunchRequest, adapterErr error) error {
	if adapterErr != nil {
		return adapterErr
	}
	if req.AdapterSpecWant != "" && brand.AdapterSpecVersion != req.AdapterSpecWant {
		return fmt.Errorf("adapter spec mismatch: have %s want %s", brand.AdapterSpecVersion, req.AdapterSpecWant)
	}
	return nil
}

func validateCapabilities(requested, allowed []string, adapterErr error) error {
	if adapterErr != nil {
		return nil
	}
	if len(requested) == 0 {
		return errors.New("capabilities are required")
	}
	if missing := missingCapabilities(requested, allowed); len(missing) > 0 {
		return errors.New("capability not allowed: " + strings.Join(missing, ","))
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
	return missing
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
