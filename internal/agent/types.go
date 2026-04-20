package agent

import "time"

type Config struct {
	Goal      string
	InputFile string
	Workspace string
	RunRoot   string
	Model     string
	MaxSteps  int
}

type RunResult struct {
	ID         string    `json:"id"`
	Goal       string    `json:"goal"`
	InputFile  string    `json:"inputFile"`
	Status     string    `json:"status"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"createdAt"`
	ReportPath string    `json:"reportPath"`
	ReportText string    `json:"reportText"`
	Error      string    `json:"error,omitempty"`
	Steps      []Step    `json:"steps"`
}

type Step struct {
	Number     int    `json:"number"`
	Prompt     string `json:"prompt"`
	ActionCode string `json:"actionCode"`
	ActionPath string `json:"actionPath"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
}
