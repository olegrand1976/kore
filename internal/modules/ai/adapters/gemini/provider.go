package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/ai/app/promptguard"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
)

const DefaultModel = "gemini-3.6-flash"
const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

type ModelResolver interface {
	ResolveModel(ctx context.Context) string
}

type Config struct {
	APIKey       string
	DefaultModel string
	BaseURL      string
	Timeout      time.Duration
	Resolver     ModelResolver
	Guard        promptguard.Guard
}

type Provider struct {
	apiKey       string
	defaultModel string
	baseURL      string
	client       *http.Client
	resolver     ModelResolver
	guard        promptguard.Guard
}

func NewProvider(cfg Config) *Provider {
	model := strings.TrimSpace(cfg.DefaultModel)
	if model == "" {
		model = DefaultModel
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Provider{
		apiKey:       cfg.APIKey,
		defaultModel: model,
		baseURL:      baseURL,
		client:       &http.Client{Timeout: timeout},
		resolver:     cfg.Resolver,
		guard:        guardOrDefault(cfg.Guard),
	}
}

func guardOrDefault(g promptguard.Guard) promptguard.Guard {
	if g.MaxFieldLen == 0 && !g.BlockOnDetection {
		return promptguard.DefaultGuard()
	}
	if g.MaxFieldLen == 0 {
		g.MaxFieldLen = promptguard.DefaultGuard().MaxFieldLen
	}
	return g
}

func (p *Provider) modelFor(ctx context.Context) string {
	if p.resolver != nil {
		if model := strings.TrimSpace(p.resolver.ResolveModel(ctx)); model != "" {
			return model
		}
	}
	return p.defaultModel
}

type generateRequest struct {
	Contents []content `json:"contents"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func (p *Provider) Complete(ctx context.Context, req ports.CompletionRequest) (ports.CompletionResponse, error) {
	if p.apiKey == "" {
		return ports.CompletionResponse{}, fmt.Errorf("gemini: missing API key")
	}

	secured, err := p.guard.SecureCompletion(req)
	if err != nil {
		if errors.Is(err, promptguard.ErrInjectionDetected) {
			return ports.CompletionResponse{}, domain.ErrPromptInjectionBlocked
		}
		return ports.CompletionResponse{}, err
	}
	req = secured

	model := p.modelFor(ctx)

	contents := make([]content, 0, 2)
	if strings.TrimSpace(req.SystemPrompt) != "" {
		contents = append(contents, content{
			Role:  "user",
			Parts: []part{{Text: "Instructions système:\n" + req.SystemPrompt}},
		})
		contents = append(contents, content{
			Role:  "model",
			Parts: []part{{Text: "Compris."}},
		})
	}
	contents = append(contents, content{
		Parts: []part{{Text: req.UserPrompt}},
	})

	body, err := json.Marshal(generateRequest{Contents: contents})
	if err != nil {
		return ports.CompletionResponse{}, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", p.baseURL, model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ports.CompletionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return ports.CompletionResponse{}, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return ports.CompletionResponse{}, err
	}

	var parsed generateResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ports.CompletionResponse{}, fmt.Errorf("gemini decode: %w", err)
	}
	if parsed.Error != nil {
		return ports.CompletionResponse{}, fmt.Errorf("gemini api: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return ports.CompletionResponse{}, fmt.Errorf("gemini http %d: %s", resp.StatusCode, string(raw))
	}

	var text strings.Builder
	for _, candidate := range parsed.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				if text.Len() > 0 {
					text.WriteString("\n")
				}
				text.WriteString(part.Text)
			}
		}
	}
	out := strings.TrimSpace(text.String())
	if out == "" {
		return ports.CompletionResponse{}, fmt.Errorf("gemini: empty response")
	}
	return ports.CompletionResponse{Text: out, Model: model}, nil
}
