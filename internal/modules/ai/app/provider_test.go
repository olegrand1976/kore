package app

import (
	"testing"

	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/platform/config"
)

func TestNewLLMProvider_geminiWithoutKeyFallsBackToStub(t *testing.T) {
	p := NewLLMProvider(config.Config{AILLMProvider: "gemini"}, nil)
	if _, ok := p.(*stub.Provider); !ok {
		t.Fatalf("expected stub fallback, got %T", p)
	}
}

func TestNewLLMProvider_unknownUsesStub(t *testing.T) {
	p := NewLLMProvider(config.Config{AILLMProvider: "unknown"}, nil)
	if _, ok := p.(*stub.Provider); !ok {
		t.Fatalf("expected stub fallback, got %T", p)
	}
}
