package app

import (
	"context"
	"strings"

	aigemini "github.com/kore/kore/internal/modules/ai/adapters/gemini"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/app/promptguard"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/platform/config"
)

type GeminiModelResolver interface {
	CurrentGeminiModel(ctx context.Context) string
}

func NewLLMProvider(cfg config.Config, resolver GeminiModelResolver) ports.LLMProvider {
	provider := strings.ToLower(strings.TrimSpace(cfg.AILLMProvider))
	guard := promptguard.Guard{BlockOnDetection: cfg.PromptGuardBlock}
	switch provider {
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return stub.NewProvider()
		}
		var modelResolver aigemini.ModelResolver
		if resolver != nil {
			modelResolver = resolverAdapter{resolver: resolver}
		}
		defaultModel := strings.TrimSpace(cfg.GeminiModel)
		if defaultModel == "" {
			defaultModel = aigemini.DefaultModel
		}
		return aigemini.NewProvider(aigemini.Config{
			APIKey:       cfg.GeminiAPIKey,
			DefaultModel: defaultModel,
			Resolver:     modelResolver,
			Guard:        guard,
		})
	case "stub", "":
		return stub.NewProvider()
	default:
		return stub.NewProvider()
	}
}

type resolverAdapter struct {
	resolver GeminiModelResolver
}

func (a resolverAdapter) ResolveModel(ctx context.Context) string {
	return a.resolver.CurrentGeminiModel(ctx)
}
