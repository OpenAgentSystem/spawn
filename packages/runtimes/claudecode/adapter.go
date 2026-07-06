package claudecode

import "spwn.sh/packages/runtimes"

// Adapter is the claude-code umbrella: install recipe + render + spawn.
// All three facets are implemented, so the Adapter fills every field.
var Adapter = runtimes.Adapter{
	Name:            "claude-code",
	DefaultProvider: "anthropic",
	Tool:            Tool,
	Render:          Renderer,
	Spawn:           Spawner,
	Spec: runtimes.AdapterSpec{
		SchemaVersion:  runtimes.AdapterSpecV1,
		AgentFamily:    "claude-code",
		AdapterVersion: "1.0.0",
		RuntimeSurface: runtimes.RuntimeSurfaceCLI,
		LegalEvidence: runtimes.LegalEvidence{
			ToSURL:                  "https://code.claude.com/docs/en/legal-and-compliance",
			PrivacyURL:              "https://docs.claude.com/en/docs/claude-code/data-usage",
			ReviewState:             runtimes.ToSReviewConditional,
			IntendedUses:            []runtimes.IntendedUse{runtimes.IntendedUseInternalRnD, runtimes.IntendedUseControlledPoC},
			CustomerDataAllowed:     false,
			SourceCodeAllowed:       true,
			ProductionLogAllowed:    false,
			TrainingOptOutAvailable: true,
		},
	},
}

func init() { runtimes.Register(Adapter) }
