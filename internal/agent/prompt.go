package agent

import (
	"fmt"
	"path/filepath"
	"strings"
)

func buildPrompt(goal, workspace, inputPath, reportPath, inputPreview, feedback string) string {
	var b strings.Builder
	b.WriteString("Goal:\n")
	b.WriteString(goal)
	b.WriteString("\n\nWorkspace:\n")
	b.WriteString(workspace)
	b.WriteString("\n\nInput file path:\n")
	b.WriteString(inputPath)
	b.WriteString("\n\nReport output path:\n")
	b.WriteString(reportPath)
	b.WriteString("\n\nInput preview:\n")
	b.WriteString(inputPreview)
	b.WriteString("\n\nWrite one complete Go program that:\n")
	b.WriteString("- reads CODEACT_INPUT_FILE\n")
	b.WriteString("- writes a concise markdown report to CODEACT_REPORT_PATH\n")
	b.WriteString("- prints a short execution summary to stdout\n")
	b.WriteString("- uses only the Go standard library\n")
	b.WriteString("- exits non-zero only for real errors\n")
	if feedback != "" {
		b.WriteString("\nPrevious execution feedback:\n")
		b.WriteString(feedback)
	}
	return b.String()
}

func systemInstructions() string {
	return `You are a CodeAct agent for a data analysis demo.
Return one complete Go program only.
No markdown explanation.
No backstory.
The Go program must use environment variables CODEACT_INPUT_FILE and CODEACT_REPORT_PATH.
Use only the Go standard library.
Keep the program deterministic and easy to read.`
}

func describeRun(id, inputFile string) string {
	return fmt.Sprintf("run %s on %s", id, filepath.Base(inputFile))
}
