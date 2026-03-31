package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"

type OpenRouterClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewOpenRouterClient(httpClient *http.Client) (*OpenRouterClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is not set")
	}

	baseURL := strings.TrimSpace(os.Getenv("OPENROUTER_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultOpenRouterBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &OpenRouterClient{
		httpClient: httpClient,
		baseURL:    normalizeOpenRouterBaseURL(baseURL),
		apiKey:     apiKey,
	}, nil
}

func (c *OpenRouterClient) Run(ctx context.Context, prompt PromptResponse, model string) (Decision, string, error) {
	if strings.TrimSpace(model) == "" {
		model = prompt.ModelHint
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": prompt.SystemPrompt},
			{"role": "user", "content": prompt.UserPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0,
		"max_tokens":      256,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Decision{}, "", fmt.Errorf("marshal openrouter request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Decision{}, "", fmt.Errorf("build openrouter request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if referer := strings.TrimSpace(os.Getenv("OPENROUTER_HTTP_REFERER")); referer != "" {
		req.Header.Set("HTTP-Referer", referer)
	}
	if title := strings.TrimSpace(os.Getenv("OPENROUTER_X_TITLE")); title != "" {
		req.Header.Set("X-Title", title)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Decision{}, "", fmt.Errorf("send openrouter request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Decision{}, "", fmt.Errorf("read openrouter response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return Decision{}, string(respBody), fmt.Errorf("openrouter request failed: %s", resp.Status)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return Decision{}, string(respBody), fmt.Errorf("decode openrouter response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return Decision{}, string(respBody), fmt.Errorf("openrouter response contains no choices")
	}

	raw, err := parseOpenRouterMessage(parsed.Choices[0].Message.Content)
	if err != nil {
		return Decision{}, string(respBody), err
	}
	decision, err := ParseDecisionText(raw)
	if err != nil {
		return Decision{}, raw, err
	}
	return decision, raw, nil
}

func parseOpenRouterMessage(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	var chunks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &chunks); err == nil {
		var builder strings.Builder
		for _, chunk := range chunks {
			if chunk.Text != "" {
				builder.WriteString(chunk.Text)
			}
		}
		if builder.Len() > 0 {
			return builder.String(), nil
		}
	}

	return "", fmt.Errorf("openrouter response content is not a supported text payload")
}

func normalizeOpenRouterBaseURL(value string) string {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return defaultOpenRouterBaseURL
	}
	if strings.HasSuffix(value, "/chat/completions") {
		return strings.TrimSuffix(value, "/chat/completions")
	}
	return value
}
