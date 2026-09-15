package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
)

const DefaultEmbeddingModel = "text-embedding-004"

type EmbeddingProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewEmbeddingProvider(apiKey, model, baseURL string, client *http.Client) *EmbeddingProvider {
	model = strings.TrimSpace(model)
	if model == "" {
		model = DefaultEmbeddingModel
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if client == nil {
		client = &http.Client{}
	}
	return &EmbeddingProvider{apiKey: apiKey, model: model, baseURL: baseURL, client: client}
}

type embedRequest struct {
	Content embedContent `json:"content"`
}

type embedContent struct {
	Parts []part `json:"parts"`
}

type embedResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *EmbeddingProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini embeddings: missing API key")
	}
	out := make([][]float32, 0, len(texts))
	for _, text := range texts {
		vec, err := p.embedOne(ctx, text)
		if err != nil {
			return nil, err
		}
		out = append(out, vec)
	}
	return out, nil
}

func (p *EmbeddingProvider) embedOne(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(embedRequest{Content: embedContent{Parts: []part{{Text: text}}}})
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/models/%s:embedContent", p.baseURL, p.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed embedResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("gemini embeddings: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gemini embeddings http %d: %s", resp.StatusCode, string(raw))
	}
	if len(parsed.Embedding.Values) != domain.EmbeddingDimensions {
		return nil, fmt.Errorf(
			"gemini embeddings: got %d dims, want %d",
			len(parsed.Embedding.Values),
			domain.EmbeddingDimensions,
		)
	}
	vec := make([]float32, domain.EmbeddingDimensions)
	for i, v := range parsed.Embedding.Values {
		vec[i] = float32(v)
	}
	return vec, nil
}

var _ ports.EmbeddingsProvider = (*EmbeddingProvider)(nil)
