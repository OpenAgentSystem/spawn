package architect

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"spwn.sh/packages/platform"
	"spwn.sh/packages/supplylayer"
	"spwn.sh/packages/world/models"
)

const (
	launchCapabilitySpawn = "agent.spawn"
	launchCapabilityTalk  = "agent.talk"
)

type launchAdmission struct {
	event ActionJournalEnvelope
}

type launchGateError struct {
	Codes   []string
	Details []string
}

func (e launchGateError) Error() string {
	return fmt.Sprintf("supply layer denied agent launch (reason codes: %s; details: %s)", strings.Join(e.Codes, ","), strings.Join(e.Details, "; "))
}

// AuthorizeAgentLaunch applies the OAS supply-layer gate for callers that
// launch through their own exec path while still relying on Architect state.
// It records pre/allow/deny Action Journal envelopes and returns a
// reason-coded error on denial.
func (a *Architect) AuthorizeAgentLaunch(ctx context.Context, world *models.World, agentName, capability, launchPath string) error {
	_, err := a.authorizeAgentLaunch(ctx, world, agentName, capability, launchPath)
	return err
}

// RecordAgentLaunchOutcome lets external launch paths complete the Action
// Journal envelope after they have run their runtime command.
func (a *Architect) RecordAgentLaunchOutcome(ctx context.Context, world *models.World, agentName, capability, launchPath string, err error) {
	if world == nil {
		return
	}
	admission := &launchAdmission{event: ActionJournalEnvelope{
		Phase:       "post-launch",
		WorldID:     world.ID,
		AgentID:     agentIDForLaunch(world, agentName),
		AgentName:   agentName,
		Capability:  capability,
		LaunchPath:  launchPath,
		RequestedAt: time.Now().UTC(),
	}}
	a.recordAgentLaunchOutcome(ctx, admission, "post-launch", err)
}

func (a *Architect) authorizeAgentLaunch(ctx context.Context, world *models.World, agentName, capability, launchPath string) (*launchAdmission, error) {
	if world == nil {
		return nil, fmt.Errorf("world is required")
	}
	requestedAt := time.Now().UTC()
	agentID := agentIDForLaunch(world, agentName)
	base := ActionJournalEnvelope{
		Phase:       "pre-launch",
		WorldID:     world.ID,
		AgentID:     agentID,
		AgentName:   agentName,
		Capability:  capability,
		LaunchPath:  launchPath,
		RequestedAt: requestedAt,
	}
	a.recordActionJournal(ctx, base)

	req, missingCodes, missingDetails := launchRequestFromEnv(world, agentID, capability, requestedAt)
	result := supplylayer.DefaultWave1Registry().Evaluate(req)
	codes := append([]string{}, missingCodes...)
	details := append([]string{}, missingDetails...)
	if result.Decision == supplylayer.DecisionDeny {
		for _, reason := range result.Reasons {
			code := reasonCode(reason)
			if !containsString(codes, code) {
				codes = append(codes, code)
			}
			details = append(details, reason)
		}
	}
	sort.Strings(codes)

	decision := string(result.Decision)
	if len(codes) > 0 {
		decision = string(supplylayer.DecisionDeny)
	}
	admission := &launchAdmission{event: base}
	if decision == string(supplylayer.DecisionDeny) {
		denyEvent := base
		denyEvent.Phase = "launch-denied"
		denyEvent.Decision = decision
		denyEvent.ReasonCodes = codes
		denyEvent.ReasonDetails = details
		a.recordActionJournal(ctx, denyEvent)
		return admission, launchGateError{Codes: codes, Details: details}
	}

	allowEvent := base
	allowEvent.Phase = "launch-allowed"
	allowEvent.Decision = decision
	a.recordActionJournal(ctx, allowEvent)
	return admission, nil
}

func (a *Architect) recordAgentLaunchOutcome(ctx context.Context, admission *launchAdmission, phase string, err error) {
	if admission == nil {
		return
	}
	event := admission.event
	event.Phase = phase
	event.Decision = "completed"
	if err != nil {
		event.Decision = "failed"
		event.ReasonCodes = []string{"runtime_launch_failed"}
		event.ReasonDetails = []string{err.Error()}
	}
	a.recordActionJournal(ctx, event)
}

