package app

import (
	"context"
	"testing"

	"github.com/kore/kore/internal/modules/org/domain"
)

func TestNormalizeGeminiModel_valid(t *testing.T) {
	model, err := normalizeGeminiModel("gemini-3.5-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "gemini-3.5-flash" {
		t.Fatalf("got %q", model)
	}
}

func TestNormalizeGeminiModel_invalid(t *testing.T) {
	_, err := normalizeGeminiModel("../evil")
	if err == nil || err != domain.ErrInvalidGeminiModel {
		t.Fatalf("expected ErrInvalidGeminiModel, got %v", err)
	}
}

func TestPlatformService_CurrentGeminiModel_fallback(t *testing.T) {
	s := &platformService{defaultGeminiModel: "gemini-3.5-flash"}
	if got := s.CurrentGeminiModel(context.Background()); got != "gemini-3.5-flash" {
		t.Fatalf("got %q", got)
	}
}
