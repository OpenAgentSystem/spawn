package agentplatform

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"spwn.sh/packages/supplylayer"
)

const adapterSpecV1 = "1.0"

// AgentManifest is the register-time contract for Adapter Spec v1 agents.
type AgentManifest struct {
	AdapterSpecVersion string   `json:"adapter_spec_version"`
	AgentID            string   `json:"agent_id"`
	Name               string   `json:"name"`
	Brand              string   `json:"brand"`
	Capabilities       []string `json:"capabilities"`
	DispatchURL        string   `json:"dispatch_url"`
}

// RegisterRequest is POST /agents/register.
type RegisterRequest struct {
	Manifest AgentManifest `json:"manifest"`
}

// RegisterResponse is returned after a manifest is accepted.
type RegisterResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

// DispatchRequest is POST /dispatch.
type DispatchRequest struct {
	AgentID       string                          `json:"agent_id"`
	HumanID       string                          `json:"human_id"`
	IssueURL      string                          `json:"issue_url"`
	Capabilities  []string                        `json:"capabilities,omitempty"`
	ActionJournal supplylayer.ActionJournalConfig `json:"action_journal"`
	PoLPDecision  supplylayer.PoLPDecision        `json:"polp"`
	CustomerData  bool                            `json:"customer_data,omitempty"`
	Payload       json.RawMessage                 `json:"payload"`
}

// DispatchResponse is returned after a task is accepted or denied.
type DispatchResponse struct {
	TaskID   string   `json:"task_id"`
	Status   string   `json:"status"`
	Decision string   `json:"decision"`
	Reasons  []string `json:"reasons,omitempty"`
}

