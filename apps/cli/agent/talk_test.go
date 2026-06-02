package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"spwn.sh/packages/architect"
	"spwn.sh/packages/world"
)

func TestExtractSessionID_ClaudeStreamJSON(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			name: "system init carries session_id",
			line: `{"type":"system","subtype":"init","session_id":"abc-123","cwd":"/workspace"}`,
			want: "abc-123",
		},
		{
			name: "assistant message carries session_id",
			line: `{"type":"assistant","session_id":"abc-123","message":{"role":"assistant"}}`,
			want: "abc-123",
		},
		{
			name: "result event carries session_id",
			line: `{"type":"result","session_id":"final-id","subtype":"success"}`,
			want: "final-id",
		},
		{
			name: "non-json line returns empty",
			line: `not json at all`,
			want: "",
		},
		{
			name: "json without session_id returns empty",
			line: `{"type":"tool_use","input":{}}`,
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractSessionID("claude-code", []byte(tc.line)); got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestExtractSessionID_CodexJSONL(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			name: "thread.started carries thread_id",
			line: `{"type":"thread.started","thread_id":"t-9999"}`,
			want: "t-9999",
		},
		{
			name: "item.completed without thread_id returns empty",
			line: `{"type":"item.completed","item":{"type":"text","text":"hi"}}`,
			want: "",
		},
		{
			name: "claude session_id is ignored for codex runtime",
			line: `{"type":"system","session_id":"not-a-thread-id"}`,
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractSessionID("codex", []byte(tc.line)); got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
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

func TestAuthorizeTalkLaunch_AllowsOneShotWithSupplyInputs(t *testing.T) {
	setAllowLaunchEnv(t, "agent.talk")
	arc := architect.New(nil, nil)
	w := &world.World{
		ID:      "world-cli-allow",
		Runtime: "codex",
		Agents:  []world.AgentRecord{{Name: "editor", AgentID: "agent-editor-12345"}},
	}

	if err := authorizeTalkLaunch(context.Background(), arc, w, "editor", "one-shot"); err != nil {
		t.Fatalf("authorize talk: %v", err)
	}
}

func TestAuthorizeTalkLaunch_DeniesBeforeCLIExecWithReasonCodes(t *testing.T) {
	clearLaunchEnv(t)
	arc := architect.New(nil, nil)
	w := &world.World{
		ID:      "world-cli-deny",
		Runtime: "codex",
		Agents:  []world.AgentRecord{{Name: "editor", AgentID: "agent-editor-12345"}},
	}

	err := authorizeTalkLaunch(context.Background(), arc, w, "editor", "one-shot")
	if err == nil {
		t.Fatal("expected gate denial")
	}
	for _, want := range []string{"missing_action_journal", "polp_denied", "missing_tos_customer_data"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("denial missing %q: %v", want, err)
		}
	}
}

func TestFormatExecError_AuthenticationError(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("Error: authentication_error - invalid credentials")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "spwn auth login") {
		t.Errorf("expected mention of 'spwn auth login', got: %s", result.Error())
	}
	if !strings.Contains(result.Error(), "authentication failed") {
		t.Errorf("expected mention of 'authentication failed', got: %s", result.Error())
	}
}

func TestFormatExecError_OAuthExpired(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("OAuth token has expired, please refresh")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "refresh") {
		t.Errorf("expected mention of 'refresh', got: %s", result.Error())
	}
	if !strings.Contains(result.Error(), "expired") {
		t.Errorf("expected mention of 'expired', got: %s", result.Error())
	}
}

func TestFormatExecError_InvalidAPIKey(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("Invalid API key provided")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "spwn auth") {
		t.Errorf("expected mention of 'spwn auth', got: %s", result.Error())
	}
	if !strings.Contains(result.Error(), "invalid API key") {
		t.Errorf("expected mention of 'invalid API key', got: %s", result.Error())
	}
}

func TestFormatExecError_InvalidXAPIKey(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("invalid x-api-key header")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "spwn auth") {
		t.Errorf("expected mention of 'spwn auth', got: %s", result.Error())
	}
}

func TestFormatExecError_GenericWithOutput(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("some unexpected error occurred")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "some unexpected error occurred") {
		t.Errorf("expected output in error message, got: %s", result.Error())
	}
	if !strings.Contains(result.Error(), "spwn auth check") {
		t.Errorf("expected hint about 'spwn auth check', got: %s", result.Error())
	}
}

func TestFormatExecError_GenericTruncatesLongOutput(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte(strings.Repeat("x", 600))
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "...") {
		t.Errorf("expected truncated output with '...', got length: %d", len(result.Error()))
	}
}

func TestFormatExecError_EmptyOutput(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	result := formatExecError(err, nil)

	if !strings.Contains(result.Error(), "spwn auth check") {
		t.Errorf("expected hint about 'spwn auth check', got: %s", result.Error())
	}
}

func TestFormatExecError_EmptyByteOutput(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	result := formatExecError(err, []byte(""))

	if !strings.Contains(result.Error(), "spwn auth check") {
		t.Errorf("expected hint about 'spwn auth check', got: %s", result.Error())
	}
}

func TestFormatExecError_NetworkError(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("Could not resolve host api.anthropic.com")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "network error") {
		t.Errorf("expected 'network error', got: %s", result.Error())
	}
}

func TestFormatExecError_RateLimit(t *testing.T) {
	err := fmt.Errorf("exit status 1")
	output := []byte("rate_limit exceeded, try again later")
	result := formatExecError(err, output)

	if !strings.Contains(result.Error(), "rate limited") {
		t.Errorf("expected 'rate limited', got: %s", result.Error())
	}
}