func launchRequestFromEnv(world *models.World, agentID, capability string, requestedAt time.Time) (supplylayer.LaunchRequest, []string, []string) {
	var codes []string
	var details []string
	humanID := strings.TrimSpace(os.Getenv("SPWN_SUPPLY_HUMAN_ID"))
	issueURL := strings.TrimSpace(os.Getenv("SPWN_SUPPLY_ISSUE_URL"))
	actionEndpoint := strings.TrimSpace(os.Getenv("SPWN_ACTION_JOURNAL_ENDPOINT"))
	actionTenant := strings.TrimSpace(os.Getenv("SPWN_ACTION_JOURNAL_TENANT_ID"))
	polpDecision := strings.TrimSpace(os.Getenv("SPWN_POLP_DECISION"))
	polpReason := strings.TrimSpace(os.Getenv("SPWN_POLP_REASON"))
	customerDataRaw := strings.TrimSpace(os.Getenv("SPWN_SUPPLY_CUSTOMER_DATA"))

	if humanID == "" {
		codes = append(codes, "missing_human_id")
		details = append(details, "SPWN_SUPPLY_HUMAN_ID is required")
	}
	if issueURL == "" {
		codes = append(codes, "missing_issue_url")
		details = append(details, "SPWN_SUPPLY_ISSUE_URL is required")
	}
	if actionEndpoint == "" || actionTenant == "" {
		codes = append(codes, "missing_action_journal")
		details = append(details, "SPWN_ACTION_JOURNAL_ENDPOINT and SPWN_ACTION_JOURNAL_TENANT_ID are required")
	}
	if polpDecision == "" {
		codes = append(codes, "missing_polp_decision")
		details = append(details, "SPWN_POLP_DECISION is required")
	}

	customerData := false
	if customerDataRaw == "" {
		codes = append(codes, "missing_tos_customer_data")
		details = append(details, "SPWN_SUPPLY_CUSTOMER_DATA must be set to true or false")
	} else if parsed, err := strconv.ParseBool(customerDataRaw); err != nil {
		codes = append(codes, "invalid_tos_customer_data")
		details = append(details, "SPWN_SUPPLY_CUSTOMER_DATA must parse as a boolean")
	} else {
		customerData = parsed
	}

	return supplylayer.LaunchRequest{
		Brand:        runtimeBrand(world.Runtime),
		AgentID:      agentID,
		HumanID:      humanID,
		IssueURL:     issueURL,
		Capabilities: []string{capability},
		ActionJournal: supplylayer.ActionJournalConfig{
			Endpoint: actionEndpoint,
			TenantID: actionTenant,
		},
		PoLPDecision: supplylayer.PoLPDecision{
			Decision:     supplylayer.Decision(polpDecision),
			SubjectID:    strings.TrimSpace(os.Getenv("SPWN_POLP_SUBJECT_ID")),
			ScopeID:      strings.TrimSpace(os.Getenv("SPWN_POLP_SCOPE_ID")),
			CapabilityID: strings.TrimSpace(os.Getenv("SPWN_POLP_CAPABILITY_ID")),
			Reason:       polpReason,
		},
		CustomerData:    customerData,
		RequestedAt:     requestedAt,
		AdapterSpecWant: "1.0",
	}, codes, details
}

func runtimeBrand(runtime string) string {
	if runtime == "" {
		return defaultRuntimeName
	}
	return runtime
}

func agentIDForLaunch(world *models.World, name string) string {
	if world.Agent == name && world.AgentID != "" {
		return world.AgentID
	}
	for _, rec := range world.Agents {
		if rec.Name == name && rec.AgentID != "" {
			return rec.AgentID
		}
	}
	return platform.GenerateAgentID(name)
}

func reasonCode(reason string) string {
	switch {
	case strings.Contains(reason, "action journal sink is required"):
		return "missing_action_journal"
	case strings.Contains(reason, "polp denied"):
		return "polp_denied"
	case strings.Contains(reason, "missing subject, scope, or capability"):
		return "polp_scope_incomplete"
	case strings.Contains(reason, "missing agent id"):
		return "missing_agent_id"
	case strings.Contains(reason, "missing human id"):
		return "missing_human_id"
	case strings.Contains(reason, "missing issue url"):
		return "missing_issue_url"
	case strings.Contains(reason, "tos gate not approved"):
		return "tos_gate_not_approved"
	case strings.Contains(reason, "tos conditional brand cannot receive customer data"):
		return "tos_customer_data_denied"
	case strings.Contains(reason, "capability not allowed"):
		return "capability_not_allowed"
	case strings.Contains(reason, "adapter spec mismatch"):
		return "adapter_spec_mismatch"
	case strings.Contains(reason, "unknown brand"):
		return "unknown_brand"
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(reason)), " ", "_")
	}
}

func containsString(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
