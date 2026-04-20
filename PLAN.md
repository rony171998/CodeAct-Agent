# CodeAct Agent Web Demo Plan

## Summary

Build a Go + React demo for a CodeAct-style data analyst. The web app shows every step: user goal, model prompt, generated Go action, execution output, and final report.

## Implementation

- Go backend serves the API and the built React app.
- React frontend lets the user choose `sample.log` or `sales.csv`, enter a goal, run the agent, and inspect each artifact.
- OpenAI is required through `OPENAI_API_KEY`.
- The agent uses the OpenAI Responses API to generate one Go program.
- The generated program runs locally with `go run`.
- The generated program receives:
  - `CODEACT_WORKSPACE`
  - `CODEACT_INPUT_FILE`
  - `CODEACT_REPORT_PATH`
- If execution fails, the agent captures output and asks the model for a corrected action.

## API

- `POST /api/runs`
  - request: `{ "goal": "...", "inputFile": "sample.log" }`
  - response: run result with generated code, prompt, output, report, and status
- `GET /api/runs/{id}`
  - response: saved run result

## Tests

- `go test ./...`
- `go run ./cmd/codeact -goal "analyze sample log"`
- `npm install` and `npm run build` inside `web`
- `go run ./cmd/server`, open `http://localhost:8080`
- Missing `OPENAI_API_KEY` must show a clear error.
