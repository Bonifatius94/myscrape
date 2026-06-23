// Package llm is a thin OpenAI-compatible chat client used by web_research's LLM
// synthesis. Runtime is swappable by base URL alone: Ollama / llama.cpp / LM Studio
// / any OpenAI-compatible endpoint. Only used when synthesis=llm (opt-in, GPU).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Message is one chat turn.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatClient completes a chat conversation. Mockable in tests.
type ChatClient interface {
	Complete(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)
}

// Client talks to an OpenAI-compatible /chat/completions endpoint.
type Client struct {
	hc      *http.Client
	baseURL string
	model   string
	apiKey  string
}

// New builds a Client. baseURL is the OpenAI-compatible root (e.g.
// http://host:11434/v1). An empty apiKey sends no Authorization header.
func New(baseURL, model, apiKey string, timeout time.Duration) *Client {
	return &Client{
		hc:      &http.Client{Timeout: timeout},
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		apiKey:  apiKey,
	}
}

func (c *Client) Complete(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error) {
	payload := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llm http %d", resp.StatusCode)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}
