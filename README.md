# CodeAct Agent

CodeAct Agent is a Go + React demo of the code-as-action pattern.

Instead of calling fixed tools, the agent asks a model to write a small Go program, runs that program, captures the result, and retries with execution feedback if needed.

## What it does

This project is a data analyst agent for small files.

You choose a sample file, write a goal, and the app shows:

1. User goal
2. Prompt sent to the model
3. Generated Go action
4. Execution output
5. Final report

## Install

Install Go:

```powershell
winget install --id GoLang.Go -e
```

Install frontend dependencies:

```powershell
cd D:\rony\CodeAct-Agent\web
npm install
npm run build
```

## Configure OpenAI

PowerShell:

```powershell
$env:OPENAI_API_KEY="your_key"
$env:CODEACT_MODEL="gpt-5.4-mini"
```

Optional:

```powershell
$env:OPENAI_BASE_URL="https://api.openai.com/v1"
```

## Run web demo

```powershell
cd D:\rony\CodeAct-Agent
go run ./cmd/server
```

Open:

```text
http://localhost:8080
```

## Run CLI demo

```powershell
cd D:\rony\CodeAct-Agent
go run ./cmd/codeact -goal "analyze sample log"
```

CSV example:

```powershell
go run ./cmd/codeact -goal "summarize sales by category" -input sales.csv
```

## Deploy

Recommended split:

- Vercel hosts the Vite frontend from `web`.
- Render hosts the Go backend from the repository root.

### Render backend

Use the included `render.yaml` Blueprint.

Required Render environment variables:

```text
OPENAI_API_KEY
CODEACT_ALLOWED_ORIGIN
```

Optional:

```text
CODEACT_MODEL=gpt-5.4-mini
```

The backend uses Docker so the Go toolchain is available at runtime for generated actions.

### Vercel frontend

Deploy either the repository root with `vercel.json`, or the `web` directory directly.

Required Vercel environment variable:

```text
VITE_API_BASE_URL=https://your-render-service.onrender.com
```

Do not commit real API keys or environment values.

## Interview script

Use this short explanation:

```text
This is a CodeAct agent. It receives a goal, builds a prompt, asks the model for a Go program, saves that program as an action, executes it with go run, and shows the output. If the generated code fails, the agent sends the error back to the model and asks for a corrected action. The action writes a markdown report inside the workspace.
```

Demo steps:

1. Open the web app.
2. Select `sample.log`.
3. Enter `analyze errors and warnings in this log`.
4. Click Run.
5. Show the prompt.
6. Show the generated Go code.
7. Show the execution output.
8. Show the final report.
9. Explain that the generated code is the action.

## Project structure

```text
cmd/codeact      CLI entrypoint
cmd/server       Web/API server
internal/agent   CodeAct loop and OpenAI provider
web              React frontend
workspace        Sample files and generated reports
```

## Security

Generated code runs locally. Only use a controlled workspace.

Do not commit `.env`, API keys, generated runs, or generated reports.
