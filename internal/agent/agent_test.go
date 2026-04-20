package agent

import (
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
