package codex

import "spwn.sh/packages/runtimes"

// Adapter is the codex umbrella: install recipe + spawn-time
// (credential sync, prelaunch shell, one-shot flag synthesis,
// output parsing) + source→Tree renderer. Fully first-class — same
// surface as claude-code.
var Adapter = runtimes.Adapter{
	Name:            "codex",
	DefaultProvider: "openai",
	Tool:            Tool,
	Render:          Renderer,
	Spawn:           Spawner,
	Spec: runtimes.AdapterSpec{
		SchemaVersion:  runtimes.AdapterSpecV1,
		AgentFamily:    "codex",
		AdapterVersion: "1.0.0",
		RuntimeSurface: runtimes.RuntimeSurfaceMixed,
		LegalEvidence: runtimes.LegalEvidence{
			ToSURL:                  "https://help.openai.com/en/articles/11096431",
			PrivacyURL:              "https://openai.com/policies/api-data-usage-policies/",
			LicenseURL:              "https://github.com/openai/codex",
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