// TaskStatus is returned by GET /tasks/:id/status.
type TaskStatus struct {
	TaskID      string          `json:"task_id"`
	AgentID     string          `json:"agent_id"`
	Status      string          `json:"status"`
	Decision    string          `json:"decision"`
	Reasons     []string        `json:"reasons,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type taskRecord struct {
	TaskStatus
}

type dispatchEnvelope struct {
	TaskID  string          `json:"task_id"`
	AgentID string          `json:"agent_id"`
	Payload json.RawMessage `json:"payload"`
}

type agentResult struct {
	Status string          `json:"status"`
	Result json.RawMessage `json:"result,omitempty"`
}

// Server is an in-memory, auth-stubbed implementation of the W3-04 API.
type Server struct {
	mu       sync.RWMutex
	registry supplylayer.Registry
	agents   map[string]AgentManifest
	tasks    map[string]taskRecord
	client   *http.Client
	now      func() time.Time
	newID    func() (string, error)
}

// NewServer returns a W3-04 server using the Wave 1 supply-layer registry.
func NewServer() *Server {
	return &Server{
		registry: supplylayer.DefaultWave1Registry(),
		agents:   map[string]AgentManifest{},
		tasks:    map[string]taskRecord{},
		client:   &http.Client{Timeout: 5 * time.Second},
		now:      time.Now,
		newID:    randomTaskID,
	}
}

// Handler returns the HTTP router for the W3-04 endpoints.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents/register", s.handleRegister)
	mux.HandleFunc("/dispatch", s.handleDispatch)
	mux.HandleFunc("/tasks/", s.handleTaskStatus)
	return mux
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req RegisterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	manifest := normalizeManifest(req.Manifest)
	if err := validateManifest(manifest); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	s.agents[manifest.AgentID] = manifest
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, RegisterResponse{AgentID: manifest.AgentID, Status: "registered"})
}

func (s *Server) handleDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req DispatchRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.RLock()
	manifest, ok := s.agents[strings.TrimSpace(req.AgentID)]
	s.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "agent is not registered")
		return
	}

	caps := req.Capabilities
	if len(caps) == 0 {
		caps = manifest.Capabilities
	}
	gate := s.registry.Evaluate(supplylayer.LaunchRequest{
		Brand:           manifest.Brand,
		AgentID:         manifest.AgentID,
		HumanID:         strings.TrimSpace(req.HumanID),
		IssueURL:        strings.TrimSpace(req.IssueURL),
		Capabilities:    caps,
		ActionJournal:   req.ActionJournal,
		PoLPDecision:    req.PoLPDecision,
		CustomerData:    req.CustomerData,
		RequestedAt:     s.now().UTC(),
		AdapterSpecWant: adapterSpecV1,
	})

	taskID, err := s.newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to allocate task id")
		return
	}
	record := taskRecord{TaskStatus: TaskStatus{
		TaskID:    taskID,
		AgentID:   manifest.AgentID,
		Status:    "denied",
		Decision:  string(gate.Decision),
		Reasons:   gate.Reasons,
		CreatedAt: s.now().UTC(),
	}}
	if gate.Decision != supplylayer.DecisionAllow {
		now := s.now().UTC()
		record.CompletedAt = &now
		s.storeTask(record)
		writeJSON(w, http.StatusForbidden, DispatchResponse{
			TaskID:   taskID,
			Status:   record.Status,
			Decision: string(gate.Decision),
			Reasons:  gate.Reasons,
		})
		return
	}

	record.Status = "running"
	record.Decision = string(supplylayer.DecisionAllow)
	s.storeTask(record)

	result, err := s.callAgent(manifest, dispatchEnvelope{
		TaskID:  taskID,
		AgentID: manifest.AgentID,
		Payload: req.Payload,
	})
	now := s.now().UTC()
	record.CompletedAt = &now
	if err != nil {
		record.Status = "failed"
		record.Reasons = []string{err.Error()}
	} else {
		record.Status = normalizeTaskStatus(result.Status)
		record.Result = result.Result
	}
	s.storeTask(record)

	status := http.StatusAccepted
	if record.Status == "failed" {
		status = http.StatusBadGateway
	}
	writeJSON(w, status, DispatchResponse{
		TaskID:   taskID,
		Status:   record.Status,
		Decision: record.Decision,
		Reasons:  record.Reasons,
	})
}

func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	taskID := strings.TrimPrefix(r.URL.Path, "/tasks/")
	taskID = strings.TrimSuffix(taskID, "/status")
	if taskID == "" || r.URL.Path != "/tasks/"+taskID+"/status" {
		writeError(w, http.StatusNotFound, "unknown task status route")
		return
	}
	s.mu.RLock()
	record, ok := s.tasks[taskID]
	s.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, record.TaskStatus)
}

func (s *Server) storeTask(record taskRecord) {
	s.mu.Lock()
	s.tasks[record.TaskID] = record
	s.mu.Unlock()
}

func (s *Server) callAgent(manifest AgentManifest, envelope dispatchEnvelope) (agentResult, error) {
	body, err := json.Marshal(envelope)
	if err != nil {
		return agentResult{}, err
	}
	resp, err := s.client.Post(manifest.DispatchURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return agentResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return agentResult{}, fmt.Errorf("agent dispatch failed: status %d %s", resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	var result agentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return agentResult{}, fmt.Errorf("agent dispatch response invalid: %w", err)
	}
	return result, nil
}

func validateManifest(manifest AgentManifest) error {
	var missing []string
	if manifest.AdapterSpecVersion != adapterSpecV1 {
		return fmt.Errorf("adapter_spec_version must be %s", adapterSpecV1)
	}
	if manifest.AgentID == "" {
		missing = append(missing, "agent_id")
	}
	if manifest.Name == "" {
		missing = append(missing, "name")
	}
	if manifest.Brand == "" {
		missing = append(missing, "brand")
	}
	if len(manifest.Capabilities) == 0 {
		missing = append(missing, "capabilities")
	}
	if manifest.DispatchURL == "" {
		missing = append(missing, "dispatch_url")
	}
	if len(missing) > 0 {
		return fmt.Errorf("manifest missing required fields: %s", strings.Join(missing, ", "))
	}
	u, err := url.Parse(manifest.DispatchURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("dispatch_url must be an absolute http(s) URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("dispatch_url must use http or https")
	}
	return nil
}

func normalizeManifest(manifest AgentManifest) AgentManifest {
	manifest.AdapterSpecVersion = strings.TrimSpace(manifest.AdapterSpecVersion)
	manifest.AgentID = strings.TrimSpace(manifest.AgentID)
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Brand = strings.TrimSpace(manifest.Brand)
	manifest.DispatchURL = strings.TrimSpace(manifest.DispatchURL)
	for i := range manifest.Capabilities {
		manifest.Capabilities[i] = strings.TrimSpace(manifest.Capabilities[i])
	}
	return manifest
}

func normalizeTaskStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "completed", "failed":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "completed"
	}
}

func readJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func randomTaskID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "task_" + hex.EncodeToString(b[:]), nil
}
