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

// VoyageClient calls Voyage AI's /v1/embeddings endpoint with the voyage-3
// model (1024 dims). Tokenization isn't returned per-text; total usage comes
// back as `usage.total_tokens` so cost can be computed batch-level.
type VoyageClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewVoyageClient(apiKey string) *VoyageClient {
	return &VoyageClient{
		apiKey: apiKey,
		model:  "voyage-3",
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *VoyageClient) Name() string { return "voyage-3" }

type voyageRequest struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type,omitempty"`
}

type voyageResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *VoyageClient) Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	if c.apiKey == "" {
		return nil, errors.New("voyage: missing API key")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(voyageRequest{Input: texts, Model: c.model, InputType: inputType})
	if err != nil {
		return nil, fmt.Errorf("voyage: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.voyageai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("voyage: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("voyage: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("voyage: %s: %s", resp.Status, raw)
	}

	var out voyageResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("voyage: decode: %w", err)
	}
	if len(out.Data) != len(texts) {
		return nil, fmt.Errorf("voyage: expected %d embeddings, got %d", len(texts), len(out.Data))
	}

	// Voyage returns vectors in input order via the `index` field. Reorder
	// defensively in case the server ever shuffles.
	result := make([][]float32, len(texts))
	for _, d := range out.Data {
		if d.Index < 0 || d.Index >= len(texts) {
			return nil, fmt.Errorf("voyage: out-of-range index %d", d.Index)
		}
		if len(d.Embedding) != Dim {
			return nil, fmt.Errorf("voyage: expected %d-dim vector, got %d", Dim, len(d.Embedding))
		}
		result[d.Index] = d.Embedding
	}
	return result, nil
}
