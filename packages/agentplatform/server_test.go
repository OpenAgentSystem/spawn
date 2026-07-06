package agentplatform

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spwn.sh/packages/supplylayer"
)

func TestMockFeishuDemoRegistersReceivesCompletesDispatchViaHTTP(t *testing.T) {
	var received dispatchEnvelope
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("agent received method %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("agent decode dispatch: %v", err)
		}
		writeJSON(w, http.StatusOK, agentResult{
			Status: "completed",
			Result: json.RawMessage(`{"reply":"mock Feishu task routed to platform owner"}`),
		})
	}))
	defer agent.Close()

	api := NewServer()
	api.newID = func() (string, error) { return "task_demo_feishu", nil }
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	register := RegisterRequest{Manifest: AgentManifest{
		AdapterSpecVersion: "1.0",
		AgentID:            "agent:mock-feishu-router",
		Name:               "mock-feishu-router",
		Brand:              "codex",
		Capabilities:       []string{"agent.spawn", "agent.talk"},
		DispatchURL:        agent.URL,
	}}
	resp := postJSON(t, server.URL+"/agents/register", register)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", resp.StatusCode, readBody(t, resp))
	}

	dispatch := DispatchRequest{
		AgentID:      "agent:mock-feishu-router",
		HumanID:      "user:operator",
		IssueURL:     "https://github.com/OpenAgentSystem/spawn/issues/18",
		Capabilities: []string{"agent.spawn", "agent.talk"},
		ActionJournal: supplylayer.ActionJournalConfig{
			Endpoint: "https://audit.example.com/events",
			TenantID: "example",
		},
		PoLPDecision: supplylayer.PoLPDecision{
			Decision:     supplylayer.DecisionAllow,
			SubjectID:    "user:operator",
			ScopeID:      "issue:29",
			CapabilityID: "agent.spawn",
		},
		Payload: json.RawMessage(`{"source":"mock-feishu","text":"route launch demo"}`),
	}
	resp = postJSON(t, server.URL+"/dispatch", dispatch)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dispatch status=%d body=%s", resp.StatusCode, readBody(t, resp))
	}
	var dispatchResp DispatchResponse
	decodeBody(t, resp, &dispatchResp)
	if dispatchResp.TaskID != "task_demo_feishu" || dispatchResp.Status != "completed" {
		t.Fatalf("unexpected dispatch response: %+v", dispatchResp)
	}
	if received.TaskID != "task_demo_feishu" || received.AgentID != "agent:mock-feishu-router" {
		t.Fatalf("agent did not receive expected task: %+v", received)
	}

	statusResp, err := http.Get(server.URL + "/tasks/task_demo_feishu/status")
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("status code=%d body=%s", statusResp.StatusCode, readBody(t, statusResp))
	}
	var status TaskStatus
	decodeBody(t, statusResp, &status)
	if status.Status != "completed" || status.Decision != string(supplylayer.DecisionAllow) {
		t.Fatalf("unexpected task status: %+v", status)
	}
}

func TestDispatchDeniesBeforeCallingAgentWhenPoLPGateFails(t *testing.T) {
	called := false
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer agent.Close()

	api := NewServer()
	api.newID = func() (string, error) { return "task_denied", nil }
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	resp := postJSON(t, server.URL+"/agents/register", RegisterRequest{Manifest: AgentManifest{
		AdapterSpecVersion: "1.0",
		AgentID:            "agent:mock-feishu-router",
		Name:               "mock-feishu-router",
		Brand:              "codex",
		Capabilities:       []string{"agent.spawn"},
		DispatchURL:        agent.URL,
	}})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", resp.StatusCode, readBody(t, resp))
	}

	resp = postJSON(t, server.URL+"/dispatch", DispatchRequest{
		AgentID:  "agent:mock-feishu-router",
		HumanID:  "user:operator",
		IssueURL: "https://github.com/OpenAgentSystem/spawn/issues/18",
		ActionJournal: supplylayer.ActionJournalConfig{
			Endpoint: "https://audit.example.com/events",
			TenantID: "example",
		},
		PoLPDecision: supplylayer.PoLPDecision{
			Decision: supplylayer.DecisionDeny,
			Reason:   "no scoped grant",
		},
		Payload: json.RawMessage(`{"source":"mock-feishu"}`),
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("dispatch status=%d body=%s", resp.StatusCode, readBody(t, resp))
	}
	if called {
		t.Fatal("agent callback should not be called when PoLP gate denies")
	}
}

func TestRegisterRejectsInvalidAdapterSpecManifest(t *testing.T) {
	api := NewServer()
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	resp := postJSON(t, server.URL+"/agents/register", RegisterRequest{Manifest: AgentManifest{
		AdapterSpecVersion: "0.9",
		AgentID:            "agent:bad",
		Name:               "bad",
		Brand:              "codex",
		Capabilities:       []string{"agent.spawn"},
		DispatchURL:        "https://agent.example/dispatch",
	}})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("register status=%d body=%s", resp.StatusCode, readBody(t, resp))
	}
}

func postJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	return resp
}

func decodeBody(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode body: %v", err)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return buf.String()
}
