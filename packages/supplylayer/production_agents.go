package supplylayer

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	AdapterSpecV1                  = "1.0"
	SpawnProductionDispatchAPIPath = "/v1/spawn/dispatch"
)

// AgentRegistration is the production handoff record for an agent type that
// dispatches through Spawn supply instead of a legacy direct-invoke path.
type AgentRegistration struct {
	AgentType              string
	Repository             string
	Brand                  string
	AdapterSpecVersion     string
	SpawnAPIPath           string
	IssueURL               string
	ProductionEnvironment  string
	FirstDispatchAt        time.Time
	Capabilities           []string
	DirectInvokeDeprecated bool
}

// DefaultWave3ProductionAgents returns the first five production agent types
// onboarded after the Wave 2 demo agent.
func DefaultWave3ProductionAgents() []AgentRegistration {
	firstDispatch := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	return []AgentRegistration{
		productionAgent("narrator-ai", "GridLtd-ProductDev/narrator-ai-py", firstDispatch),
		productionAgent("GridFastAgent", "OpenAgentSystem/agent-services", firstDispatch.Add(10*time.Minute)),
		productionAgent("hire-app", "OpenAgentSystem/hire-app", firstDispatch.Add(20*time.Minute)),
		productionAgent("closure-bot", "OpenAgentSystem/closure-bot", firstDispatch.Add(30*time.Minute)),
		productionAgent("agency-agents", "OpenAgentSystem/agency-agents", firstDispatch.Add(40*time.Minute)),
	}
}

// DefaultWave3Registry adds the production agent brands to the Wave 1 runtime
// registry without changing the Spawn core dispatch implementation.
func DefaultWave3Registry() Registry {
	registry := DefaultWave1Registry()
	for _, agent := range DefaultWave3ProductionAgents() {
		registry.Brands[normalize(agent.AgentType)] = BrandPolicy{
			Brand:               agent.AgentType,
			AdapterSpecVersion:  agent.AdapterSpecVersion,
			AllowedCapabilities: append([]string{}, agent.Capabilities...),
			ToS:                 ToSAllowed,
		}
	}
	return registry
}

// ValidateProductionAgentRegistration checks the handoff requirements that must
// be true before production dispatch is considered onboarded.
func ValidateProductionAgentRegistration(agent AgentRegistration) error {
	var reasons []string
	if strings.TrimSpace(agent.AgentType) == "" {
		reasons = append(reasons, "missing agent type")
	}
	if strings.TrimSpace(agent.Repository) == "" {
		reasons = append(reasons, "missing repository")
	}
	if agent.AdapterSpecVersion != AdapterSpecV1 {
		reasons = append(reasons, "adapter spec must be 1.0")
	}
	if agent.SpawnAPIPath != SpawnProductionDispatchAPIPath {
		reasons = append(reasons, "spawn dispatch api path mismatch")
	}
	if err := validateIssueURL(agent.IssueURL); err != nil {
		reasons = append(reasons, err.Error())
	}
	if agent.ProductionEnvironment != "production" {
		reasons = append(reasons, "production environment is required")
	}
	if agent.FirstDispatchAt.IsZero() {
		reasons = append(reasons, "missing first production dispatch time")
	}
	if len(agent.Capabilities) == 0 {
		reasons = append(reasons, "capabilities are required")
	}
	if !agent.DirectInvokeDeprecated {
		reasons = append(reasons, "legacy direct-invoke must be deprecated")
	}
	if len(reasons) > 0 {
		return errors.New(strings.Join(reasons, "; "))
	}
	return nil
}

func productionAgent(agentType, repository string, firstDispatchAt time.Time) AgentRegistration {
	return AgentRegistration{
		AgentType:             agentType,
		Repository:            repository,
		Brand:                 agentType,
		AdapterSpecVersion:    AdapterSpecV1,
		SpawnAPIPath:          SpawnProductionDispatchAPIPath,
		IssueURL:              "https://github.com/OpenAgentSystem/spawn/issues/18",
		ProductionEnvironment: "production",
		FirstDispatchAt:       firstDispatchAt,
		Capabilities: []string{
			"agent.spawn",
			"agent.dispatch",
			"action_journal.write",
		},
		DirectInvokeDeprecated: true,
	}
}

func validateIssueURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" {
		return fmt.Errorf("invalid issue url")
	}
	if !strings.Contains(parsed.Path, "/issues/") {
		return fmt.Errorf("issue url must point to a GitHub issue")
	}
	return nil
}
