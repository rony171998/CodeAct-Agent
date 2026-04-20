package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type openAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type responsesRequest struct {
	Model        string `json:"model"`
	Instructions string `json:"instructions"`
	Input        string `json:"input"`
}

type responsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newOpenAIClient(model string) (*openAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is required")
	}
	if model == "" {
		model = envOr("CODEACT_MODEL", "gpt-5.4-mini")
	}
	return &openAIClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(envOr("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (c *openAIClient) generateAction(ctx context.Context, prompt string) (string, error) {
	body := responsesRequest{
		Model:        c.model,
		Instructions: systemInstructions(),
		Input:        prompt,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("OpenAI request failed: %s: %s", resp.Status, string(respBody))
	}

	var parsed responsesResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil {
		return "", errors.New(parsed.Error.Message)
	}

	text := parsed.OutputText
	if text == "" {
		text = collectOutputText(parsed)
	}
	code := extractGoCode(text)
	if strings.TrimSpace(code) == "" {
		return "", errors.New("model returned empty action")
	}
	return code, nil
}

func collectOutputText(resp responsesResponse) string {
	var b strings.Builder
	for _, item := range resp.Output {
		for _, content := range item.Content {
			if content.Text != "" {
				b.WriteString(content.Text)
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func extractGoCode(text string) string {
	text = strings.TrimSpace(text)
	start := strings.Index(text, "```")
	if start == -1 {
		return text
	}
	rest := strings.TrimSpace(text[start+3:])
	if strings.HasPrefix(strings.ToLower(rest), "go") {
		rest = strings.TrimSpace(rest[2:])
	}
	end := strings.Index(rest, "```")
	if end == -1 {
		return rest
	}
	return strings.TrimSpace(rest[:end]) + "\n"
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
