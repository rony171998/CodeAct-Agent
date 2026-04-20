package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codeact-agent/internal/agent"
)

type server struct {
	root      string
	workspace string
	runRoot   string
	model     string
	mu        sync.RWMutex
	runs      map[string]agent.RunResult
}

type runRequest struct {
	Goal      string `json:"goal"`
	InputFile string `json:"inputFile"`
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	s := &server{
		root:      root,
		workspace: filepath.Join(root, "workspace"),
		runRoot:   filepath.Join(root, ".codeact", "runs"),
		model:     envOr("CODEACT_MODEL", "gpt-5.4-mini"),
		runs:      map[string]agent.RunResult{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/runs", s.handleRuns)
	mux.HandleFunc("/api/runs/", s.handleRunByID)
	mux.HandleFunc("/", s.handleStatic)

	addr := envOr("CODEACT_ADDR", ":8080")
	log.Printf("CodeAct Agent listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Goal = strings.TrimSpace(req.Goal)
	if req.Goal == "" {
		writeError(w, http.StatusBadRequest, "goal is required")
		return
	}
	if req.InputFile == "" {
		req.InputFile = "sample.log"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	result, err := agent.Run(ctx, agent.Config{
		Goal:      req.Goal,
		InputFile: req.InputFile,
		Workspace: s.workspace,
		RunRoot:   s.runRoot,
		Model:     s.model,
		MaxSteps:  2,
	})
	if result.ID != "" {
		s.save(result)
	}
	if err != nil {
		writeJSON(w, http.StatusBadGateway, result)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *server) handleRunByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}

	result, err := s.load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *server) save(result agent.RunResult) {
	s.mu.Lock()
	s.runs[result.ID] = result
	s.mu.Unlock()

	runDir := filepath.Join(s.runRoot, result.ID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(runDir, "result.json"), data, 0o644)
}

func (s *server) load(id string) (agent.RunResult, error) {
	s.mu.RLock()
	result, ok := s.runs[id]
	s.mu.RUnlock()
	if ok {
		return result, nil
	}

	data, err := os.ReadFile(filepath.Join(s.runRoot, id, "result.json"))
	if err != nil {
		return agent.RunResult{}, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return agent.RunResult{}, err
	}
	return result, nil
}

func (s *server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	dist := filepath.Join(s.root, "web", "dist")
	index := filepath.Join(dist, "index.html")
	if _, err := os.Stat(index); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, "<h1>Frontend not built</h1><p>Run <code>cd web && npm install && npm run build</code>, then restart the server.</p>")
		return
	}

	path := filepath.Clean(r.URL.Path)
	if path == "." || path == string(filepath.Separator) {
		http.ServeFile(w, r, index)
		return
	}

	file := filepath.Join(dist, strings.TrimPrefix(path, string(filepath.Separator)))
	if _, err := os.Stat(file); err == nil {
		http.ServeFile(w, r, file)
		return
	}
	http.ServeFile(w, r, index)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

var errNotFound = errors.New("not found")
