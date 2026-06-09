// Package anthropic is a thin raw-HTTP client for the Claude Messages API,
// mirroring the house style of the Voyage/OpenAI embedding clients (no SDK).
// It exposes two primitives the hybrid search uses: forced tool-use (for
// structured spec extraction) and plain text completion (for match
// explanations).
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	messagesURL      = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
)

// Client talks to the Claude Messages API. A zero/empty apiKey makes every call
// return an error, so callers should treat the client as optional and degrade
// gracefully (skip extraction / explanations) when no key is configured.
type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Enabled reports whether an API key is configured. Used by the service to
// decide between the live pipeline and a degraded path.
func (c *Client) Enabled() bool { return c.apiKey != "" }

// Tool describes a single tool Claude may call. InputSchema is a JSON Schema
// object; with ToolUse we force this tool so the response is guaranteed to be
// one tool_use block whose `input` matches the schema.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type messageRequest struct {
	Model      string         `json:"model"`
	MaxTokens  int            `json:"max_tokens"`
	System     string         `json:"system,omitempty"`
	Messages   []message      `json:"messages"`
	Tools      []Tool         `json:"tools,omitempty"`
	ToolChoice map[string]any `json:"tool_choice,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messageResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
}

// ToolUse forces Claude to call `tool` and returns the decoded tool input.
func (c *Client) ToolUse(ctx context.Context, system, user string, tool Tool) (map[string]any, error) {
	resp, err := c.do(ctx, messageRequest{
		Model:      c.model,
		MaxTokens:  1024,
		System:     system,
		Messages:   []message{{Role: "user", Content: user}},
		Tools:      []Tool{tool},
		ToolChoice: map[string]any{"type": "tool", "name": tool.Name},
	})
	if err != nil {
		return nil, err
	}
	for _, block := range resp.Content {
		if block.Type == "tool_use" && len(block.Input) > 0 {
			var out map[string]any
			if err := json.Unmarshal(block.Input, &out); err != nil {
				return nil, fmt.Errorf("anthropic: decode tool input: %w", err)
			}
			return out, nil
		}
	}
	return nil, errors.New("anthropic: no tool_use block in response")
}

// Text runs a plain completion and returns the concatenated text blocks.
func (c *Client) Text(ctx context.Context, system, user string, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		maxTokens = 256
	}
	resp, err := c.do(ctx, messageRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []message{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	for _, block := range resp.Content {
		if block.Type == "text" {
			b.WriteString(block.Text)
		}
	}
	return b.String(), nil
}

func (c *Client) do(ctx context.Context, body messageRequest) (*messageResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("anthropic: missing API key")
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, messagesURL, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("anthropic: build request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic: %s: %s", resp.Status, msg)
	}

	var out messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("anthropic: decode: %w", err)
	}
	return &out, nil
}
