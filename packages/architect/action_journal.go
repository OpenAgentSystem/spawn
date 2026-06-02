package architect

import (
	"context"
	"time"
)

// ActionJournalSink is the adapter boundary for Dev #4's append-only
// Action Journal service. Architect owns the envelope shape and call sites;
// storage stays outside this repository.
type ActionJournalSink interface {
	RecordAction(ctx context.Context, event ActionJournalEnvelope) error
}

// ActionJournalEnvelope is emitted around agent launch admission and outcome.
type ActionJournalEnvelope struct {
	SchemaVersion string            `json:"schema_version"`
	Phase         string            `json:"phase"`
	WorldID       string            `json:"world_id"`
	AgentID       string            `json:"agent_id"`
	AgentName     string            `json:"agent_name"`
	Capability    string            `json:"capability"`
	LaunchPath    string            `json:"launch_path"`
	Decision      string            `json:"decision,omitempty"`
	ReasonCodes   []string          `json:"reason_codes,omitempty"`
	ReasonDetails []string          `json:"reason_details,omitempty"`
	RequestedAt   time.Time         `json:"requested_at"`
	RecordedAt    time.Time         `json:"recorded_at"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type noopActionJournalSink struct{}

func (noopActionJournalSink) RecordAction(context.Context, ActionJournalEnvelope) error {
	return nil
}

// SetActionJournalSink swaps the Action Journal adapter. Tests and external
// integrations use this to inject a real or mock sink without changing
// architect launch behavior.
func (a *Architect) SetActionJournalSink(s ActionJournalSink) {
	if s == nil {
		a.actionJournalSink = noopActionJournalSink{}
		return
	}
	a.actionJournalSink = s
}

func (a *Architect) recordActionJournal(ctx context.Context, event ActionJournalEnvelope) {
	if event.SchemaVersion == "" {
		event.SchemaVersion = "oas.action_journal.launch.v1"
	}
	if event.RecordedAt.IsZero() {
		event.RecordedAt = time.Now().UTC()
	}
	sink := a.actionJournalSink
	if sink == nil {
		sink = noopActionJournalSink{}
	}
	_ = sink.RecordAction(ctx, event)
}
