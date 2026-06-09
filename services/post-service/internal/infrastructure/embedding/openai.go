package embedding

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

// OpenAIClient is the fallback path when Voyage is unreachable. It pins
// dimensions=1024 so the result fits the same vector(1024) Postgres column —
// without this OpenAI would default to 1536 dims and the INSERT would fail.
type OpenAIClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  "text-embedding-3-small",
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OpenAIClient) Name() string { return "openai:text-embedding-3-small" }

type openaiRequest struct {
	Input      []string `json:"input"`
	Model      string   `json:"model"`
	Dimensions int      `json:"dimensions"`
}

type openaiResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *OpenAIClient) Embed(ctx context.Context, texts []string, _ string) ([][]float32, error) {
	if c.apiKey == "" {
		return nil, errors.New("openai: missing API key")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(openaiRequest{Input: texts, Model: c.model, Dimensions: Dim})
	if err != nil {
		return nil, fmt.Errorf("openai: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai: %s: %s", resp.Status, raw)
	}

	var out openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("openai: decode: %w", err)
	}
	if len(out.Data) != len(texts) {
		return nil, fmt.Errorf("openai: expected %d embeddings, got %d", len(texts), len(out.Data))
	}

	result := make([][]float32, len(texts))
	for _, d := range out.Data {
		if d.Index < 0 || d.Index >= len(texts) {
			return nil, fmt.Errorf("openai: out-of-range index %d", d.Index)
		}
		if len(d.Embedding) != Dim {
			return nil, fmt.Errorf("openai: expected %d-dim vector, got %d", Dim, len(d.Embedding))
		}
		result[d.Index] = d.Embedding
	}
	return result, nil
}
