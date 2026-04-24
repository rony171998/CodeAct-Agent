package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractGoCode(t *testing.T) {
	got := extractGoCode("```go\npackage main\nfunc main(){}\n```")
	if !strings.Contains(got, "package main") {
		t.Fatalf("expected Go code, got %q", got)
	}
}

func TestBuildPromptIncludesExecutionContract(t *testing.T) {
	prompt := buildPrompt("analyze", "workspace", "input.log", "report.md", "INFO ok", "")
	for _, want := range []string{"CODEACT_INPUT_FILE", "CODEACT_REPORT_PATH", "INFO ok"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}

func TestCheckOpenAIAvailabilityMissingKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	status := CheckOpenAIAvailability(context.Background(), "gpt-5.4-mini")
	if status.Available {
		t.Fatal("expected unavailable status")
	}
	if status.State != "missing_key" {
		t.Fatalf("expected missing_key, got %q", status.State)
	}
}

func TestCheckOpenAIAvailabilityReachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_BASE_URL", server.URL)
	status := CheckOpenAIAvailability(context.Background(), "gpt-5.4-mini")
	if !status.Available {
		t.Fatalf("expected available status, got %+v", status)
	}
	if status.State != "available" {
		t.Fatalf("expected available state, got %q", status.State)
	}
}
