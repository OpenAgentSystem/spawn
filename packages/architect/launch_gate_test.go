package architect

import (
	"context"
	"strings"
	"testing"

	"spwn.sh/packages/world/models"
)

type mockActionJournalSink struct {
	events []ActionJournalEnvelope
}

func (m *mockActionJournalSink) RecordAction(_ context.Context, event ActionJournalEnvelope) error {
	m.events = append(m.events, event)
	return nil
}

func setAllowLaunchEnv(t *testing.T, capability string) {
	t.Helper()
	t.Setenv("SPWN_SUPPLY_HUMAN_ID", "user:paul")
	t.Setenv("SPWN_SUPPLY_ISSUE_URL", "https://github.com/Gridltd-DevOps/architecture-decisions/issues/29")
	t.Setenv("SPWN_ACTION_JOURNAL_ENDPOINT", "https://action-journal.internal/events")
	t.Setenv("SPWN_ACTION_JOURNAL_TENANT_ID", "gridltd")
	t.Setenv("SPWN_POLP_DECISION", "allow")
	t.Setenv("SPWN_POLP_SUBJECT_ID", "user:paul")
	t.Setenv("SPWN_POLP_SCOPE_ID", "issue:29")
	t.Setenv("SPWN_POLP_CAPABILITY_ID", capability)
	t.Setenv("SPWN_SUPPLY_CUSTOMER_DATA", "false")
}

func clearLaunchEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"SPWN_SUPPLY_HUMAN_ID",
		"SPWN_SUPPLY_ISSUE_URL",
		"SPWN_ACTION_JOURNAL_ENDPOINT",
		"SPWN_ACTION_JOURNAL_TENANT_ID",
		"SPWN_POLP_DECISION",
		"SPWN_POLP_SUBJECT_ID",
		"SPWN_POLP_SCOPE_ID",
		"SPWN_POLP_CAPABILITY_ID",
		"SPWN_POLP_REASON",
		"SPWN_SUPPLY_CUSTOMER_DATA",
	} {
		t.Setenv(name, "")
	}
}

func TestAuthorizeAgentLaunch_AllowsWithJournalAndPoLPInputs(t *testing.T) {
	setAllowLaunchEnv(t, launchCapabilityTalk)
	mb := newMockBackend()
	arch, _ := newTestArchitect(t, mb)
	sink := &mockActionJournalSink{}
	arch.SetActionJournalSink(sink)

	world := &models.World{
		ID:      "world-gate-allow",
		Runtime: "codex",
		Agents:  []models.AgentRecord{{Name: "editor", AgentID: "agent-editor-12345"}},
	}

	if _, err := arch.authorizeAgentLaunch(context.Background(), world, "editor", launchCapabilityTalk, "interactive"); err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if len(sink.events) != 2 {
		t.Fatalf("journal events = %d, want 2", len(sink.events))
	}
	if sink.events[0].Phase != "pre-launch" || sink.events[1].Phase != "launch-allowed" {
		t.Fatalf("unexpected phases: %q, %q", sink.events[0].Phase, sink.events[1].Phase)
	}
	if sink.events[1].AgentID != "agent-editor-12345" {
		t.Errorf("agent id = %q", sink.events[1].AgentID)
	}
}

func TestAuthorizeAgentLaunch_DeniesMissingAuditPoLPAndToSInputs(t *testing.T) {
	clearLaunchEnv(t)
	mb := newMockBackend()
	arch, _ := newTestArchitect(t, mb)
	sink := &mockActionJournalSink{}
	arch.SetActionJournalSink(sink)

	world := &models.World{
		ID:      "world-gate-deny",
		Runtime: "codex",
		Agents:  []models.AgentRecord{{Name: "editor", AgentID: "agent-editor-12345"}},
	}

	_, err := arch.authorizeAgentLaunch(context.Background(), world, "editor", launchCapabilityTalk, "interactive")
	if err == nil {
		t.Fatal("expected denial")
	}
	msg := err.Error()
	for _, want := range []string{"missing_action_journal", "polp_denied", "missing_tos_customer_data"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("denial missing %q: %v", want, err)
		}
	}
	if len(sink.events) != 2 {
		t.Fatalf("journal events = %d, want 2", len(sink.events))
	}
	if sink.events[1].Phase != "launch-denied" {
		t.Fatalf("second event phase = %q", sink.events[1].Phase)
	}
	if !containsString(sink.events[1].ReasonCodes, "missing_action_journal") {
		t.Errorf("journal reason codes = %v", sink.events[1].ReasonCodes)
	}
}

func TestSpawnAgent_DenyHappensBeforeDockerExec(t *testing.T) {
	t.Setenv("SPWN_HOME", t.TempDir())
	clearLaunchEnv(t)
	mb := newMockBackend()
	arch, _ := newTestArchitect(t, mb)
	w := models.World{
		ID:          "world-before-exec",
		Config:      "gate-test",
		Runtime:     "codex",
		ContainerID: "mock-gate-1",
		Status:      models.StatusRunning,
		Agents:      []models.AgentRecord{{Name: "editor", AgentID: "agent-editor-12345"}},
	}
	seedWorld(mb, w)

	err := arch.SpawnAgent(context.Background(), "world-before-exec", "editor")
	if err == nil {
		t.Fatal("expected gate denial")
	}
	if len(mb.execCalls) != 0 {
		t.Fatalf("exec calls = %d, want 0 before gate allow", len(mb.execCalls))
	}
}

func TestSpawnAgent_AllowPreservesRuntimeExec(t *testing.T) {
	t.Setenv("SPWN_HOME", t.TempDir())
	setAllowLaunchEnv(t, launchCapabilityTalk)
	mb := newMockBackend()
	arch, _ := newTestArchitect(t, mb)
	sink := &mockActionJournalSink{}
	arch.SetActionJournalSink(sink)
	w := models.World{
		ID:          "world-runtime-exec",
		Config:      "gate-test",
		Runtime:     "codex",
		ContainerID: "mock-gate-2",
		Status:      models.StatusRunning,
		Agents:      []models.AgentRecord{{Name: "editor", AgentID: "agent-editor-12345"}},
	}
	seedWorld(mb, w)

	if err := arch.SpawnAgent(context.Background(), "world-runtime-exec", "editor"); err != nil {
		t.Fatalf("spawn agent: %v", err)
	}
	if len(mb.execCalls) != 1 {
		t.Fatalf("exec calls = %d, want 1", len(mb.execCalls))
	}
	if !cmdContains(mb.execCalls[0].cfg.Cmd, "codex") {
		t.Fatalf("runtime command = %v", mb.execCalls[0].cfg.Cmd)
	}
	if sink.events[len(sink.events)-1].Phase != "post-launch" {
		t.Fatalf("last journal phase = %q", sink.events[len(sink.events)-1].Phase)
	}
}
