package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"codeact-agent/internal/agent"
)

func main() {
	goal := flag.String("goal", "", "task for the agent")
	input := flag.String("input", "sample.log", "workspace input file")
	workspace := flag.String("workspace", "workspace", "workspace directory")
	model := flag.String("model", os.Getenv("CODEACT_MODEL"), "OpenAI model")
	maxSteps := flag.Int("max-steps", 2, "maximum correction attempts")
	flag.Parse()

	if *goal == "" {
		fmt.Fprintln(os.Stderr, "missing -goal")
		os.Exit(2)
	}

	runRoot := filepath.Join(".codeact", "runs")
	result, err := agent.Run(context.Background(), agent.Config{
		Goal:      *goal,
		InputFile: *input,
		Workspace: *workspace,
		RunRoot:   runRoot,
		Model:     *model,
		MaxSteps:  *maxSteps,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent failed: %v\n", err)
	}

	encoded, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, "encode result: %v\n", jsonErr)
		os.Exit(1)
	}
	fmt.Println(string(encoded))

	if err != nil {
		os.Exit(1)
	}
}
