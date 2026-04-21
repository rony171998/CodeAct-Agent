package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Run(ctx context.Context, cfg Config) (RunResult, error) {
	if cfg.Goal == "" {
		return RunResult{}, errors.New("goal is required")
	}
	if cfg.InputFile == "" {
		cfg.InputFile = "sample.log"
	}
	if cfg.MaxSteps < 1 {
		cfg.MaxSteps = 2
	}
	if cfg.Model == "" {
		cfg.Model = envOr("CODEACT_MODEL", "gpt-5.4-mini")
	}

	workspace, err := filepath.Abs(cfg.Workspace)
	if err != nil {
		return RunResult{}, err
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return RunResult{}, err
	}

	inputPath, err := safeInputPath(workspace, cfg.InputFile)
	if err != nil {
		return RunResult{}, err
	}
	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return RunResult{}, fmt.Errorf("read input file: %w", err)
	}

	id := newRunID()
	if cfg.RunRoot == "" {
		cfg.RunRoot = filepath.Join(".codeact", "runs")
	}
	runDir := filepath.Join(cfg.RunRoot, id)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return RunResult{}, err
	}

	reportPath := filepath.Join(workspace, "report-"+id+".md")
	result := RunResult{
		ID:         id,
		Goal:       cfg.Goal,
		InputFile:  filepath.Base(inputPath),
		Status:     "running",
		Model:      cfg.Model,
		CreatedAt:  time.Now().UTC(),
		ReportPath: reportPath,
	}

	client, err := newOpenAIClient(cfg.Model)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	var feedback string
	for stepNum := 1; stepNum <= cfg.MaxSteps; stepNum++ {
		prompt := buildPrompt(cfg.Goal, workspace, inputPath, reportPath, preview(inputBytes), feedback)
		code, err := client.generateAction(ctx, prompt)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result, err
		}

		actionPath := filepath.Join(runDir, fmt.Sprintf("action-%d.go", stepNum))
		if err := os.WriteFile(actionPath, []byte(code), 0o644); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result, err
		}

		output, runErr := runGoAction(ctx, actionPath, workspace, inputPath, reportPath)
		step := Step{
			Number:     stepNum,
			Prompt:     prompt,
			ActionCode: code,
			ActionPath: actionPath,
			Output:     strings.TrimSpace(output),
		}
		if runErr != nil {
			step.Error = runErr.Error()
		}
		result.Steps = append(result.Steps, step)

		if runErr == nil {
			report, err := os.ReadFile(reportPath)
			if err == nil && len(strings.TrimSpace(string(report))) > 0 {
				result.ReportText = string(report)
				result.Status = "completed"
				return result, nil
			}
			runErr = errors.New("action completed but did not write report")
		}

		feedback = fmt.Sprintf("Command output:\n%s\n\nError:\n%s", output, runErr.Error())
	}

	result.Status = "failed"
	result.Error = "all generated actions failed"
	return result, errors.New(result.Error)
}

func runGoAction(ctx context.Context, actionPath, workspace, inputPath, reportPath string) (string, error) {
	goBin, err := goCommand()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, actionTimeout())
	defer cancel()

	env := append(os.Environ(),
		"CODEACT_WORKSPACE="+workspace,
		"CODEACT_INPUT_FILE="+inputPath,
		"CODEACT_REPORT_PATH="+reportPath,
	)
	env = withGoCache(env)

	var out bytes.Buffer
	binaryPath := filepath.Join(filepath.Dir(actionPath), actionBinaryName(actionPath))

	build := exec.CommandContext(ctx, goBin, "build", "-o", binaryPath, actionPath)
	build.Env = env
	build.Stdout = &out
	build.Stderr = &out
	if err := build.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out.String(), ctx.Err()
		}
		return out.String(), err
	}

	run := exec.CommandContext(ctx, binaryPath)
	run.Env = env
	run.Stdout = &out
	run.Stderr = &out
	err = run.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return out.String(), ctx.Err()
	}
	return out.String(), err
}

func actionTimeout() time.Duration {
	seconds, err := strconv.Atoi(envOr("CODEACT_ACTION_TIMEOUT_SECONDS", "120"))
	if err != nil || seconds < 10 {
		seconds = 120
	}
	return time.Duration(seconds) * time.Second
}

func actionBinaryName(actionPath string) string {
	name := strings.TrimSuffix(filepath.Base(actionPath), filepath.Ext(actionPath))
	if filepath.Separator == '\\' {
		return name + ".exe"
	}
	return name
}

func withGoCache(env []string) []string {
	for _, item := range env {
		if strings.HasPrefix(item, "GOCACHE=") {
			return env
		}
	}
	cacheDir := envOr("CODEACT_GO_CACHE", filepath.Join(os.TempDir(), "codeact-go-build-cache"))
	_ = os.MkdirAll(cacheDir, 0o755)
	return append(env, "GOCACHE="+cacheDir)
}

func goCommand() (string, error) {
	if path, err := exec.LookPath("go"); err == nil {
		return path, nil
	}
	candidates := []string{
		`C:\Program Files\Go\bin\go.exe`,
		`C:\Program Files (x86)\Go\bin\go.exe`,
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("go executable not found")
}

func safeInputPath(workspace, inputFile string) (string, error) {
	if filepath.Base(inputFile) != inputFile {
		return "", errors.New("inputFile must be a file name in workspace")
	}
	switch filepath.Ext(inputFile) {
	case ".log", ".csv", ".txt", ".json":
	default:
		return "", errors.New("unsupported input file type")
	}
	path := filepath.Join(workspace, inputFile)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func preview(data []byte) string {
	text := string(data)
	if len(text) > 4000 {
		return text[:4000] + "\n... truncated ..."
	}
	return text
}

func newRunID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().Format("20060102-150405")
	}
	return time.Now().Format("20060102-150405") + "-" + hex.EncodeToString(b[:])
}
